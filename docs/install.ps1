$ErrorActionPreference = 'Stop'
$repo = "smilemilks2021/easy-web"
$installDir = "$env:LOCALAPPDATA\easy-web"

$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name.TrimStart('v')

Write-Host "Installing easy-web v$version (windows/amd64)..."
$url = "https://github.com/$repo/releases/download/v$version/easy-web_${version}_windows_amd64.zip"
$tmpZip = "$env:TEMP\easy-web.zip"

Invoke-WebRequest $url -OutFile $tmpZip
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Expand-Archive -Path $tmpZip -DestinationPath $installDir -Force
Remove-Item $tmpZip

# Add to user PATH if not already present
$currentPath = [Environment]::GetEnvironmentVariable('PATH', 'User')
if ($currentPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable('PATH', "$currentPath;$installDir", 'User')
    Write-Host "Added $installDir to PATH (restart terminal to apply)"
}
Write-Host "easy-web $version installed. Run: easy-web --help"
