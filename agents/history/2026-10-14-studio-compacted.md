# Studio fork and FileBrowser integration

## Scope and source

User requested a new GitHub fork of viliusle/miniPaint, Photoshop UI recreation, ordinary-file open/save integration with their FileBrowser, standalone operation with optional integration, and MyAccounts login. New public fork fahimc/miniPaint, local I:/Projects/minipaint-studio. Upstream baseline a79733e (miniPaint 4.14.3), master. Repository-history skill used before source changes; no upstream AGENTS/history found. New fork source publication is part of this request; existing FileBrowser changes are left uncommitted and unreleased.

## Implementation

- Independently branded Studio dark workspace retains miniPaint's real engine/tools/layers/filters. Compact File/Edit/Image/Layer/Select/Filter/View/Window/Help menus, tool rail, options, document tab, Navigator/Color/Properties/Layers panels with hide/reset controls. Not Adobe software or a claim of pixel-identical/full Photoshop/PSD functionality.
- src/js/studio/workspace.js routes standard Ctrl+O/S/Shift+S/N, device File System Access open/save or download fallback, server source/folder picker, remote originals, editable layered JSON and flattened PNG/JPEG/WebP. View-only files allow local edits/copies. Unsaved warning, visible failures, 412 conflict preservation, revision checks and captured save destinations prevent async save/open switches from writing to another document or clearing newer edits. Existing miniPaint file/export actions retained.
- desktop/ is a separate Go module, maintained go-oidc v3.21.0 and oauth2 v0.37.0. Embeds generated website, defaults localhost:8870, optional adjacent Studio.json/CLI FileBrowser URL. Local editing works with no connected server. Proxy only /api/, fixed configured upstream, loopback listener, Host and mutation Origin checks, HttpOnly upstream cookies retained, no secrets in browser URLs/localStorage.
- Optional independent MyAccounts OIDC: discovery, confidential code + S256 PKCE, state cookie + nonce binding, issuer/audience/signature/expiry checks, server-held expiring sessions. Secret supplied by environment or CurrentUser DPAPI launcher, never JSON. No FileBrowser permissions/session are automatically granted by this independent identity. Standalone shared FileBrowser cookies require same hostname; remote users use integrated Studio.
- integrations/filebrowser contains studio.go, Studio.vue, idempotent source installer. Actual adapter installed in I:/Projects/filebrowser-drive: authenticated /studio route + sidebar, Edit in Studio in file context/preview overflow, static /studio/editor/ embed, /api/studio/file GET/PUT. Uses existing permissions/scopes/access/private path checks, read-only sources, quotas, atomic resource handlers. 64 MiB limit, SHA256 ETags, If-Match existing save, If-None-Match:* new save, no replacing existing names in Save as. Per-process hashed write locks serialize Studio clients; external/filesystem/non-Studio writes are still optimistic, not a universal transaction or collaboration protocol. Public shares unsupported. No new media DB.
- scripts/build-windows.ps1 stages source locally, npm ci/build + Go build; default output build/windows-amd64. scripts/run-with-myaccounts.ps1 unprotects CurrentUser DPAPI secret. Full usage/limits in STUDIO.md. Built dist bundle committed as upstream does, so static hosting works after clone. Generated desktop/web ignored except sentinel; upstream MIT notices retained.

## Builds and evidence

Production webpack builds, complete FileBrowser Vite builds and native CGO_ENABLED=0 Go builds succeeded with Go 1.27.2. Reusable PowerShell build script executed successfully; script syntax and Git whitespace checks passed. No unit/integration/browser functional tests added/run (not requested). Desktop screenshots visually reviewed: actual sample JPEG fetched from FileBrowser displayed in Studio canvas and Navigator; standalone MyAccounts link reached the real public sign-in screen. Save round trip, OIDC callback/login and full permission regression are not verified by those screenshots. Inherited npm ci reports 24 advisories (3 critical); no broad dependency migration performed. Webpack bundle-size and FileBrowser existing dynamic-import warnings retained.

Introduced compile issue ResourcePostHandler capitalization corrected before successful core builds. Preview exe replacement briefly raced process termination; wait for process exit before copying. Keep generated asset copy/build ordering sequential; backend embeds must receive the latest Studio + FileBrowser frontend bundles before Go build.

## Windows paths and runtime

Build stages: C:/Projects/.cache/minipaint-studio-build and filebrowser-office-build. Outputs under I:/Projects/tooling/releases/studio-preview; portable Studio package under package/. Studio.exe roughly 14 MiB; full FileBrowser roughly 543 MiB because it retains Office assets. No GitHub release requested/created.

