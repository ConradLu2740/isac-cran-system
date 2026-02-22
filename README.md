# ISAC-CRAN System

## 智能反射面辅助C-RAN通信感知一体化实验原型系统

### 项目简介

本项目是一个基于Go语言开发的实验原型系统，用于智能反射面（IRS）辅助C-RAN通信感知一体化（ISAC）的研究与验证。系统实现了IRS相移配置、信道数据采集、波束成形优化、DOA估计等核心功能，并提供了完整的RESTful API接口和Web管理界面。

### 技术栈

| 类别 | 技术 |
|------|------|
| 后端框架 | Go 1.21+ / Gin |
| ORM | GORM |
| 时序数据库 | InfluxDB 2.x |
| 关系数据库 | MySQL 8.0 |
| 缓存 | Redis |
| 消息队列 | MQTT |
| 日志 | Zap |
| 前端 | React 18 + Ant Design 5 |
| 图表 | ECharts |

### 项目结构

```
isac-cran-system/
├── cmd/server/           # 主程序入口
├── internal/             # 私有业务代码
│   ├── config/           # 配置管理
│   ├── handler/          # HTTP处理器
│   ├── service/          # 业务逻辑层
│   ├── device/           # 设备控制层（IRS/USRP/传感器）
│   ├── algorithm/        # 算法引擎（波束成形/DOA/调度/编码）
│   ├── repository/       # 数据访问层
│   ├── model/            # 数据模型
│   ├── middleware/       # 中间件
│   └── router/           # 路由配置
├── pkg/                  # 公共库
├── configs/              # 配置文件
├── scripts/              # 脚本文件
├── test/                 # 测试代码
├── web/                  # 前端应用
└── docs/                 # 项目文档
```

### 快速开始

#### 1. 环境要求

- Go 1.21+
- Docker Desktop
- Node.js 18+ (前端开发)

#### 2. 启动数据库

```bash
# 启动MySQL和InfluxDB
docker-compose up -d mysql influxdb

# 初始化InfluxDB
docker exec isac-influxdb influx setup --username admin --password admin123 --org isac-lab --bucket channel-data --force
```

#### 3. 编译运行

```bash
# 编译
go build -o bin/server ./cmd/server

# 运行
./bin/server -config configs/config.yaml
```

#### 4. 启动前端

```bash
cd web
npm install
npm start
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

### 测试

```bash
# 运行所有测试
go test ./... -v

# 运行压力测试
go test ./test/e2e/... -v

# 运行基准测试
go test -bench=. ./...
```

### 性能指标

| 指标 | 数值 |
|------|------|
| QPS | 2174 请求/秒 |
| 平均延迟 | 1.35ms |
| 最大延迟 | 2.68ms |

### 作者

ISAC-CRAN System - 2024

### 许可证

MIT License
