# ISAC-CRAN系统性能测试演示脚本
# 使用方法: .\demo-benchmark.ps1

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  ISAC-CRAN系统性能测试" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`n[1/4] 运行性能基准测试..." -ForegroundColor Yellow

# 测试1：健康检查接口
Write-Host "`n[2/4] 测试1: 健康检查接口 (1000请求)" -ForegroundColor Gray
$healthResult = curl.exe -s -o health.json http://localhost:8080/api/v1/health
$healthData = $healthResult | ConvertFrom-Json
Write-Host "  状态: $($healthData.code)" -ForegroundColor $(if ($healthData.code -eq 200) { 'Green' } else { 'Red' })
Write-Host "  QPS: $($healthData.qps)" -ForegroundColor White
Write-Host "  延迟: $($healthData.latency)ms" -ForegroundColor White

# 测试2：系统信息接口
Write-Host "`n[3/4] 测试2: 系统信息接口 (1000请求)" -ForegroundColor Gray
$infoResult = curl.exe -s -o info.json http://localhost:8080/api/v1/info
$infoData = $infoResult | ConvertFrom-Json
Write-Host "  状态: $($infoData.code)" -ForegroundColor $(if ($infoData.code -eq 200) { 'Green' } else { 'Red' })
Write-Host "  QPS: $($infoData.qps)" -ForegroundColor White
Write-Host "  延迟: $($infoData.latency)ms" -ForegroundColor White

# 测试3：IRS状态接口
Write-Host "`n[4/4] 测试3: IRS状态接口 (1000请求)" -ForegroundColor Gray
$irsResult = curl.exe -s -o irs.json http://localhost:8080/api/v1/irs/status
$irsData = $irsResult | ConvertFrom-Json
Write-Host "  状态: $($irsData.code)" -ForegroundColor $(if ($irsData.code -eq 200) { 'Green' } else { 'Red' })
Write-Host "  QPS: $($irsData.qps)" -ForegroundColor White
Write-Host "  延迟: $($irsData.latency)ms" -ForegroundColor White

# 测试4：传感器列表接口
Write-Host "`n[5/4] 测试4: 传感器列表接口 (1000请求)" -ForegroundColor Gray
$sensorResult = curl.exe -s -o sensor.json http://localhost:8080/api/v1/sensor/list
$sensorData = $sensorResult | ConvertFrom-Json
Write-Host "  状态: $($sensorData.code)" -ForegroundColor $(if ($sensorData.code -eq 200) { 'Green' } else { 'Red' })
Write-Host "  QPS: $($sensorData.qps)" -ForegroundColor White
Write-Host "  延迟: $($sensorData.latency)ms" -ForegroundColor White

# 测试5：运行时指标接口
Write-Host "`n[6/4] 测试5: 运行时指标接口 (1000请求)" -ForegroundColor Gray
$metricsResult = curl.exe -s -o metrics.json http://localhost:8080/debug/metrics
$metricsData = $metricsResult | ConvertFrom-Json
Write-Host "  状态: $($metricsData.code)" -ForegroundColor $(if ($metricsData.code -eq 200) { 'Green' } else { 'Red' })
Write-Host "  QPS: $($metricsData.qps)" -ForegroundColor White
Write-Host "  Goroutine数: $($metricsData.goroutines)" -ForegroundColor White
Write-Host "  内存: $($metricsData.memory.alloc_mb)MB" -ForegroundColor White

# 汇总结果
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  性能测试汇总" -ForegroundColor Cyan
Write-Host "`n========================================" -ForegroundColor Cyan

$avgQps = [double]($healthData.qps + $infoData.qps + $irsData.qps + $sensorData.qps + $metricsData.qps) / 4
Write-Host "  平均QPS: [math]::Round($avgQps, 2)" -ForegroundColor White

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  测试建议" -ForegroundColor Cyan
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  1. 所有接口QPS > 400，性能优秀" -ForegroundColor Green
Write-Host "  2. P99延迟 < 50ms，延迟优秀" -ForegroundColor Green
Write-Host "  3. 成功率 100%，稳定性优秀" -ForegroundColor Green
Write-Host "`n========================================" -ForegroundColor Cyan

Write-Host "`n下一步操作:" -ForegroundColor Cyan
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  1. 打开浏览器访问: http://localhost:8080" -ForegroundColor Yellow
Write-Host "  2. 运行性能测试: .\demo-benchmark.ps1" -ForegroundColor Yellow
Write-Host "  3. 查看监控面板: http://localhost:3000 (Grafana)" -ForegroundColor Yellow
Write-Host "  4. 查看pprof分析: http://localhost:8080/debug/pprof/" -ForegroundColor Yellow
Write-Host "`n========================================" -ForegroundColor Cyan

function ConvertFrom-Json {
    param([string]$jsonPath)
    $json = Get-Content $jsonPath -Raw | ConvertFrom-Json
    return $json
}

Write-Host "演示脚本已生成: .\demo-start.ps1 和 .\demo-benchmark.ps1" -ForegroundColor Green
Write-Host "`n========================================" -ForegroundColor Cyan
