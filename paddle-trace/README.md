# 🏓 基于区块链的乒乓球拍防伪溯源系统

> **Paddle Traceability System** — 信息安全工程概论课程项目  
> 基于百度超级链（XuperChain）+ Go + NFC 的防伪溯源原型系统

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![XuperChain](https://img.shields.io/badge/XuperChain-5.3.0-blue)](https://xuper.baidu.com/)

---

## 📋 项目简介

本项目以乒乓球拍为切入点，基于信息安全工程方法论，设计并实现了一套基于**百度超级链（XuperChain）**的乒乓球拍防伪溯源原型系统。系统通过**Go语言**编写的智能合约实现产品全生命周期的信息上链存证，确保从原材料采购、生产加工、仓储物流到终端零售的每一个环节数据**真实可信、不可篡改、全程可追溯**。

### 核心特性

- 🔗 **区块链存证** — 基于XuperChain 7节点PBFT联盟链，数据不可篡改
- 📱 **NFC芯片验证** — NTAG 424 DNA SUN动态认证，防标签克隆与重放攻击
- 🛡️ **国密算法** — SM2签名 + SM3哈希，符合国密合规要求
- 🎯 **角色权限控制** — 基于RBAC的智能合约访问控制
- ⚡ **高性能查询** — Redis缓存 + P99响应≤2秒
- 🐳 **一键部署** — Docker Compose编排，环境一键复现

---

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                    前端展示层 (Frontend)                     │
│  ┌──────────┐  ┌──────────────┐  ┌───────────────────┐    │
│  │ 消费者小程序│  │  管理后台Web  │  │ 监管审计Dashboard  │    │
│  └─────┬─────┘  └──────┬───────┘  └────────┬──────────┘    │
└────────┼────────────────┼───────────────────┼───────────────┘
         │                │                   │
         v                v                   v
┌─────────────────────────────────────────────────────────────┐
│                   业务逻辑层 (Go API)                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  RESTful API Gateway (Gin)  +  JWT Auth  +  RBAC     │  │
│  │  用户管理 │ 产品溯源 │ NFC验证 │ 数据分析             │  │
│  └──────────────────────┬───────────────────────────────┘  │
└─────────────────────────┼──────────────────────────────────┘
                          │
         ┌────────────────┼────────────────┐
         v                v                v
┌─────────────────────────────────────────────────────────────┐
│              智能合约层 (Smart Contracts)                     │
│  ┌──────────────────┐ ┌────────────────┐ ┌──────────────┐ │
│  │ ProductRegistry  │ │TraceabilityLog │ │AccessControl │ │
│  └──────────────────┘ └────────────────┘ └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
                          │
                          v
┌─────────────────────────────────────────────────────────────┐
│            区块链数据层 (XuperChain 7-Node PBFT)              │
│  Node1(品牌A) Node2(品牌B) Node3(物流A) Node4(物流B)         │
│  Node5(监管)  Node6(协会)  Node7(运维)                      │
│  密码学: SM2/SM3  │  共识: PBFT  │  存储: LevelDB           │
└─────────────────────────────────────────────────────────────┘
```

---

## 📁 项目结构

```
paddle-trace/
├── contracts/                     # 智能合约 (Go)
│   ├── access/
│   │   └── access_control.go      # 访问控制合约 (RBAC)
│   └── product/
│       ├── product_registry.go    # 产品注册合约
│       └── traceability_log.go    # 溯源记录合约
├── api/                           # Go API 业务服务
│   ├── cmd/
│   │   └── main.go                # 服务入口
│   ├── config/
│   │   └── config.go              # 环境配置管理
│   ├── internal/
│   │   ├── handler/
│   │   │   └── handlers.go        # HTTP请求处理器
│   │   ├── middleware/
│   │   │   └── auth.go            # JWT/RBAC/CORS中间件
│   │   ├── model/
│   │   │   └── models.go          # 数据结构定义
│   │   └── service/
│   │       ├── blockchain_service.go  # 区块链交互服务
│   │       └── nfc_service.go         # NFC安全验证服务
│   ├── Dockerfile                 # API服务Docker镜像
│   ├── go.mod
│   └── go.sum
├── frontend/
│   └── consumer/
│       └── index.html             # 消费者扫码验证页面
├── config/
│   └── nginx.conf                 # Nginx前端代理配置
├── scripts/
│   └── init-db.sql                # PostgreSQL初始化脚本
├── test/
│   ├── contract_test.go           # 智能合约单元测试
│   └── api_integration_test.go    # API集成测试
├── docker-compose.yml             # Docker Compose编排
└── README.md
```

---

## 🚀 快速启动

### 前置要求

- Docker 24.0+ & Docker Compose 2.20+
- Go 1.21+ (仅本地开发)
- NFC读写器 (可选，用于标签初始化)

### 一键启动

```bash
# 克隆项目
git clone https://github.com/fansicheng/paddle-trace.git
cd paddle-trace

# 启动所有服务（7节点区块链 + DB + API + 前端）
docker compose up -d

# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f api
```

启动后访问：
- 🔍 **消费者扫码验证页面**: http://localhost:3000/consumer
- 📊 **管理后台**: http://localhost:3000/admin
- 🏥 **API健康检查**: http://localhost:8080/api/v1/health

### 本地开发

```bash
# 仅启动基础设施（区块链 + 数据库 + 缓存）
docker compose up -d xchain-node1 postgres redis

# 本地运行API服务
cd api
go run cmd/main.go

# 运行测试
cd test
go test -v ./...
```

---

## 📡 API接口

### 公开接口（无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 服务健康检查 |
| GET | `/api/v1/products/:id` | 查询产品溯源信息 |
| POST | `/api/v1/products/verify-nfc` | NFC扫码真伪验证 |

### 认证接口（需要JWT Token + 角色权限）

| 方法 | 路径 | 所需角色 |
|------|------|---------|
| POST | `/api/v1/products` | manufacturer / admin |
| POST | `/api/v1/products/:id/transfer` | manufacturer / logistics / distributor |
| POST | `/api/v1/products/:id/trace` | manufacturer / logistics / distributor |

### NFC验证请求示例

```json
POST /api/v1/products/verify-nfc
{
    "tag_uid":    "04A2B3C4D5E6F7",
    "sun_code":   "abcdef1234567890abcdef1234567890",
    "counter":    42,
    "product_id": "pingpong101"
}
```

### NFC验证响应示例

```json
{
    "code": 200,
    "message": "verification complete",
    "data": {
        "authentic": true,
        "product_id": "pingpong101",
        "brand": "Butterfly",
        "model": "VISCARIA FL",
        "nfc_verified": true,
        "chain_verified": true,
        "message": "✅ 验证通过：该球拍为正品"
    }
}
```

---

## 🛡️ 安全特性

| 安全属性 | 实现机制 | 验收指标 |
|---------|---------|---------|
| 数据完整性 | 区块链链式存储 + SM3哈希 | 篡改成功率 0% |
| 身份认证 | SM2数字签名 + PKI证书 | 未授权操作率 0% |
| 访问控制 | 智能合约RBAC + JWT鉴权 | 权限越权测试通过 |
| 不可否认性 | 每笔交易附带数字签名 | 签名覆盖率 100% |
| NFC安全 | NTAG 424 DNA SUN动态认证 | 克隆/重放攻击防御 |
| 可用性 | PBFT共识 + Redis缓存 | P99延迟 ≤2秒 |

---

## 🧪 测试

```bash
# 智能合约单元测试
cd test && go test -v -run TestContract

# API集成测试
cd test && go test -v -run TestAPI

# 安全测试
cd test && go test -v -run TestSecurity

# 性能基准测试
cd test && go test -bench=. -benchmem
```

---

## 💰 经济成本

| 阶段 | 成本项 | 金额 |
|------|--------|------|
| 开发测试 | NFC标签 + 读写器 | 约200元 |
| 开发测试 | 区块链/软件工具 | 0元（开源） |
| 商用预估 | 月度运营（云+维护） | 3,000-5,000元/月 |

---

## 📚 相关文档

- [期中开题与需求分析报告](../基于区块链的乒乓球拍防伪溯源系统_开题与需求分析报告.md)
- [期末结题报告](../基于区块链的乒乓球拍防伪溯源系统_期末结题报告.md)

---

## 📄 License

MIT License © 2025 樊思成

---

*本项目为 2025年《信息安全工程概论》课程期末作业，所有代码均为课程项目原创成果。*
