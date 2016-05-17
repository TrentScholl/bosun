$env:GOPATH = (Get-Item $PSScriptRoot).Parent.Parent.FullName

$env:GOOS = "windows"
$env:GOARCH = "amd64"

$OperatingSystems = @("windows","linux")
$Architectures = @("amd64", "386")

$VersionDate = Get-Date -Format yyyyMMddHHmmss
$GitSHA = (git rev-parse HEAD) | Out-String


Function Build
{
    Param([string]$os, [string]$arch)

    $env:GOOS = $os
    $env:GOARCH = $arch

    $ext = ""

    If ($os -eq "windows")
    {
        $ext = ".exe"
    }

    Write-Host "Building scollector for $os ($arch)"

    go build -o "bin\scollector-$os-$arch$ext" -ldflags "-X bosun.org/_version.VersionSHA=$GitSHA -X bosun.org/_version.OfficialBuild=true -X bosun.org/_version.VersionDate=$VersionDate" .\cmd\scollector 
}

Write-Host "Running collector tests..."

& go test -v .\cmd\scollector\collectors\

If ($LASTEXITCODE -ne 0)
{
    Exit $LASTEXITCODE
}

ForEach ($OperatingSystem in $OperatingSystems) {
    ForEach ($Architecture in $Architectures) {
        Build -os $OperatingSystem -arch $Architecture
    }   
}