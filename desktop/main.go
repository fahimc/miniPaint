// Studio is a portable local editor; optional FileBrowser and OIDC integrations
// do not replace or replicate the upstream file store.
package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

//go:embed web/*
var assets embed.FS
var version = "studio-preview"

type config struct {
	Listen      string `json:"listen"`
	FileBrowser string `json:"filebrowser"`
	Issuer      string `json:"issuer"`
	ClientID    string `json:"clientId"`
	RedirectURL string `json:"redirectUrl"`
}
type loginAttempt struct {
	Verifier, Nonce string
	Expires         time.Time
}
type session struct {
	Name, Subject string
	Expires       time.Time
}

func random() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
func main() {
	exe, _ := os.Executable()
	if logFile, e := os.OpenFile(filepath.Join(filepath.Dir(exe), "Studio.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); e == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
	}
	cfg := config{Listen: "localhost:8870"}
	if body, err := os.ReadFile(filepath.Join(filepath.Dir(exe), "Studio.json")); err == nil {
		if err = json.Unmarshal([]byte(strings.TrimPrefix(string(body), "\ufeff")), &cfg); err != nil {
			log.Fatal("Invalid Studio.json: ", err)
		}
	}
	listen := flag.String("listen", cfg.Listen, "Loopback listen address")
	server := flag.String("filebrowser", cfg.FileBrowser, "Optional FileBrowser URL")
	noBrowser := flag.Bool("no-browser", false, "Do not open the browser")
	flag.Parse()
	host, port, err := net.SplitHostPort(*listen)
	if err != nil || (host != "localhost" && (net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback())) {
		log.Fatal("Use a loopback listen address")
	}
	origin := "http://" + *listen
	var proxy *httputil.ReverseProxy
	if *server != "" {
		target, e := url.Parse(*server)
		if e != nil || target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") || target.User != nil || target.RawQuery != "" || target.Fragment != "" {
			log.Fatal("Invalid FileBrowser URL")
		}
		proxy = httputil.NewSingleHostReverseProxy(target)
		director := proxy.Director
		proxy.Director = func(r *http.Request) {
			director(r)
			r.Host = target.Host
			for _, h := range []string{"Forwarded", "X-Forwarded-Host", "X-Forwarded-Proto", "X-Forwarded-Port"} {
				r.Header.Del(h)
			}
			if r.Header.Get("Origin") != "" {
				r.Header.Set("Origin", target.Scheme+"://"+target.Host)
			}
		}
		proxy.ModifyResponse = func(res *http.Response) error {
			cookies := res.Cookies()
			res.Header.Del("Set-Cookie")
			for _, c := range cookies {
				c.Domain = ""
				c.Path = "/"
				c.Secure = false
				res.Header.Add("Set-Cookie", c.String())
			}
			return nil
		}
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			http.Error(w, "FileBrowser is unavailable; local editing remains available", 502)
		}
	}
	var provider *oidc.Provider
	var oauth oauth2.Config
	var verifier *oidc.IDTokenVerifier
	if cfg.Issuer != "" || cfg.ClientID != "" {
		if cfg.Issuer == "" || cfg.ClientID == "" || cfg.RedirectURL == "" || os.Getenv("STUDIO_OIDC_CLIENT_SECRET") == "" {
			log.Fatal("MyAccounts requires issuer, clientId, exact redirectUrl and STUDIO_OIDC_CLIENT_SECRET")
		}
		if cfg.RedirectURL != origin+"/auth/callback" {
			log.Fatal("MyAccounts callback must match the local listen origin exactly")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		provider, err = oidc.NewProvider(ctx, cfg.Issuer)
		cancel()
		if err != nil {
			log.Fatal("MyAccounts discovery failed: ", err)
		}
		oauth = oauth2.Config{ClientID: cfg.ClientID, ClientSecret: os.Getenv("STUDIO_OIDC_CLIENT_SECRET"), RedirectURL: cfg.RedirectURL, Endpoint: provider.Endpoint(), Scopes: []string{oidc.ScopeOpenID, "profile", "email"}}
		verifier = provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
	}
	var mu sync.Mutex
	attempts := map[string]loginAttempt{}
	sessions := map[string]session{}
	cookie := func(w http.ResponseWriter, name, value string, age int) {
		http.SetCookie(w, &http.Cookie{Name: name, Value: value, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: age})
	}
	current := func(r *http.Request) (session, bool) {
		c, e := r.Cookie("studio_session")
		if e != nil {
			return session{}, false
		}
		mu.Lock()
		defer mu.Unlock()
		s, ok := sessions[c.Value]
		return s, ok && s.Expires.After(time.Now())
	}
	root, _ := fs.Sub(assets, "web")
	static := http.FileServer(http.FS(root))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /studio-config.json", func(w http.ResponseWriter, r *http.Request) {
		s, _ := current(r)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(map[string]any{"version": version, "filebrowserURL": *server, "myaccounts": provider != nil, "account": s.Name})
	})
	mux.HandleFunc("GET /auth/login", func(w http.ResponseWriter, r *http.Request) {
		if provider == nil {
			http.Error(w, "Standalone MyAccounts is not configured; use FileBrowser sign-in", 503)
			return
		}
		state := random()
		nonce := random()
		pkce := oauth2.GenerateVerifier()
		mu.Lock()
		for k, v := range attempts {
			if v.Expires.Before(time.Now()) {
				delete(attempts, k)
			}
		}
		if len(attempts) > 100 {
			mu.Unlock()
			http.Error(w, "Too many sign-in attempts", 429)
			return
		}
		attempts[state] = loginAttempt{pkce, nonce, time.Now().Add(5 * time.Minute)}
		mu.Unlock()
		cookie(w, "studio_login", state, 300)
		http.Redirect(w, r, oauth.AuthCodeURL(state, oidc.Nonce(nonce), oauth2.S256ChallengeOption(pkce)), 302)
	})
	mux.HandleFunc("GET /auth/callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		c, e := r.Cookie("studio_login")
		mu.Lock()
		a, ok := attempts[state]
		delete(attempts, state)
		mu.Unlock()
		cookie(w, "studio_login", "", -1)
		if provider == nil || e != nil || c.Value != state || !ok || time.Now().After(a.Expires) {
			http.Error(w, "Sign-in expired; start again", 400)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		token, e := oauth.Exchange(ctx, r.URL.Query().Get("code"), oauth2.VerifierOption(a.Verifier))
		if e != nil {
			http.Error(w, "MyAccounts token exchange failed", 400)
			return
		}
		raw, ok := token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "Missing identity token", 400)
			return
		}
		identity, e := verifier.Verify(ctx, raw)
		if e != nil || identity.Nonce != a.Nonce {
			http.Error(w, "MyAccounts identity verification failed", 400)
			return
		}
		var claims struct {
			Name     string `json:"name"`
			Username string `json:"preferred_username"`
		}
		if identity.Claims(&claims) != nil {
			http.Error(w, "Invalid identity claims", 400)
			return
		}
		name := claims.Username
		if name == "" {
			name = claims.Name
		}
		if name == "" {
			name = identity.Subject
		}
		id := random()
		mu.Lock()
		for k, s := range sessions {
			if s.Expires.Before(time.Now()) {
				delete(sessions, k)
			}
		}
		sessions[id] = session{name, identity.Subject, time.Now().Add(8 * time.Hour)}
		mu.Unlock()
		cookie(w, "studio_session", id, 28800)
		http.Redirect(w, r, "/", 302)
	})
	mux.HandleFunc("POST /auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if c, e := r.Cookie("studio_session"); e == nil {
			mu.Lock()
			delete(sessions, c.Value)
			mu.Unlock()
		}
		cookie(w, "studio_session", "", -1)
		w.WriteHeader(204)
	})
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		if provider != nil {
			if _, ok := current(r); !ok {
				http.Error(w, "Sign in with MyAccounts first", 401)
				return
			}
		}
		if proxy == nil {
			http.Error(w, "FileBrowser is not configured", 503)
			return
		}
		proxy.ServeHTTP(w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		if provider != nil && (r.URL.Path == "/" || r.URL.Path == "/index.html") {
			if _, ok := current(r); !ok {
				http.Redirect(w, r, "/auth/login", 302)
				return
			}
		}
		if strings.HasSuffix(r.URL.Path, ".html") {
			w.Header().Set("Cache-Control", "no-store")
		}
		static.ServeHTTP(w, r)
	})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != *listen && r.Host != "localhost:"+port {
			http.Error(w, "Invalid host", 400)
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" {
			if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host {
				http.Error(w, "Invalid origin", 403)
				return
			}
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		mux.ServeHTTP(w, r)
	})
	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Studio %s at %s", version, origin)
	fmt.Println(origin)
	if !*noBrowser {
		openBrowser(origin)
	}
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 15 * time.Second, IdleTimeout: 90 * time.Second}
	log.Fatal(srv.Serve(listener))
}
