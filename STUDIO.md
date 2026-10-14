# Studio

A Photoshop-style desktop workspace built on miniPaint, with its existing drawing tools, layers, filters, selection, text and undo engine. Studio is independently branded; it does not contain Adobe code/assets or implement all Photoshop features. No PSD engine, Smart Objects, Camera Raw or Adobe cloud services are included.

## Portable Windows app

Double-click **Studio.exe**. It opens your browser at `http://localhost:8870`. Local image editing needs no FileBrowser, Docker, Node, Office installation or account. **Open** selects a device file or FileBrowser file. **Save / Ctrl+S** writes back to the opened file when writable; **Save as / Ctrl+Shift+S** selects a new destination. Local direct save uses the browser File System Access API when available; other browsers download a copy.

An optional `Studio.json` next to the executable configures a FileBrowser connection:

```json
{"listen":"localhost:8870","filebrowser":"http://localhost:8810"}
```

CLI `--filebrowser URL`, `--listen localhost:8871` and `--no-browser` override those defaults. Without a configured/available FileBrowser, local editing stays available. **FileBrowser** opens a permission-aware source/folder picker. The editor uses ordinary files through FileBrowser, without a new media database.

The server accepts loopback clients and checks Host and mutation Origin. It proxies only `/api/` to the configured FileBrowser. Studio.log records startup errors. Browser cookies remain session credentials; passwords/tokens are not put in editor URLs or browser localStorage. Studio itself does not start FileBrowser or modify its account permissions.

## Integrated FileBrowser

Install the updated FileBrowser build with the Studio assets and API. Choose **Image Studio** in the sidebar, or **Edit in Studio** from a file's context menu/preview overflow. `/studio?source=My%20Drive&path=/image.png` opens an existing image. The editor iframe and APIs share FileBrowser's origin and authenticated session; MyAccounts works through the existing FileBrowser OIDC login. Public share editing is deliberately not exposed by this integration.

Supported opens: browser-decodable PNG, JPEG, WebP, GIF, BMP, SVG, and miniPaint layered JSON. Saves: PNG, JPEG, WebP, layered JSON. Opening animated GIFs as still images does not preserve animation on export; use the native miniPaint export tools for its supported animation workflow. Raster image saves flatten the visible layers, and JPEG may lose transparency/metadata; use layered JSON to keep editable layers. PSD is not supported. FileBrowser image/project open/save limit: 64 MiB. Filename extensions must match the encoded format. New Save as refuses existing filenames.

The `/api/studio/file` endpoint requires the existing view permission; writes additionally require modify or create and honour read-only sources, scopes, access rules, quotas and the normal atomic FileBrowser write pipeline. Existing-file saves require a SHA-256 `If-Match` ETag; changed originals return 412 and the editor keeps its draft. New files require `If-None-Match: *`. Studio saves are serialized within the process. External filesystem writes and other FileBrowser endpoints do not acquire the Studio lock, so this is an optimistic conflict check, not a filesystem-wide transaction or real-time collaborative editor. Refresh/reopen or save a copy when a conflict is reported. No auto-save overwrites are performed.

The source files for the core adapter are provided under `integrations/filebrowser/`, with an idempotent installer for the current Drive fork. Existing Office/Photos work must be preserved. Build the full FileBrowser frontend and backend after installing; copy this app's generated `desktop/web` into the core `backend/internal/web/embed/studio`. Keep the Office engine assets while updating embeds. See the FileBrowser fork's OFFICE.md for its full portable build process.

## MyAccounts login

**Through FileBrowser:** use the account button → **Sign in to FileBrowser / MyAccounts**, finish sign-in, then **I'm signed in**. The standalone companion must use the same hostname as FileBrowser's callback, for example both `localhost` (different ports work). `127.0.0.1` and `localhost` do not share host-only cookies. For remote FileBrowser, use integrated Studio on that remote origin to share its authenticated session. The local companion does not copy remote browser sessions.

**Independent standalone login:** register a separate confidential OIDC client in your MyAccounts Applications console, using the exact callback `http://localhost:8870/auth/callback`, authorization code flow and standard `openid profile email` scopes. Supply:

```json
{
  "listen":"localhost:8870",
  "issuer":"https://login.m8e.co.uk",
  "clientId":"YOUR_REGISTERED_STUDIO_CLIENT_ID",
  "redirectUrl":"http://localhost:8870/auth/callback"
}
```

Supply the secret through `STUDIO_OIDC_CLIENT_SECRET` or `scripts/run-with-myaccounts.ps1` with a CurrentUser DPAPI-protected secret file. Never commit the secret. When configured, the app gates its entry page and proxy API behind MyAccounts. It uses provider discovery, maintained go-oidc verification, code + PKCE S256, nonce/state binding, signature/issuer/audience/expiry validation, HttpOnly SameSite cookies and server-held expiring sessions. Reopening after restart requires a fresh app session. Local sign-out is `POST /auth/logout`; it does not sign out every MyAccounts relying party. The issuer account does not automatically grant a FileBrowser session or folder access: FileBrowser independently enforces its own sign-in and permissions. Existing MyAccounts registrations are not changed by the build.

## Build

Requirements: Windows, Node/npm compatible with package.json and Go 1.27+.

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build-windows.ps1
```

Output is `build/windows-amd64`. The script stages source on a local drive, builds miniPaint's production bundle and embeds it in a Go executable. The generated website can also be statically hosted for local-only editing. No tests are run by this build script. Retain upstream MIT notices; original miniPaint attribution remains in About and MIT-LICENSE.txt.
