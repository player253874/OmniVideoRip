$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Get-GitHubHeaders {
    $headers = @{
        Accept                 = "application/vnd.github+json"
        "X-GitHub-Api-Version" = "2022-11-28"
        "User-Agent"          = "OmniVideoRip-Build"
    }
    if ($env:GITHUB_TOKEN) {
        $headers.Authorization = "Bearer $($env:GITHUB_TOKEN)"
    }
    return $headers
}

function Get-LatestReleaseAsset {
    param(
        [Parameter(Mandatory = $true)][string]$Repository,
        [Parameter(Mandatory = $true)][string]$AssetName
    )

    Write-Host "Resolving latest $Repository release..."
    $release = Invoke-RestMethod `
        -Uri "https://api.github.com/repos/$Repository/releases/latest" `
        -Headers (Get-GitHubHeaders)
    $asset = $release.assets | Where-Object { $_.name -eq $AssetName } | Select-Object -First 1
    if (-not $asset) {
        throw "Release asset '$AssetName' was not found in $Repository release '$($release.tag_name)'."
    }
    if ($asset.digest -notmatch '^sha256:([0-9a-fA-F]{64})$') {
        throw "GitHub did not provide a valid SHA-256 digest for $Repository/$AssetName."
    }

    return [pscustomobject]@{
        Repository = $Repository
        Name       = $AssetName
        Tag        = $release.tag_name
        Url        = $asset.browser_download_url
        Sha256     = $Matches[1].ToLowerInvariant()
    }
}

function Save-VerifiedDownload {
    param(
        [Parameter(Mandatory = $true)][string]$Url,
        [Parameter(Mandatory = $true)][string]$ExpectedSha256,
        [Parameter(Mandatory = $true)][string]$Destination
    )

    $parent = Split-Path -Parent $Destination
    New-Item -ItemType Directory -Path $parent -Force | Out-Null
    $temporary = Join-Path $parent (".{0}.{1}.download" -f (Split-Path -Leaf $Destination), [guid]::NewGuid())
    try {
        Invoke-WebRequest -UseBasicParsing -Uri $Url -OutFile $temporary
        $actualSha256 = (Get-FileHash -Algorithm SHA256 -LiteralPath $temporary).Hash.ToLowerInvariant()
        if ($actualSha256 -ne $ExpectedSha256) {
            throw "SHA-256 verification failed for $Url. Expected $ExpectedSha256, got $actualSha256."
        }
        Move-Item -LiteralPath $temporary -Destination $Destination
    } finally {
        if (Test-Path -LiteralPath $temporary) {
            Remove-Item -LiteralPath $temporary -Force
        }
    }
}

function Assert-WindowsExecutable {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$DisplayName
    )

    $stream = [System.IO.File]::OpenRead($Path)
    try {
        if ($stream.Length -lt 2 -or $stream.ReadByte() -ne 0x4d -or $stream.ReadByte() -ne 0x5a) {
            throw "$DisplayName is not a valid Windows executable: $Path"
        }
    } finally {
        $stream.Dispose()
    }
}

function Ensure-YtDlp {
    param([Parameter(Mandatory = $true)][string]$BinDir)

    $destination = Join-Path $BinDir "yt-dlp.exe"
    if (Test-Path -LiteralPath $destination -PathType Leaf) {
        Assert-WindowsExecutable -Path $destination -DisplayName "yt-dlp"
        Write-Host "Using existing bin/yt-dlp.exe"
        return
    }

    $asset = Get-LatestReleaseAsset -Repository "yt-dlp/yt-dlp" -AssetName "yt-dlp.exe"
    Write-Host "Downloading yt-dlp $($asset.Tag)..."
    Save-VerifiedDownload -Url $asset.Url -ExpectedSha256 $asset.Sha256 -Destination $destination
    Assert-WindowsExecutable -Path $destination -DisplayName "yt-dlp"
    Write-Host "Verified yt-dlp SHA-256: $($asset.Sha256)"
}

function Ensure-FFmpeg {
    param([Parameter(Mandatory = $true)][string]$BinDir)

    $destination = Join-Path $BinDir "ffmpeg.exe"
    if (Test-Path -LiteralPath $destination -PathType Leaf) {
        Assert-WindowsExecutable -Path $destination -DisplayName "ffmpeg"
        Write-Host "Using existing bin/ffmpeg.exe"
        return
    }

    $assetName = "ffmpeg-master-latest-win64-lgpl.zip"
    $asset = Get-LatestReleaseAsset -Repository "BtbN/FFmpeg-Builds" -AssetName $assetName
    $temporaryRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("omnivideorip-build-{0}" -f [guid]::NewGuid())
    New-Item -ItemType Directory -Path $temporaryRoot | Out-Null
    try {
        $archive = Join-Path $temporaryRoot $assetName
        Write-Host "Downloading FFmpeg $($asset.Tag)..."
        Save-VerifiedDownload -Url $asset.Url -ExpectedSha256 $asset.Sha256 -Destination $archive

        $expanded = Join-Path $temporaryRoot "expanded"
        Expand-Archive -LiteralPath $archive -DestinationPath $expanded
        $candidate = Get-ChildItem -LiteralPath $expanded -Recurse -File -Filter "ffmpeg.exe" |
            Select-Object -First 1
        if (-not $candidate) {
            throw "The verified FFmpeg archive did not contain ffmpeg.exe."
        }
        Assert-WindowsExecutable -Path $candidate.FullName -DisplayName "ffmpeg"

        New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
        $temporaryExecutable = Join-Path $BinDir (".ffmpeg.{0}.download" -f [guid]::NewGuid())
        Copy-Item -LiteralPath $candidate.FullName -Destination $temporaryExecutable
        Move-Item -LiteralPath $temporaryExecutable -Destination $destination
        Write-Host "Verified FFmpeg archive SHA-256: $($asset.Sha256)"
    } finally {
        if (Test-Path -LiteralPath $temporaryRoot) {
            Remove-Item -LiteralPath $temporaryRoot -Recurse -Force
        }
    }
}

$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
Push-Location $repoRoot
try {
    $binDir = Join-Path $repoRoot "bin"
    Ensure-YtDlp -BinDir $binDir
    Ensure-FFmpeg -BinDir $binDir

    Push-Location "frontend"
    try {
        npm run build
        if ($LASTEXITCODE -ne 0) { throw "Frontend build failed" }
    } finally {
        Pop-Location
    }

    $env:GOCACHE = Join-Path $repoRoot ".gocache"
    $versionInfoTool = Get-Command "goversioninfo" -ErrorAction SilentlyContinue
    if ($versionInfoTool) {
        & $versionInfoTool.Source -64 -o "resource.syso" "versioninfo.json"
    } else {
        go run "github.com/josephspurrier/goversioninfo/cmd/goversioninfo@v1.7.0" -64 -o "resource.syso" "versioninfo.json"
    }
    if ($LASTEXITCODE -ne 0) {
        throw "goversioninfo failed with exit code $LASTEXITCODE"
    }

    go build -buildvcs=false -trimpath -tags "desktop,production" -ldflags "-H windowsgui -s -w" -o "OmniVideoRip.exe" .
    if ($LASTEXITCODE -ne 0) { throw "Go build failed" }
} finally {
    Pop-Location
}
