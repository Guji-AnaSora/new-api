# Windows-friendly development startup script for new-api
param()

$ErrorActionPreference = 'Stop'
$BACKEND_PORT = 3000
$FRONTEND_PORT = 5173

function Write-Color {
    param(
        [string]$Text,
        [string]$Color = 'White'
    )
    Write-Host $Text -ForegroundColor $Color
}

function Ask-YesNo {
    param(
        [string]$Prompt
    )
    $answer = Read-Host "$Prompt [y/N]"
    return $answer -match '^[Yy]'
}

function Check-Port {
    param(
        [int]$Port,
        [string]$Name
    )

    try {
        $conn = Get-NetTCPConnection -LocalPort $Port -ErrorAction Stop
    } catch {
        $conn = $null
    }

    if ($conn) {
        $owningPid = $conn.OwningProcess
        Write-Color "Warning: port $Port ($Name) is in use (PID: $owningPid)" Red
        try {
            $proc = Get-Process -Id $owningPid -ErrorAction Stop
            Write-Color "Process: $($proc.Path)" Yellow
        } catch {
            Write-Color "Process: PID $owningPid" Yellow
        }

        if (Ask-YesNo 'Attempt to kill this process and continue?') {
            Stop-Process -Id $owningPid -Force -ErrorAction SilentlyContinue
            Start-Sleep -Seconds 1
        } else {
            Write-Color 'Startup aborted. Please free the port or change configuration.' Red
            exit 1
        }
    }
}

Write-Color '>>> Initializing New API development environment...' Cyan
Check-Port -Port $BACKEND_PORT -Name 'backend service'
Check-Port -Port $FRONTEND_PORT -Name 'frontend dev server'

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Color 'Error: Go is not found. Install Go 1.22+ first.' Red
    exit 1
}

if (-not (Get-Command bun -ErrorAction SilentlyContinue)) {
    Write-Color 'Error: Bun is not found. Install Bun for frontend support.' Red
    exit 1
}

$goPath = (& go env GOPATH) 2>$null
$airPath = $null
if (Get-Command air -ErrorAction SilentlyContinue) {
    $airPath = (Get-Command air).Source
} elseif ($goPath) {
    $testAir = Join-Path $goPath 'bin\air.exe'
    if (Test-Path $testAir) { $airPath = $testAir }
}

if (-not $airPath) {
    Write-Color '>>> Air is not installed, trying to install it for backend hot reload...' Yellow
    try {
        go install github.com/air-verse/air@latest
        $airPath = (Get-Command air -ErrorAction SilentlyContinue).Source
        if (-not $airPath -and $goPath) {
            $testAir = Join-Path $goPath 'bin\air.exe'
            if (Test-Path $testAir) { $airPath = $testAir }
        }
        if ($airPath) {
            Write-Color '>>> Air installed successfully!' Green
        } else {
            Write-Color '>>> Air install failed, backend will use normal mode.' Yellow
        }
    } catch {
        Write-Color '>>> Air install failed, backend will use normal mode.' Yellow
    }
} else {
    Write-Color '>>> Air found, backend will use hot reload mode!' Green
}

Write-Color '>>> Checking backend dependencies...' Cyan
go mod download

Write-Color '>>> Checking frontend dependencies with Bun...' Cyan
Push-Location web
if (-not (Test-Path 'node_modules')) {
    Write-Color '>>> Installing frontend dependencies, this may take a moment...' Yellow
    bun install
}
Pop-Location

if (-not (Test-Path 'web\dist')) {
    Write-Color '>>> Creating temporary web/dist directory...' Cyan
    New-Item -ItemType Directory -Path 'web\dist' -Force | Out-Null
    New-Item -ItemType File -Path 'web\dist\index.html' -Force | Out-Null
}

Write-Color '>>> Ready to start services...' Green

if ($airPath) {
    Write-Color '>>> Air found, using Air for hot reload...' Green
    $backendProcess = Start-Process -FilePath $airPath -WorkingDirectory $PWD -NoNewWindow -PassThru
} else {
    Write-Color '>>> Air not available, using go run for backend...' Yellow
    $backendProcess = Start-Process -FilePath 'go' -ArgumentList 'run','main.go' -WorkingDirectory $PWD -NoNewWindow -PassThru
}

$frontendProcess = Start-Process -FilePath 'bun' -ArgumentList 'run','dev' -WorkingDirectory (Join-Path $PWD 'web') -NoNewWindow -PassThru

Write-Color '>>> Services started!' Green
Write-Color "Backend: http://localhost:$BACKEND_PORT" Cyan
Write-Color "Frontend: http://localhost:$FRONTEND_PORT" Cyan
Write-Color '>>> Press Ctrl+C to stop all services' Yellow

$cancelHandler = Register-EngineEvent PowerShell.Exiting -Action {
    if (-not $backendProcess.HasExited) { Stop-Process -Id $backendProcess.Id -Force -ErrorAction SilentlyContinue }
    if (-not $frontendProcess.HasExited) { Stop-Process -Id $frontendProcess.Id -Force -ErrorAction SilentlyContinue }
}

try {
    Wait-Process -Id $backendProcess.Id,$frontendProcess.Id
} finally {
    Unregister-Event -SourceIdentifier $cancelHandler.Name -ErrorAction SilentlyContinue
    if (-not $backendProcess.HasExited) { Stop-Process -Id $backendProcess.Id -Force -ErrorAction SilentlyContinue }
    if (-not $frontendProcess.HasExited) { Stop-Process -Id $frontendProcess.Id -Force -ErrorAction SilentlyContinue }
}
