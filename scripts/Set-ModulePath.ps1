param(
    [Parameter(Mandatory = $true)]
    [string]$ModulePath
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$currentModulePath = "github.com/AzureIvory/winui"
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)

if ([string]::IsNullOrWhiteSpace($ModulePath)) {
    throw "ModulePath must not be empty."
}

$files = Get-ChildItem -Path $root -Recurse -File -Include *.go,go.mod,README.md,DEVELOPING.md

foreach ($file in $files) {
    $content = [System.IO.File]::ReadAllText($file.FullName, $utf8NoBom)
    $updated = $content.Replace($currentModulePath, $ModulePath)
    if ($updated -ne $content) {
        [System.IO.File]::WriteAllText($file.FullName, $updated, $utf8NoBom)
    }
}

Write-Host "Updated module path to $ModulePath"
