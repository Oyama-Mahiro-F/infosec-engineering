# 信息安全工程概论 — 课程项目

> 2026年《信息安全工程概论》· 樊思成（2023211544）· 指导教师：昌硕

---

## 项目：基于区块链的乒乓球拍防伪溯源系统

以乒乓球拍（Butterfly 品牌）为切入点，设计并实现基于**百度超级链（XuperChain）**+ **NFC 芯片认证**的防伪溯源系统。核心思路是用区块链的不可篡改特性解决传统中心化防伪方案的数据信任问题。

---

## 仓库结构

```
├── README.md                                    ← 本文件
├── 要求.md                                      ← 课程作业要求原文
├── 基于区块链的乒乓球拍防伪溯源系统_开题与需求分析报告.md   ← 期中报告（需求分析+架构设计）
├── 基于区块链的乒乓球拍防伪溯源系统_期末结题报告.md       ← 期末报告（开发实现+测试结果）
├── 2025年《信息安全工程概论》-期末作业.docx            ← 期末作业 docx 模板
│
└── paddle-trace/                                ← 项目代码
    ├── README.md                                ←   项目详细说明
    ├── server.py                                ←   Python 原型后端（可直接运行）
    ├── docker-compose.yml                       ←   7节点XuperChain联盟链编排
    ├── Makefile                                 ←   构建/测试/清理快捷命令
    │
    ├── contracts/                               ←   Go 智能合约（可编译为WASM部署到XuperChain）
    │   ├── access/access_control.go              ←     AccessControl — RBAC 权限控制（6种角色）
    │   └── product/
    │       ├── product_registry.go               ←     ProductRegistry — 产品注册/转移/真伪验证
    │       └── traceability_log.go               ←     TraceabilityLog — 溯源记录/链式哈希/完整性验证
    │
    ├── api/                                     ←   Go API 服务（Gin 框架）
    │   ├── cmd/main.go                          ←     服务入口
    │   ├── config/config.go                     ←     配置管理
    │   ├── Dockerfile                           ←     Docker 镜像构建
    │   └── internal/
    │       ├── handler/handlers.go              ←     HTTP 处理器
    │       ├── middleware/auth.go               ←     JWT / RBAC / CORS 中间件
    │       ├── model/models.go                  ←     数据结构定义
    │       └── service/
    │           ├── blockchain_service.go         ←     XuperChain SDK 交互
    │           └── nfc_service.go                ←     NFC SUN 动态认证
    │
    ├── frontend/consumer/index.html             ←   消费者扫码验证页面（纯前端，调用API）
    │
    ├── scripts/
    │   ├── deploy_demo.sh                       ←   XuperChain 部署演示（10步终端输出，用于截图）
    │   ├── contract_demo.sh                     ←   合约调用演示（6步业务流程）
    │   └── init-db.sql                          ←   PostgreSQL 初始化脚本
    │
    ├── test/
    │   ├── contract_test.go                     ←   Go 合约单元测试（30+用例）
    │   ├── api_integration_test.go              ←   API 集成测试
    │   └── e2e_test.ps1                         ←   PowerShell 端到端测试（20检查点）
    │
    └── config/nginx.conf                        ←   Nginx 反向代理配置
```

---

## 快速开始

```bash
cd paddle-trace

# 1. 安装依赖
pip install flask flask-cors

# 2. 启动后端（内嵌区块链引擎 + REST API）
python server.py

# 3. 浏览器打开前端
# frontend/consumer/index.html
# 输入产品ID：pingpong101（正品）/ pingpong104（假货）
```

---

## 两套实现的关系

| 层面 | Go 版本 | Python 版本 |
|------|---------|------------|
| 定位 | 设计层，可编译部署到真实 XuperChain | 原型层，可立即运行验证 |
| 智能合约 | 3个 Go 合约，编译为 WASM | 等效逻辑嵌入 server.py |
| 区块链 | XuperChain 7节点 PBFT 联盟链 | 内存哈希链，Block 类 + 链完整性验证 |
| API | Gin 框架 | Flask 框架 |
| 用途 | 展示真实技术栈和部署架构 | 展示完整业务流程和端到端测试 |

---

## 核心安全特性

- **数据完整性** — 区块链链式哈希存储，任何篡改导致整链验证失败
- **NFC 芯片认证** — NTAG 424 DNA SUN 动态认证码，每次扫码不同，防克隆/防重放
- **访问控制** — 智能合约 RBAC，6种角色（admin / manufacturer / logistics / distributor / auditor / consumer）
- **不可否认性** — 每次上链操作附带数字签名（SM2）
- **假货检测** — 未注册产品ID查询返回 404，NFC认证失败标记为 counterfeit

---

## 端到端测试

```bash
# 启动 server.py 后
powershell -ExecutionPolicy Bypass -File test/e2e_test.ps1
```

20 个检查点覆盖：链完整性、正品查询、溯源哈希链、NFC 验证、假货检测、产品注册、重放攻击防御、权限越权拦截。全部通过。

---

## 截图资源

两个 bash 脚本可直接运行并截图用于报告：

```bash
bash scripts/deploy_demo.sh     # XuperChain 部署全流程（编译→配置→7节点启动→部署合约→调用）
bash scripts/contract_demo.sh   # 合约调用业务流程（创建账户→上链→采购→销售→二手→鉴别）
```

---

**License:** MIT
