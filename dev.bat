@echo off
setlocal enabledelayedexpansion

echo Initializing New API Development Environment...

REM Port definitions
set BACKEND_PORT=3000
set FRONTEND_PORT=5173

REM Check port usage
echo Checking port usage...
netstat -ano | findstr ":%BACKEND_PORT%" | findstr "LISTENING" >nul 2>&1
if %errorlevel% equ 0 (
    echo Warning: Port %BACKEND_PORT% is in use, cleaning up...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%BACKEND_PORT%" ^| findstr "LISTENING"') do taskkill /F /PID %%a >nul 2>&1
    timeout /t 1 /nobreak >nul
)

netstat -ano | findstr ":%FRONTEND_PORT%" | findstr "LISTENING" >nul 2>&1
if %errorlevel% equ 0 (
    echo Warning: Port %FRONTEND_PORT% is in use, cleaning up...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%FRONTEND_PORT%" ^| findstr "LISTENING"') do taskkill /F /PID %%a >nul 2>&1
    timeout /t 1 /nobreak >nul
)

REM Check Go environment
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go not found. Please install Go 1.22+
    pause
    exit /b 1
)

REM Check Bun environment
where bun >nul 2>&1
if %errorlevel% neq 0 (
    echo Warning: Bun not found, trying npm...
    set USE_NPM=1
) else (
    set USE_NPM=0
)

REM Check and install Air
where air >nul 2>&1
if %errorlevel% neq 0 (
    echo Air not found, installing for backend hot reload...
    go install github.com/air-verse/air@latest
    if %errorlevel% equ 0 (
        echo Air installed successfully!
    ) else (
        echo Air installation failed, backend will use normal mode.
    )
)

REM Initialize backend dependencies
echo Checking backend dependencies...
go mod download

REM Initialize frontend dependencies
echo Checking frontend dependencies...
cd web
if not exist "node_modules" (
    echo Installing frontend dependencies, this may take a while...
    if !USE_NPM! equ 1 (
        call npm install
    ) else (
        call bun install
    )
)
cd ..

REM Ensure web/dist exists
if not exist "web\dist" (
    echo Creating temporary dist directory...
    mkdir web\dist
    echo. > web\dist\index.html
)

REM Start services
echo Ready, starting services...

REM Start backend
where air >nul 2>&1
if %errorlevel% equ 0 (
    start "New-API Backend (Air)" cmd /c "air"
) else (
    start "New-API Backend" cmd /c "go run main.go"
)

REM Wait for backend to start
timeout /t 2 /nobreak >nul

REM Start frontend
cd web
if !USE_NPM! equ 1 (
    start "New-API Frontend (Vite)" cmd /c "npm run dev"
) else (
    start "New-API Frontend (Vite)" cmd /c "bun run dev"
)
cd ..

echo.
echo All services started!
echo Backend: http://localhost:%BACKEND_PORT%
echo Frontend: http://localhost:%FRONTEND_PORT% (with hot reload)
echo Close this window or press Ctrl+C to stop all services
echo.

REM Keep window open
pause
