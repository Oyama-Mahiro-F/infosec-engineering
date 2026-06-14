# 🏓 信息安全工程概论 — 课程资料

> 2025年《信息安全工程概论》 · 北京邮电大学  
> 指导教师：昌硕 ｜ 学生：樊思成（2023211544）

---

## 📂 仓库内容

### 期末大作业：基于区块链的乒乓球拍防伪溯源系统

完整的项目代码与报告，详见 [paddle-trace/](paddle-trace/)

```
paddle-trace/
├── contracts/          智能合约（Go）
│   ├── access/         AccessControl — RBAC权限控制
│   └── product/        ProductRegistry + TraceabilityLog
├── api/                RESTful API服务（Gin框架）
├── frontend/           消费者扫码验证页面
├── test/               30+测试用例
├── config/             Nginx配置
├── scripts/            PostgreSQL初始化
└── docker-compose.yml  7节点XuperChain + DB + Redis
```

### 课程报告

| 文档 | 说明 |
|------|------|
| [期中开题与需求分析报告](基于区块链的乒乓球拍防伪溯源系统_开题与需求分析报告.md) | 项目背景、需求分析、架构设计、技术选型 |
| [期末结题报告](基于区块链的乒乓球拍防伪溯源系统_期末结题报告.md) | 完整开发过程、代码实现、测试结果、最终成果 |

---

## 🚀 快速体验

浏览器直接打开前端演示页面（无需任何环境）：

```
paddle-trace/frontend/consumer/index.html
```

输入产品ID `pingpong101`（正品）或 `pingpong104`（假货）查看验证效果。

完整运行方式见 [paddle-trace/README.md](paddle-trace/README.md)
