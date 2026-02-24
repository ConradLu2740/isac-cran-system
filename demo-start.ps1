# ISAC-CRAN系统演示启动脚本
# 使用方法: .\demo-start.ps1

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  ISAC-CRAN系统演示环境启动" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# 检查Docker是否运行
Write-Host "`n[1/4] 检查Docker环境..." -ForegroundColor Yellow
$dockerStatus = docker ps 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "请先启动Docker Desktop!" -ForegroundColor Red
    exit 1
}
Write-Host "Docker运行正常" -ForegroundColor Green

# 停止旧容器
Write-Host "`n[2/4] 清理旧容器..." -ForegroundColor Yellow
docker compose down 2>$null

# 启动服务
Write-Host "`n[3/4] 启动所有服务..." -ForegroundColor Yellow
docker compose up -d

# 等待服务就绪
Write-Host "`n[4/4] 等待服务就绪..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# 健康检查
$maxRetry = 30
$retry = 0
while ($retry -lt $maxRetry) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/health" -TimeoutSec 2 -UseBasicParsing
        if ($response.StatusCode -eq 200) {
            break
        }
    } catch {
        Start-Sleep -Seconds 1
        $retry++
        Write-Host "等待服务启动... ($retry/$maxRetry)" -ForegroundColor Gray
    }
}

if ($retry -eq $maxRetry) {
    Write-Host "服务启动失败，请检查日志" -ForegroundColor Red
    docker compose logs server --tail=20
    exit 1
}

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "  服务启动成功!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

Write-Host "`n服务地址:" -ForegroundColor Cyan
Write-Host "  API服务:     http://localhost:8080" -ForegroundColor White
Write-Host "  健康检查:    http://localhost:8080/api/v1/health" -ForegroundColor White
Write-Host "  运行时指标:  http://localhost:8080/debug/metrics" -ForegroundColor White
Write-Host "  pprof分析:   http://localhost:8080/debug/pprof/" -ForegroundColor White
Write-Host "  Web界面:     在浏览器打开 web/dashboard.html" -ForegroundColor White

Write-Host "`n容器状态:" -ForegroundColor Cyan
docker compose ps

Write-Host "`n提示: 运行 '.\demo-benchmark.ps1' 进行性能测试" -ForegroundColor Yellow