Loopback standalone visual demo localhost:8870 configured against localhost:8830; source files are copies of Windows sample wallpapers in filebrowser-office-preview/photos-files. 8830 was upgraded to the new core; its isolated noauth demo remains loopback only and must never be tunnelled. Production 8810/8794 and original Photos binaries remain unchanged.

Independent standalone MyAccounts runtime F:/MyAccounts/studio-preview, localhost:8871, protected dedicated client-secret.dpapi, Studio.json with exact registered callback, Start-Studio.ps1. New distinct Studio OIDC application registered at existing issuer; secret never committed. Launcher holds secret in process environment; protected runtime ACL restricted to current user and SYSTEM. Client registration did not modify MyAccounts source or existing relying parties. Read current runtime config for client identity, do not copy secrets/history artifacts into Git.

Existing authenticated public integration runtime F:/MyAccounts/integration-preview can be upgraded using its Start-FileBrowser.ps1 after a source-build-complete binary is available. Preserve protected client, settings DB, configs, original Photos/gateway/tunnel and their session origin. Runtime README/public-origin.txt holds temporary hostname; never commit it. Do not tunnel the 8830 noauth demo. Reconcile owned listeners before stopping. See FileBrowser history for final installed process/binary update.

## Resume constraints

Preserve FileBrowser's uncommitted Photos work and unrelated drive-preview.png/frontend/drive-preview.html. Its compact MyAccounts/Photos histories describe existing protected previews. Do not claim save/login callback acceptance or exact Photoshop/PSD parity without requested validation. Future work: requested functional/permission/save conflict tests, dependency remediation, independent SSO callback verification, richer Photoshop tool/panel behaviour, PSD support if explicitly scoped. No new Android work.

## Final activation

Implementation f60ed93 pushed to the new fork master and remote SHA confirmed. Portable package Studio.exe is 14,724,096 bytes, SHA256 8696fc7e934cfc6db616843c663a7757b9171826ede4361968ab5a693fbd2d09; ZIP beside package directory. Final core filebrowser-studio.exe is 569,330,688 bytes, SHA256 cddf96f91b616ba3c50e749122c239872ace82c01261afc1ad6db8ce0b774f04.

Authenticated public integration core8850 upgraded through its original DPAPI launcher, changing only its executable path; previous wrapper retained as Start-FileBrowser.ps1.before-studio. Current core PID88836, launcher90980; existing Photos8851/gateway8852/tunnel unchanged. Updated processes.json/README in protected runtime. Public /studio and /studio/editor/ respond200 (SPA/static availability only, not authenticated callback proof); exact temporary link stays outside Git. Integrated screenshot8830 shows actual sample JPEG in canvas plus Image Studio sidebar; no public credentials entered.

Current isolated8830 core PID66412 uses same new core; standalone8870 PID91448 runs package/Studio.exe with --filebrowser http://localhost:8830 override. Independent MyAccounts8871 child91692/launcher91524 runs exact package EXE from protected runtime. Verify listeners/paths before stopping. Package Studio.json defaults localhost:8810, whose original published core is unchanged and needs the new adapter/core to provide Studio file endpoints. Final user links should distinguish upgraded integration preview from older8810 installation. No GitHub release or FileBrowser publication performed.

## Narrow header styling correction

User reported St badge overlapping File beneath the Studio wordmark at narrow widths. The <=720px media rule now hides the redundant menu pseudo-element and sets 8px menu padding; desktop badge remains. Rebuilt webpack dist and portable EXE. Visually reviewed actual running page at 600px and 1440px: no badge/menu overlap. No functional suites added/run. Changes remain local, uncommitted/unpushed in this follow-up.

Current8870 process95832 runs tooling/releases/studio-preview/Studio-style-fix.exe with localhost:8830 override. Package/Studio.exe and companion ZIP updated; new EXE SHA256 3630e314e15b9369b450cf4aa6643a8dec02b13e971b67a87687be0e305dbe88 supersedes the earlier standalone package hash only. Existing authenticated8871 and embedded FileBrowser cores were not restarted/rebuilt for this small standalone preview fix; their next build must include current Studio assets. Screenshot artifacts remain outside Git under filebrowser-office-preview/studio-header-fix-*.png. Restart briefly attempted before the new EXE was ready, then recovered after successful Go build; preserve build-before-restart ordering.

## Locally built Windows release preparation

User explicitly authorized pushing the styling correction and releasing the EXE, with no GitHub Actions. No workflows are registered on the fork. Planned tag studio-v0.1.0, preview release because save/SSO callbacks and full functional acceptance remain unverified. Add a build-script Version parameter so the locally compiled EXE reports the release tag. Release source includes the narrow-menu correction; package only EXE/config/docs/license files, never runtime Studio.log or protected MyAccounts configuration. Build/package/upload locally through gh; preserve all running previews. Final publication evidence follows after successful upload.
