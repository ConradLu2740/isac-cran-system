# ISAC-CRAN System

## 智能反射面辅助C-RAN通信感知一体化实验原型系统

### 项目简介

本项目是一个基于Go语言开发的实验原型系统，用于智能反射面（IRS）辅助C-RAN通信感知一体化（ISAC）的研究与验证。系统实现了IRS相移配置、信道数据采集、波束成形优化、DOA估计等核心功能，并提供了完整的RESTful API接口。

### 技术栈

| 类别 | 技术 |
|------|------|
| 后端框架 | Go 1.24 / Gin |
| ORM | GORM |
| 时序数据库 | InfluxDB 2.7 |
| 关系数据库 | MySQL 8.0 |
| 缓存 | Redis 7.0 |
| 消息队列 | MQTT / RabbitMQ |
| 日志 | Zap |
| 容器化 | Docker / Kubernetes |
| 监控 | Prometheus / Grafana |
| 服务发现 | Consul |

### 项目结构

```
isac-cran-system/
├── cmd/                      # 主程序入口
│   ├── server/               # API服务
│   └── benchmark/            # 性能测试工具
├── internal/                 # 私有业务代码
│   ├── config/               # 配置管理
│   ├── handler/              # HTTP处理器
│   ├── service/              # 业务逻辑层
│   ├── device/               # 设备控制层
│   ├── algorithm/            # 算法引擎
│   ├── repository/           # 数据访问层
│   ├── model/                # 数据模型
│   ├── middleware/           # 中间件
│   └── router/               # 路由配置
├── pkg/                      # 公共库
│   ├── pool/                 # Worker Pool
│   ├── cache/                # 多级缓存
│   ├── queue/                # 异步任务队列
│   ├── errors/               # 错误处理
│   ├── discovery/            # 服务发现
│   ├── mq/                   # 消息队列
│   └── gateway/              # API网关
├── api/                      # API定义
│   ├── proto/                # gRPC协议
│   └── grpc/                 # gRPC服务
├── k8s/                      # Kubernetes配置
├── helm/                     # Helm Chart
├── monitoring/               # 监控配置
├── matlab/                   # MATLAB验证脚本
├── docs/                     # 项目文档
├── scripts/                  # 部署脚本
└── configs/                  # 配置文件
```

### 功能特性

#### 核心功能
- IRS智能反射面控制与优化
- 信道数据采集与存储
- 波束成形算法（MVDR/MMSE/深度学习）
- DOA估计算法（MUSIC/ESPRIT）
- 传感器数据管理

#### 拓展功能
- Worker Pool并发处理
- 多级缓存（L1本地+L2 Redis）
- gRPC高性能通信
- 服务发现与负载均衡
- 消息队列异步处理
- Prometheus监控告警
- Kubernetes云原生部署

### 快速开始

#### 1. 环境要求

- Go 1.24+
- Docker Desktop
- (可选) Minikube / Kubernetes

#### 2. Docker Compose部署（推荐）

```bash
# 启动所有服务
docker compose up -d

# 查看服务状态
docker compose ps

# 访问API
curl http://localhost:8080/api/v1/health
```

#### 3. 本地开发

```bash
# 安装依赖
go mod download

# 编译运行
go build -o bin/server ./cmd/server
./bin/server -config configs/config.yaml
```

#### 4. 性能测试

```bash
# 运行基准测试
go run ./cmd/benchmark

# 单元测试
go test ./... -v
```

### API接口

| 接口 | 方法 | 功能 |
|------|------|------|
| `/api/v1/health` | GET | 健康检查 |
| `/api/v1/info` | GET | 系统信息 |
| `/api/v1/irs/config` | POST | 配置IRS相移 |
| `/api/v1/irs/status` | GET | 获取IRS状态 |
| `/api/v1/irs/optimal` | POST | 应用最优相移 |
| `/api/v1/channel/collect` | POST | 采集信道数据 |
| `/api/v1/channel/data` | GET | 查询信道数据 |
| `/api/v1/algorithm/beamforming` | POST | 运行波束成形 |
| `/api/v1/algorithm/doa` | POST | 运行DOA估计 |
| `/api/v1/sensor/list` | GET | 列出传感器 |
| `/api/v1/sensor/read/:id` | GET | 读取传感器数据 |
| `/debug/metrics` | GET | 运行时指标 |
| `/debug/pprof/` | GET | 性能分析 |

### 性能指标

| 接口 | QPS | P50延迟 | P99延迟 |
|------|-----|---------|---------|
| Health Check | 525 | 18ms | 36ms |
| System Info | 494 | 18ms | 41ms |
| IRS Status | 532 | 18ms | 33ms |
| Sensor List | 510 | 18ms | 34ms |
| Runtime Metrics | 501 | 18ms | 39ms |

**平均QPS: 512**

### Kubernetes部署

```bash
# 使用kubectl
kubectl apply -f k8s/deployment.yaml

# 使用Helm
helm install isac-cran ./helm/isac-cran
```

### 文档

- [性能测试报告](docs/performance_report.md)
- [项目总结报告](docs/project_summary.md)
- [性能调优指南](docs/performance_tuning.md)
- [通信算法设计](docs/algorithm_advanced.md)

### 许可证

MIT License

### 作者

ISAC-CRAN System - 2026
