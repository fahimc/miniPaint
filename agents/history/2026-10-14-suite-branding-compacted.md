# Suite branding

User supplied photo_editor.svg; images/logo.svg and favicon.svg preserve its exact
artwork. Updated transparent PNG fallback, submenu image, desktop menu badge, browser
favicon and Windows amd64 native icon. Narrow menu hides redundant desktop badge as
before; actual600px and1440px screenshots show no overlap. Source attribution stays.

build-windows.ps1 carries generated .syso into staging. FileBrowser installer adapter
copies its sidebar icon to public assets. Production webpack and native Go build passed;
committed-style dist/bundle.js regenerated. New local EXE is in sibling
one-workspace/dist/branding/Studio.exe; original released EXE/runtime unchanged.
Actual screenshot preview used an unavailable dummy FileBrowser upstream to ensure
no live server/data touched; owned preview process stopped. No save/login regression
claim; artwork-only follow-up, no new functional tests. No commit/push/release.
Source logo and PNG/ICO synchronization lives in sibling one-workspace/scripts.

## Source publication follow-up

User requested pushing the affected repositories and adding logos to their main
READMEs. This supersedes the earlier local-only publication scope. All five main
READMEs now reference their local SVG assets. Drive publication includes the
previous Photos/Studio prerequisite source so the branded app imports build from
the pushed branch; unrelated preview files and concurrent Workspace Mail changes
remain excluded. Existing executable releases and live installations are unchanged.
GitHub Actions is disabled before publication; builds remain local.

Source and main README branding pushed and remote SHA verified on master:
https://github.com/fahimc/miniPaint/commit/c7ac6581987f4d7d739449adfa81f6ff90d2c306
README local image targets verified in the staged Git tree before commit.
Original EXE release assets unchanged; no Actions build was used.
