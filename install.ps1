param(
    [string]$Version = '',
    [string]$InstallDir = '',
    [switch]$Source
)

$ErrorActionPreference = 'Stop'

$Repo = 'amit-vikramaditya/v1claw'
$Binary = 'v1claw.exe'

function Write-Info {
    param([string]$Message)
    Write-Host "  ->  $Message" -ForegroundColor Blue
}

function Write-Ok {
    param([string]$Message)
    Write-Host "  [ok] $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "  [!]  $Message" -ForegroundColor Yellow
}

function Fail {
    param([string]$Message)
    Write-Host "  [x]  $Message" -ForegroundColor Red
    exit 1
}

function Download-File {
    param(
        [string]$Url,
        [string]$OutFile
    )
    Invoke-WebRequest -Uri $Url -OutFile $OutFile
}

function Get-LatestReleaseVersion {
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        return [string]$release.tag_name
    } catch {
        return ''
    }
}

function Get-ArchiveArch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
    switch ($arch) {
        'x64' { return 'x86_64' }
        'arm64' { return 'arm64' }
        default { Fail "Unsupported Windows architecture: $arch. Download a release manually from https://github.com/$Repo/releases" }
    }
}

function Ensure-InstallDir {
    if ($InstallDir) {
        return $InstallDir
    }
    if ($env:INSTALL_DIR) {
        return $env:INSTALL_DIR
    }

    $localAppData = [Environment]::GetFolderPath('LocalApplicationData')
    if (-not $localAppData) {
        $localAppData = Join-Path $HOME 'AppData\Local'
    }
    return Join-Path $localAppData 'Programs\V1Claw\bin'
}

