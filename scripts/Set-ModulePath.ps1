param(
    [Parameter(Mandatory = $true)]
    [string]$ModulePath
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$oldModulePath = "github.com/yourname/winui"

if ([string]::IsNullOrWhiteSpace($ModulePath)) {
    throw "ModulePath must not be empty."
}

$files = Get-ChildItem -Path $root -Recurse -File -Include *.go,go.mod,README.md,DEVELOPING.md

foreach ($file in $files) {
    $content = Get-Content -Path $file.FullName -Raw
    $updated = $content.Replace($oldModulePath, $ModulePath)
    if ($updated -ne $content) {
        Set-Content -Path $file.FullName -Value $updated -NoNewline
    }
}

Write-Host "Updated module path to $ModulePath"
