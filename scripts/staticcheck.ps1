param(
  [ValidateSet("0", "1")]
  [string]$CgoEnabled = "0",
  [string]$StaticcheckVersion = "v0.7.0",
  [string[]]$Packages = @("./...")
)

$ErrorActionPreference = "Stop"

$env:GOOS = "windows"
$env:CGO_ENABLED = $CgoEnabled

$staticcheck = Get-Command staticcheck -ErrorAction SilentlyContinue
if ($staticcheck) {
  & $staticcheck.Source @Packages
  exit $LASTEXITCODE
}

& go run "honnef.co/go/tools/cmd/staticcheck@$StaticcheckVersion" @Packages
exit $LASTEXITCODE
