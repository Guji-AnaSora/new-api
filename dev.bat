@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul

REM --- New API 增强型开发启动脚本 (Windows) ---

echo ^>^>^> 正在初始化 New API 开发环境...

REM 端口定义
set BACKEND_PORT=3000
set FRONTEND_PORT=5173

REM 端口检查函数 (通过 findstr 模拟)
echo.
echo 检查端口占用...

REM 检查后端端口
set PORT_BUSY=0
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%BACKEND_PORT%" ^| findstr "LISTENING"') do (
    set PORT_BUSY=1
    set PID=%%a
)
if !PORT_BUSY! equ 1 (
    echo [警告] 端口 %BACKEND_PORT% (后端) 已被占用 PID: !PID!
    for /f "skip=1" %%b in ('wmic process where "processid=!PID!" get name') do (
        if "%%b" neq "" (
            echo 占用程序: %%b
            goto :breakprocname
        )
    )
    :breakprocname
    set /p CONFIRM="是否强制杀掉该进程并继续？[y/N] "
    if /i "!CONFIRM!"=="y" (
        echo 正在清理端口 %BACKEND_PORT%...
        taskkill /F /PID !PID! >nul 2>&1
        timeout /t 1 /nobreak >nul
    ) else (
        echo 启动中止。请手动处理端口占用或修改配置。
        pause
        exit /b 1
    )
)

REM 检查前端端口
set PORT_BUSY=0
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%FRONTEND_PORT%" ^| findstr "LISTENING"') do (
    set PORT_BUSY=1
    set PID=%%a
)
if !PORT_BUSY! equ 1 (
    echo [警告] 端口 %FRONTEND_PORT% (前端) 已被占用 PID: !PID!
    set /p CONFIRM="是否强制杀掉该进程并继续？[y/N] "
    if /i "!CONFIRM!"=="y" (
        echo 正在清理端口 %FRONTEND_PORT%...
        taskkill /F /PID !PID! >nul 2>&1
        timeout /t 1 /nobreak >nul
    ) else (
        echo 启动中止。请手动处理端口占用或修改配置。
        pause
        exit /b 1
    )
)

REM 1. 检查基础环境
echo.
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] 未找到 Go 环境。请先安装 Go 1.22+
    pause
    exit /b 1
)

where bun >nul 2>&1
if %errorlevel% neq 0 (
    echo [警告] 未找到 Bun 环境。建议安装 Bun 以获得最佳体验。
    set USE_NPM=1
) else (
    set USE_NPM=0
)

REM 2. 检查并安装 Air (后端热更新)
echo.
where air >nul 2>&1
if %errorlevel% neq 0 (
    REM 检查 GOPATH/bin/air
    for /f %%i in ('go env GOPATH') do set GOPATH_BIN=%%i
    if exist "!GOPATH_BIN!\bin\air.exe" (
        set AIR_FOUND=1
    ) else (
        set AIR_FOUND=0
        echo [提示] 未找到 Air，正在尝试安装以支持后端热更新...
        go install github.com/air-verse/air@latest
        if !errorlevel! equ 0 (
            echo [成功] Air 安装成功！
            set AIR_FOUND=1
        ) else (
            echo [警告] Air 安装失败，后端将回退到普通模式 (不支持热更新)。
            set AIR_FOUND=0
        )
    )
) else (
    set AIR_FOUND=1
)

REM 3. 初始化后端依赖
echo.
echo 检查后端依赖...
go mod download

REM 4. 初始化前端依赖
echo.
echo 检查前端依赖...
pushd web
if not exist "node_modules" (
    echo 正在安装前端依赖，这可能需要一点时间...
    if !USE_NPM! equ 1 (
        call npm install
    ) else (
        call bun install
    )
)
popd

REM 5. 确保 web\dist 存在 (Go embed 必须)
if not exist "web\dist" (
    echo 创建临时 dist 目录...
    mkdir web\dist
    echo. > web\dist\index.html
)

REM 启动流程
echo.
echo 准备就绪，启动服务...

REM 启动后端 (使用唯一窗口标题以便后续清理)
if !AIR_FOUND! equ 1 (
    start "new-api-backend" cmd /c "air"
) else (
    start "new-api-backend" cmd /c "go run main.go"
)

REM 等待后端启动
timeout /t 2 /nobreak >nul

REM 启动前端
pushd web
if !USE_NPM! equ 1 (
    start "new-api-frontend" cmd /c "npm run dev"
) else (
    start "new-api-frontend" cmd /c "bun run dev"
)
popd

echo.
echo ========================================
echo   服务已全部启动！
echo   后端: http://localhost:%BACKEND_PORT%
echo   前端: http://localhost:%FRONTEND_PORT% (支持热更新)
echo ========================================
echo.
echo 按任意键停止所有服务并退出...
echo.

pause >nul

REM 清理逻辑
echo.
echo ^>^>^> 正在清理并退出...

REM 尝试优雅关闭 (taskkill /T 会连带子进程一起杀掉)
taskkill /F /FI "WINDOWTITLE eq new-api-backend*" /T >nul 2>&1
timeout /t 1 /nobreak >nul
taskkill /F /FI "WINDOWTITLE eq new-api-frontend*" /T >nul 2>&1

REM 如果还有残留 (如 air 启动的 go run, bun 启动的 node), 按进程名补杀
taskkill /F /IM "air.exe" >nul 2>&1
taskkill /F /IM "go.exe" >nul 2>&1
taskkill /F /IM "bun.exe" >nul 2>&1
taskkill /F /IM "node.exe" >nul 2>&1

echo 所有服务已停止。
timeout /t 1 /nobreak >nul
