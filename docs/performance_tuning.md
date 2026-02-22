# ISAC-CRAN系统性能调优指南

## 1. 性能分析工具

### 1.1 pprof使用

系统已集成pprof，可通过以下端点访问：

```bash
# 查看所有性能分析端点
curl http://localhost:8080/debug/pprof/

# CPU分析（采样30秒）
curl -o cpu.prof http://localhost:8080/debug/pprof/profile?seconds=30

# 堆内存分析
curl -o heap.prof http://localhost:8080/debug/pprof/heap

# Goroutine分析
curl -o goroutine.prof http://localhost:8080/debug/pprof/goroutine

# 阻塞分析（需要开启）
curl -o block.prof http://localhost:8080/debug/pprof/block

# 互斥锁分析（需要开启）
curl -o mutex.prof http://localhost:8080/debug/pprof/mutex
```

### 1.2 使用go tool pprof分析

```bash
# 交互式分析CPU
go tool pprof cpu.prof

# 常用命令
(pprof) top10          # 显示CPU消耗最高的10个函数
(pprof) list funcName  # 查看具体函数代码
(pprof) web            # 生成火焰图（需要graphviz）

# 分析堆内存
go tool pprof heap.prof
(pprof) top10          # 显示内存分配最多的函数
(pprof) allocs         # 查看总分配量

# 对比两个时间点的堆内存
go tool pprof -base=heap1.prof heap2.prof
```

### 1.3 运行时指标

```bash
# 获取运行时指标
curl http://localhost:8080/debug/metrics

# 响应示例
{
  "goroutines": 25,
  "memory": {
    "alloc_mb": 15.5,
    "total_mb": 120.3,
    "sys_mb": 45.2,
    "gc_count": 12,
    "gc_pause_s": 0.002
  },
  "cpu_cores": 8,
  "go_version": "go1.21.0"
}
```

## 2. 常见性能瓶颈与优化

### 2.1 内存泄漏排查

**症状**：
- 内存持续增长不释放
- GC频率异常高
- OOM崩溃

**排查步骤**：

```bash
# 1. 获取基线堆内存快照
curl -o heap1.prof http://localhost:8080/debug/pprof/heap

# 2. 运行一段时间后获取第二个快照
curl -o heap2.prof http://localhost:8080/debug/pprof/heap

# 3. 对比分析
go tool pprof -base=heap1.prof heap2.prof

# 4. 查找增长的对象
(pprof) top10
```

**常见原因**：
1. **Goroutine泄漏**：未正确关闭的goroutine
2. **闭包引用**：闭包捕获大对象导致无法GC
3. **全局变量**：无限增长的全局map/slice
4. **未关闭的资源**：文件、连接未关闭

**修复示例**：

```go
// ❌ 错误：goroutine泄漏
func processItems(items []Item) {
    for _, item := range items {
        go func() {
            process(item) // 捕获item变量
        }()
    }
}

// ✅ 正确：传递参数
func processItems(items []Item) {
    var wg sync.WaitGroup
    for _, item := range items {
        wg.Add(1)
        go func(i Item) {
            defer wg.Done()
            process(i)
        }(item)
    }
    wg.Wait()
}
```

### 2.2 CPU热点优化

**症状**：
- CPU使用率高
- 响应延迟大
- 吞吐量低

**排查步骤**：

```bash
# CPU分析
curl -o cpu.prof http://localhost:8080/debug/pprof/profile?seconds=30
go tool pprof cpu.prof

# 查看热点函数
(pprof) top10 -cum
```

**常见热点**：

| 热点类型 | 优化方法 |
|---------|---------|
| JSON序列化 | 使用sonic或json-iterator替代encoding/json |
| 字符串拼接 | 使用strings.Builder替代+拼接 |
| 内存分配 | 使用sync.Pool复用对象 |
| 锁竞争 | 使用读写锁或无锁数据结构 |

**优化示例**：

```go
// ❌ 错误：频繁内存分配
func processRequest(data []byte) Result {
    var result Result
    json.Unmarshal(data, &result) // 每次分配
    return result
}

// ✅ 正确：对象池复用
var resultPool = sync.Pool{
    New: func() interface{} {
        return &Result{}
    },
}

func processRequest(data []byte) *Result {
    result := resultPool.Get().(*Result)
    defer resultPool.Put(result)
    json.Unmarshal(data, result)
    return result
}
```

### 2.3 锁竞争优化

**症状**：
- 多核CPU利用率低
- 响应延迟波动大
- pprof显示runtime.selectgo或sync.Mutex热点

**排查方法**：

```bash
# 开启锁竞争分析
curl -o mutex.prof http://localhost:8080/debug/pprof/mutex
go tool pprof mutex.prof
```

**优化策略**：

