Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Set-Location $repoRoot

function Invoke-NativeCommand {
    param(
        [Parameter(Mandatory = $true)]
        [string]$FilePath,
        [string[]]$Arguments = @()
    )

    & $FilePath @Arguments
    if ($LASTEXITCODE -ne 0) {
        $argText = if ($Arguments.Count -gt 0) { " $($Arguments -join ' ')" } else { "" }
        throw "$FilePath$argText failed with exit code $LASTEXITCODE"
    }
}

function Assert-RequiredSourceFiles {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Action,
        [Parameter(Mandatory = $true)]
        [string[]]$Paths
    )

    $missing = @()
    foreach ($relativePath in $Paths) {
        $fullPath = Join-Path $repoRoot $relativePath
        if (-not (Test-Path -LiteralPath $fullPath -PathType Leaf)) {
            $missing += $relativePath
        }
    }

    if ($missing.Count -gt 0) {
        throw "$Action requires a complete source tree. Missing files: $($missing -join ', ')"
    }
}

try {
    Write-Host "========================================"
    Write-Host "  链享浏览器 - Build Script"
    Write-Host "========================================"
    Write-Host ""
    Write-Host "Current workdir: $repoRoot"
    Write-Host ""

    $proxyValue = [string]$env:LIANXIANG_BUILD_PROXY
    if (-not [string]::IsNullOrWhiteSpace($proxyValue)) {
        $proxyUri = $null
        if (-not [Uri]::TryCreate($proxyValue, [UriKind]::Absolute, [ref]$proxyUri) -or
            $proxyUri.Scheme -notin @("http", "https")) {
            throw "LIANXIANG_BUILD_PROXY must be an absolute HTTP(S) URL"
        }

        Write-Host "[0/7] Configuring process-local build proxy..."
        $env:HTTP_PROXY = $proxyValue
        $env:HTTPS_PROXY = $proxyValue
        $env:http_proxy = $proxyValue
        $env:https_proxy = $proxyValue
        $env:npm_config_proxy = $proxyValue
        $env:npm_config_https_proxy = $proxyValue

        Write-Host "OK process-local proxy configured"
        Write-Host ""
    }

    if (-not [string]::IsNullOrWhiteSpace([string]$env:LIANXIANG_GOPROXY)) {
        $env:GOPROXY = [string]$env:LIANXIANG_GOPROXY
    }

    Assert-RequiredSourceFiles -Action "Building from source" -Paths @(
        "go.mod",
        "go.sum",
        "main.go",
        "wails.json"
    )

    Write-Host "[1/7] Installing frontend dependencies..."
    Push-Location (Join-Path $repoRoot "frontend")
    try {
        Invoke-NativeCommand -FilePath "npm.cmd" -Arguments @("install")
        Invoke-NativeCommand -FilePath "npm.cmd" -Arguments @("run", "ensure:native")
    }
    finally {
        Pop-Location
    }

    Write-Host ""
    Write-Host "[2/7] Installing Go dependencies..."
    Invoke-NativeCommand -FilePath "go" -Arguments @("mod", "download")
    Invoke-NativeCommand -FilePath "go" -Arguments @("mod", "tidy")

    Write-Host ""
    Write-Host "[3/7] Ensuring frontend\dist exists..."
    $frontendDist = Join-Path $repoRoot "frontend/dist"
    $tempDistCreated = $false
    if (-not (Test-Path -LiteralPath $frontendDist)) {
        New-Item -ItemType Directory -Path $frontendDist -Force | Out-Null
        Set-Content -LiteralPath (Join-Path $frontendDist "index.html") -Value "" -Encoding ascii
        $tempDistCreated = $true
        Write-Host "OK temporary dist directory created"
    } else {
        Write-Host "OK dist directory already exists"
    }

    Write-Host ""
    Write-Host "[4/7] Generating Wails bindings..."
    Invoke-NativeCommand -FilePath "cmd" -Arguments @("/c", "call bat\generate-bindings.bat --no-pause")

    $binaryPath = Join-Path $repoRoot "build/bin/lianxiang-browser.exe"

    Write-Host ""
    Write-Host "[5/7] Building frontend..."
    if ($tempDistCreated -and (Test-Path -LiteralPath $frontendDist)) {
        Remove-Item -LiteralPath $frontendDist -Recurse -Force -ErrorAction SilentlyContinue
    }
    Push-Location (Join-Path $repoRoot "frontend")
    try {
        Invoke-NativeCommand -FilePath "npm.cmd" -Arguments @("run", "build:clean")
    }
    finally {
        Pop-Location
    }

    Write-Host ""
    Write-Host "[6/7] Building app..."
    Invoke-NativeCommand -FilePath "wails" -Arguments @("build")

    if ($tempDistCreated -and (Test-Path -LiteralPath $frontendDist)) {
        Remove-Item -LiteralPath $frontendDist -Recurse -Force -ErrorAction SilentlyContinue
    }

    Write-Host ""
    Write-Host "[7/7] Copying runtime dependencies..."
    $binDir = Join-Path $repoRoot "bin"
    $targetDir = Join-Path $repoRoot "build/bin/bin"
    if (Test-Path -LiteralPath $binDir -PathType Container) {
        Copy-Item -LiteralPath $binDir -Destination $targetDir -Recurse -Force
        Write-Host "OK copied bin directory to build\bin\bin\"
    } else {
        Write-Host "[WARN] bin directory not found, skipping copy"
    }

    Write-Host ""
    Write-Host "========================================"
    Write-Host "  OK build completed"
    Write-Host "========================================"
    Write-Host ""
    Write-Host "Executable: build\bin\lianxiang-browser.exe"
    exit 0
}
catch {
    Write-Host ""
    Write-Host "[ERROR] $($_.Exception.Message)"
    exit 1
}