function Add-InstallDirToPath {
    param([string]$Dir)

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $segments = @()
    if ($userPath) {
        $segments = $userPath.Split(';', [System.StringSplitOptions]::RemoveEmptyEntries)
    }

    $alreadyPresent = $false
    foreach ($segment in $segments) {
        if ($segment.TrimEnd('\') -ieq $Dir.TrimEnd('\')) {
            $alreadyPresent = $true
            break
        }
    }

    if (-not $alreadyPresent) {
        $newPath = if ($userPath) { "$userPath;$Dir" } else { $Dir }
        [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
        Write-Ok "Added $Dir to your user PATH"
    } else {
        Write-Ok "$Dir is already in your user PATH"
    }

    if (-not (($env:Path.Split(';', [System.StringSplitOptions]::RemoveEmptyEntries)) -contains $Dir)) {
        $env:Path = "$env:Path;$Dir"
    }
}

function Test-PathContainsDir {
    param(
        [string]$PathValue,
        [string]$Dir
    )

    if (-not $PathValue) {
        return $false
    }

    foreach ($segment in $PathValue.Split(';', [System.StringSplitOptions]::RemoveEmptyEntries)) {
        if ($segment.TrimEnd('\') -ieq $Dir.TrimEnd('\')) {
            return $true
        }
    }
    return $false
}

function Get-SourceDir {
    param([string]$TempDir)

    if ((Test-Path '.\go.mod') -and (Test-Path '.\cmd\v1claw')) {
        Write-Info "Using local source checkout at $(Get-Location)"
        return @{
            Path = (Get-Location).Path
            Description = 'local source build'
        }
    }

    $sourceZip = Join-Path $TempDir 'source.zip'
    Write-Info 'Downloading source archive...'
    Download-File -Url "https://github.com/$Repo/archive/refs/heads/main.zip" -OutFile $sourceZip

    $extractDir = Join-Path $TempDir 'source'
    Expand-Archive -Path $sourceZip -DestinationPath $extractDir -Force

    $sourceDir = Get-ChildItem -Path $extractDir -Directory | Select-Object -First 1
    if (-not $sourceDir) {
        Fail 'Could not locate extracted source directory.'
    }

    return @{
        Path = $sourceDir.FullName
        Description = 'source build from main'
    }
}

$ArchiveArch = Get-ArchiveArch
$forceSource = $Source -or ($env:V1CLAW_INSTALL_MODE -eq 'source')
if ((-not $forceSource) -and (-not $Version)) {
    Write-Info 'Fetching latest release version...'
    $Version = Get-LatestReleaseVersion
}

$InstallMode = 'release'
if ($forceSource) {
    $InstallMode = 'source'
} elseif (-not $Version) {
    $InstallMode = 'source'
    Write-Warn "No published GitHub release was found for $Repo."
    $goCmd = Get-Command go -ErrorAction SilentlyContinue
    if (-not $goCmd) {
        Fail 'No release is available and Go is not installed. Install Go or publish a release first.'
    }
    Write-Warn 'Falling back to a source build because Go is installed.'
} else {
    Write-Ok "Version: $Version"
}

if ($InstallMode -eq 'source' -and -not (Get-Command go -ErrorAction SilentlyContinue)) {
    Fail "Source installation requires Go, but 'go' was not found in PATH."
}

$installDirExplicit = [bool]($InstallDir -or $env:INSTALL_DIR)
$resolvedInstallDir = Ensure-InstallDir
New-Item -ItemType Directory -Force -Path $resolvedInstallDir | Out-Null

$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("v1claw-install-" + [System.Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Force -Path $tempDir | Out-Null

try {
    $binaryPath = Join-Path $tempDir $Binary

    if ($InstallMode -eq 'release') {
        $archive = "v1claw_Windows_${ArchiveArch}.zip"
        $url = "https://github.com/$Repo/releases/download/$Version/$archive"
        $archivePath = Join-Path $tempDir $archive

        Write-Info "Downloading $archive..."
        Download-File -Url $url -OutFile $archivePath

        Write-Info 'Extracting...'
        Expand-Archive -Path $archivePath -DestinationPath $tempDir -Force

        if (-not (Test-Path $binaryPath)) {
            Fail "Binary '$Binary' was not found in the release archive."
        }
    } else {
        $goCmd = Get-Command go -ErrorAction Stop
        $sourceInfo = Get-SourceDir -TempDir $tempDir

        Write-Info "Building from source with $(& $goCmd.Source version)..."
        Push-Location $sourceInfo.Path
        try {
            & $goCmd.Source build -o $binaryPath ./cmd/v1claw
        } finally {
            Pop-Location
        }
        if (-not (Test-Path $binaryPath)) {
            Fail 'Source build failed.'
        }
        $SourceDescription = $sourceInfo.Description
    }

    $targetPath = Join-Path $resolvedInstallDir $Binary
    Copy-Item -Path $binaryPath -Destination $targetPath -Force
    Write-Ok "Installed to $targetPath"

    if ($installDirExplicit) {
        if (-not (Test-PathContainsDir -PathValue $env:Path -Dir $resolvedInstallDir)) {
            Write-Warn "$resolvedInstallDir is not managed automatically because InstallDir was set explicitly."
            Write-Warn "Add it to PATH manually if you want to run 'v1claw' without the full path."
        }
    } else {
        Add-InstallDirToPath -Dir $resolvedInstallDir
    }

    Write-Host ''
    if ($InstallMode -eq 'release') {
        Write-Host "  V1Claw $Version is installed." -ForegroundColor Green
    } else {
        Write-Host "  V1Claw ($SourceDescription) is installed." -ForegroundColor Green
    }
    Write-Host ''
    Write-Host '  Next step - run the setup wizard:'
    Write-Host ''
    Write-Host '    v1claw onboard'
    Write-Host ''
    Write-Host '  Or silent cloud setup:'
    Write-Host ''
    Write-Host '    v1claw onboard --auto --provider gemini --api-key YOUR_KEY'
    Write-Host ''
    Write-Host '  Local setup example:'
    Write-Host ''
    Write-Host '    v1claw onboard --auto --provider ollama --model llama3.2'
    Write-Host ''
    Write-Host '  For CI or offline setup, add:'
    Write-Host ''
    Write-Host '    --skip-test'
    Write-Host ''
} finally {
    if (Test-Path $tempDir) {
        Remove-Item -Path $tempDir -Recurse -Force
    }
}