```go
// ❌ 错误：全局大锁
type Cache struct {
    mu    sync.Mutex
    items map[string]interface{}
}

func (c *Cache) Get(key string) interface{} {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.items[key]
}

// ✅ 正确：分片锁
type ShardedCache struct {
    shards [16]struct {
        mu    sync.RWMutex
        items map[string]interface{}
    }
}

func (c *ShardedCache) getShard(key string) int {
    hash := fnv32(key)
    return int(hash % 16)
}

func (c *ShardedCache) Get(key string) interface{} {
    shard := c.shards[c.getShard(key)]
    shard.mu.RLock()
    defer shard.mu.RUnlock()
    return shard.items[key]
}
```

### 2.4 GC压力优化

**症状**：
- GC暂停时间长
- CPU在GC上花费大量时间
- 内存碎片化

**排查方法**：

```bash
# 查看GC统计
curl http://localhost:8080/debug/metrics | jq '.memory.gc_count, .memory.gc_pause_s'

# 使用GODEBUG查看GC日志
GODEBUG=gctrace=1 ./server
```

**优化策略**：

```go
// 1. 预分配切片容量
items := make([]Item, 0, expectedSize)

// 2. 使用值类型替代指针
type Point struct {
    X, Y float64 // 直接存储，减少指针
}

// 3. 避免频繁创建临时对象
var buf bytes.Buffer
for i := 0; i < 1000; i++ {
    buf.Reset()
    buf.WriteString(...)
}

// 4. 调整GOGC（权衡内存和CPU）
// 默认GOGC=100，表示堆增长100%时触发GC
// 增大GOGC减少GC频率但增加内存
GOGC=200 ./server
```

## 3. 基准测试

### 3.1 运行基准测试

```bash
# 编译基准测试工具
go build -o bin/benchmark ./pkg/benchmark

# 运行API基准测试
curl -X POST http://localhost:8080/benchmark/run \
  -H "Content-Type: application/json" \
  -d '{
    "target_url": "http://localhost:8080/api/v1/health",
    "duration": "30s",
    "concurrency": 50
  }'

# 查看测试结果
curl http://localhost:8080/benchmark/results
```

### 3.2 性能指标解读

| 指标 | 良好 | 需关注 | 需优化 |
|------|------|--------|--------|
| P99延迟 | <100ms | 100-500ms | >500ms |
| QPS | >1000 | 100-1000 | <100 |
| 成功率 | >99.9% | 99-99.9% | <99% |
| CPU使用率 | <70% | 70-90% | >90% |
| 内存使用率 | <80% | 80-95% | >95% |

### 3.3 性能回归检测

```bash
# 保存基准结果
curl http://localhost:8080/benchmark/results > baseline.json

# 代码修改后重新测试
curl -X POST http://localhost:8080/benchmark/run -d '...'

# 对比结果
curl http://localhost:8080/benchmark/results > current.json
```

## 4. Kubernetes环境调优

### 4.1 资源限制配置

```yaml
resources:
  requests:
    cpu: 100m      # 保证资源
    memory: 256Mi
  limits:
    cpu: 500m      # 最大资源
    memory: 512Mi
```

### 4.2 HPA配置

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: isac-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: isac-server
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### 4.3 监控告警

```yaml
# Prometheus告警规则
groups:
- name: isac-alerts
  rules:
  - alert: HighCPU
    expr: rate(container_cpu_usage_seconds_total{pod=~"isac-server.*"}[5m]) > 0.8
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "CPU使用率过高"
      
  - alert: HighMemory
    expr: container_memory_usage_bytes{pod=~"isac-server.*"} / container_spec_memory_limit_bytes{pod=~"isac-server.*"} > 0.9
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "内存使用率过高"
```

## 5. 性能优化检查清单

### 5.1 代码层面
- [ ] 避免在循环中创建临时对象
- [ ] 使用sync.Pool复用对象
- [ ] 预分配切片和map容量
- [ ] 使用strings.Builder拼接字符串
- [ ] 避免全局变量无限增长
- [ ] 正确处理goroutine生命周期
- [ ] 使用context控制超时

### 5.2 数据库层面
- [ ] 使用连接池
- [ ] 创建必要的索引
- [ ] 避免N+1查询
- [ ] 使用批量操作
- [ ] 合理使用事务

### 5.3 网络层面
- [ ] 启用HTTP/2
- [ ] 使用连接复用
- [ ] 启用Gzip压缩
- [ ] 设置合理的超时时间
- [ ] 使用CDN加速静态资源

### 5.4 部署层面
- [ ] 设置合理的资源限制
- [ ] 配置HPA自动扩缩容
- [ ] 启用健康检查
- [ ] 配置监控告警
- [ ] 使用多副本高可用
