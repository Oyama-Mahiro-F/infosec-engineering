# 🏓 基于区块链的乒乓球拍防伪溯源系统

> **Paddle Traceability System** — 2026年《信息安全工程概论》课程项目  
> 基于百度超级链（XuperChain）+ Go + NFC 的防伪溯源原型系统

---

## 项目简介

本项目以乒乓球拍（Butterfly品牌）为切入点，基于信息安全工程方法论，设计并实现了一套基于**百度超级链（XuperChain）**的乒乓球拍防伪溯源原型系统。系统通过Go语言智能合约实现产品全生命周期的信息上链存证，确保从原材料采购、生产加工、仓储物流到终端零售的每一个环节数据**真实可信、不可篡改、全程可追溯**。

### 核心特性

- 🔗 **区块链存证** — 基于XuperChain 7节点PBFT联盟链，哈希链式存储，数据不可篡改
- 📱 **NFC芯片验证** — NTAG 424 DNA SUN动态认证，防标签克隆与重放攻击
- 🛡️ **国密算法** — SM2签名 + SM3哈希，符合国密合规要求
- 🎯 **角色权限控制** — 基于RBAC的智能合约访问控制
- ⚡ **原型可运行** — Python内嵌区块链引擎 + Flask API，一键启动
- 🧪 **自动化测试** — PowerShell端到端测试脚本，20检查点全部通过

---

## 快速体验

### 启动后端

```bash
cd paddle-trace
pip install flask flask-cors
python server.py
```

### 打开前端

浏览器访问 `http://127.0.0.1:5050/`，点击"模拟NFC扫码"或输入产品ID：
- `pingpong101` — ✅ 正品（Butterfly VISCARIA FL，7条溯源记录，含二手交易）
- `pingpong102` — ✅ 正品（Butterfly ZHANG JIKE ALC）
- `pingpong104` — ❌ 假货（未在区块链注册）

### 运行测试

```bash
# PowerShell 端到端测试（需要先启动 server.py）
powershell -ExecutionPolicy Bypass -File test/e2e_test.ps1
```

---

## 项目结构

```
paddle-trace/
├── contracts/                          # Go智能合约（XuperChain WASM部署）
│   ├── access/access_control.go        #   AccessControl — RBAC权限控制
│   └── product/
│       ├── product_registry.go         #   ProductRegistry — 产品注册+转移+验证
│       └── traceability_log.go         #   TraceabilityLog — 溯源记录+历史查询
├── api/                                # Go API服务（Gin框架，设计层）
│   ├── cmd/main.go                     #   服务入口
│   ├── config/config.go                #   配置管理
│   └── internal/
│       ├── handler/handlers.go         #   HTTP请求处理器
│       ├── middleware/auth.go          #   JWT/RBAC/CORS中间件
│       ├── model/models.go             #   数据结构定义
│       └── service/
│           ├── blockchain_service.go   #   XuperChain SDK交互
│           └── nfc_service.go          #   NFC SUN认证服务
├── server.py                           # Python原型后端（可直接运行）
├── frontend/consumer/index.html        # 消费者扫码验证页面
├── test/
│   ├── contract_test.go                #   Go合约单元测试
│   ├── api_integration_test.go         #   API集成测试
│   └── e2e_test.ps1                    #   PowerShell端到端测试（20检查点）
├── scripts/
│   ├── deploy_demo.sh                  #   XuperChain部署演示（10步）
│   ├── contract_demo.sh                #   合约调用演示（6步业务流程）
│   └── init-db.sql                     #   PostgreSQL初始化脚本
├── config/nginx.conf                   # Nginx配置
├── docker-compose.yml                  # 7节点联盟链Docker编排
├── Makefile
└── README.md
```

---

## 系统架构

```
前端展示层:  消费者Web页面  +  管理后台(React)  +  监管Dashboard(ECharts)
     │
业务逻辑层:  RESTful API (Python Flask)  +  JWT Auth  +  RBAC
     │
智能合约层:  ProductRegistry  |  TraceabilityLog  |  AccessControl
     │
区块链层:    XuperChain 7-Node PBFT  |  SM2/SM3  |  LevelDB
```

---

## 部署演示

两个终端脚本，直接运行即可截图用于报告：

```bash
# XuperChain 部署流程（10步：编译→配置→启动→部署合约→调用测试）
bash scripts/deploy_demo.sh

# 合约调用业务流程（6步：创建账户→上链→采购→销售→二手→鉴别）
bash scripts/contract_demo.sh
```

---

## API接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 健康检查 + 区块链状态 |
| GET | `/api/v1/products/:id` | 产品溯源查询 |
| POST | `/api/v1/products` | 产品注册上链（需制造商角色） |
| POST | `/api/v1/products/:id/transfer` | 所有权转移 |
| POST | `/api/v1/products/:id/trace` | 追加溯源记录 |
| POST | `/api/v1/products/verify-nfc` | NFC真伪验证 |
| GET | `/api/v1/admin/chain` | 区块链状态查看 |
| GET | `/api/v1/admin/reset` | 重置演示数据 |

---

## 端到端测试结果

```
=======================================================
  TOTAL: 20  PASS: 20  FAIL: 0
  >>> ALL TESTS PASSED <<<
=======================================================
```

| 测试 | 内容 | 结果 |
|------|------|------|
| Test 1 | 健康检查与链完整性 | height=17, integrity=valid |
| Test 2 | 正品溯源查询 | 7条记录, owner=0xConsumer02 |
| Test 3 | 溯源哈希链验证 | 7条prev_record链式链接完整 |
| Test 4 | NFC真伪验证 | authentic+nfc+chain三验证通过 |
| Test 5 | 假货检测 | pingpong104 → 404 |
| Test 6 | 产品注册上链 | 201, 高度17→18 |
| Test 7 | NFC重放攻击防御 | 第1次通过，第2次拒绝 |
| Test 8 | 权限越权拦截 | 403 permission denied |

---

## 经济成本

| 阶段 | 成本项 | 金额 |
|------|--------|------|
| 开发测试 | NFC标签 + 读写器 | 约200元 |
| 开发测试 | 区块链/软件工具 | 0元（开源） |
| 商用预估 | 月度运营 | 3,000-5,000元/月 |

---

## 相关文档

- [期中开题与需求分析报告](../基于区块链的乒乓球拍防伪溯源系统_开题与需求分析报告.md)
- [期末结题报告（Markdown）](../基于区块链的乒乓球拍防伪溯源系统_期末结题报告.md)
- [期末作业（Docx）](../2025年《信息安全工程概论》-期末作业.docx)

---

MIT License © 2026 樊思成
