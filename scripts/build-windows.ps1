param(
 [string]$BuildRoot=(Join-Path $env:LOCALAPPDATA 'StudioBuild'),
 [string]$OutputDirectory='',
 [string]$GoExecutable='go',
 [string]$FileBrowserUrl='http://localhost:8810'
)
$ErrorActionPreference='Stop'
$repoRoot=Split-Path $PSScriptRoot -Parent
if(!$OutputDirectory){$OutputDirectory=Join-Path $repoRoot 'build\windows-amd64'}
New-Item -ItemType Directory -Path $BuildRoot,$OutputDirectory -Force|Out-Null
robocopy $repoRoot $BuildRoot /E /XD .git node_modules build desktop agents /NFL /NDL /NJH /NJS /NP
if($LASTEXITCODE -ge 8){throw 'Could not stage source'}
Push-Location $BuildRoot
try{& npm.cmd ci;if($LASTEXITCODE){throw 'npm ci failed'};& npm.cmd run build;if($LASTEXITCODE){throw 'Frontend build failed'}}finally{Pop-Location}
$desktop=Join-Path $BuildRoot 'desktop'
New-Item -ItemType Directory -Path $desktop,(Join-Path $desktop 'web') -Force|Out-Null
Copy-Item (Join-Path $repoRoot 'desktop\*.go'),(Join-Path $repoRoot 'desktop\go.mod'),(Join-Path $repoRoot 'desktop\go.sum') $desktop -Force
$web=Join-Path $desktop 'web'
Copy-Item (Join-Path $BuildRoot 'index.html') $web -Force
foreach($directory in @('dist','images','src')){Copy-Item (Join-Path $BuildRoot $directory) $web -Recurse -Force}
Copy-Item (Join-Path $repoRoot 'MIT-LICENSE.txt') $web -Force
$previousCgo=$env:CGO_ENABLED
try{$env:CGO_ENABLED='0';Push-Location $desktop;try{& $GoExecutable build -trimpath -ldflags '-s -w -H windowsgui' -o (Join-Path $OutputDirectory 'Studio.exe') .;if($LASTEXITCODE){throw 'Go build failed'}}finally{Pop-Location}}finally{$env:CGO_ENABLED=$previousCgo}
[IO.File]::WriteAllText((Join-Path $OutputDirectory 'Studio.json'),(@{listen='localhost:8870';filebrowser=$FileBrowserUrl}|ConvertTo-Json))
Copy-Item (Join-Path $repoRoot 'STUDIO.md'),(Join-Path $repoRoot 'MIT-LICENSE.txt') $OutputDirectory -Force
Write-Host "Built portable Studio at $OutputDirectory"
