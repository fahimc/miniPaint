param([Parameter(Mandatory=$true)][string]$Exe,[Parameter(Mandatory=$true)][string]$ProtectedSecret)
$ErrorActionPreference='Stop'
# ProtectedSecret contains a Windows CurrentUser DPAPI-protected UTF-8 client
# secret. Do not pass plaintext client secrets in arguments or Studio.json.
Add-Type -AssemblyName System.Security
$encrypted=[IO.File]::ReadAllBytes($ProtectedSecret)
$plain=[Security.Cryptography.ProtectedData]::Unprotect($encrypted,$null,[Security.Cryptography.DataProtectionScope]::CurrentUser)
try{$env:STUDIO_OIDC_CLIENT_SECRET=[Text.Encoding]::UTF8.GetString($plain);& $Exe}finally{[Array]::Clear($plain,0,$plain.Length);Remove-Item Env:\STUDIO_OIDC_CLIENT_SECRET -ErrorAction SilentlyContinue}
