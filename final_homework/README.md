# 智能招聘系统

基于 **Gin + gRPC 双层微服务架构** 的智能招聘管理平台，包含 HR 管理端与候选人端两个独立前端。

## 项目结构

```
final_homework/
├── hr-frontend/         # HR 管理端 (Vue3)
├── user-frontend/       # 候选人端 (Vue3)
├── web-gin-service/     # HTTP 网关 (Gin)
├── logic-grpc-service/  # 核心业务 (gRPC)
├── api.md               # 接口文档
├── db.md                # 数据库设计
└── README.md
```

## 技术栈

- 后端：Go、Gin、gRPC、GORM、MySQL、JWT
- 存储：阿里云 OSS 私有 Bucket + 预签名 URL
- AI：CloudWeGo Eino + DashScope（通义千问）
- 前端：Vue 3、Vite、Element Plus、Pinia、Axios

## 环境要求

- Go 1.21+
- Node.js 18+
- MySQL 8+
- 阿里云 OSS 账号（私有 Bucket）
- DashScope API Key

## 配置说明

**敏感配置（数据库、JWT、OSS、API Key）仅通过本地 `.env` 注入，不会也不应提交到 Git。**

### 1. MySQL

```sql
CREATE DATABASE IF NOT EXISTS smart_recruit DEFAULT CHARSET utf8mb4;
```

### 2. Logic 服务

```bash
cd logic-grpc-service
cp .env.example .env
# 编辑 .env，填写 MYSQL_DSN、JWT_SECRET、OSS_*、DASHSCOPE_API_KEY
```

`.env` 会被自动加载。非敏感默认值（端口等）在 `config/config.example.yaml`，可提交 Git。

### 3. Web 网关

```bash
cd web-gin-service
cp .env.example .env
# 编辑 .env，填写 JWT_SECRET（需与 logic 服务相同）
```

### 4. 环境变量一览

**logic-grpc-service/.env**

| 变量 | 必填 | 说明 |
|------|------|------|
| `MYSQL_DSN` | 是 | MySQL 连接串 |
| `JWT_SECRET` | 是 | JWT 签名密钥 |
| `OSS_ENDPOINT` | 是 | OSS Endpoint |
| `OSS_ACCESS_KEY_ID` | 是 | OSS AccessKey ID |
| `OSS_ACCESS_KEY_SECRET` | 是 | OSS AccessKey Secret |
| `OSS_BUCKET_NAME` | 是 | 私有 Bucket 名称 |
| `DASHSCOPE_API_KEY` | 是 | 通义千问 API Key |

**web-gin-service/.env**

| 变量 | 必填 | 说明 |
|------|------|------|
| `JWT_SECRET` | 是 | 与 logic 服务保持一致 |

### 5. OSS / DashScope 申请

- OSS：创建私有 Bucket，关闭公共读，RAM 用户授权读写
- DashScope：在 [阿里云百炼](https://bailian.console.aliyun.com/) 创建 API Key

> `.env` 和 `config/config.yaml` 已在 `.gitignore` 中排除，请勿提交真实密钥。

## 启动顺序

**必须按以下顺序启动：**

### 1. 启动 MySQL

确保 `smart_recruit` 数据库已创建。

### 2. 启动 Logic gRPC 服务

```bash
cd logic-grpc-service
go run cmd/server/main.go
```

默认端口：`50051`

### 3. 启动 Web Gin 网关

```bash
cd web-gin-service
go run cmd/server/main.go
```

默认端口：`8080`

### 4. 启动 HR 前端

```bash
cd hr-frontend
npm install
npm run dev
```

访问：http://localhost:5173

### 5. 启动候选人前端

```bash
cd user-frontend
npm install
npm run dev
```

访问：http://localhost:5174

## 功能验证

1. **候选人端**：游客浏览岗位 → 注册/登录 → 完善档案 → OSS 上传简历 → 一键投递
2. **HR 端**：注册 HR 账号 → 发布岗位 → 查看候选人台账 → 下载简历
3. **AI 对话**：HR 端提问「我的岗位总共有多少投递？」→ 基于 MySQL 真实数据回答 → 刷新页面历史保留

## 架构说明

```
前端 → web-gin-service (JWT/CORS/参数校验) → logic-grpc-service (业务逻辑) → MySQL/OSS/DashScope
```

- Web 层不含业务逻辑，全部通过 gRPC 调用
- 简历文件直传 OSS，服务器不落盘
- AI 使用 Eino ChatModel 组件，禁止裸 HTTP 调 LLM
- AI 基于 MySQL 统计查询，不使用向量库/RAG

## 项目亮点

- 严格双层 gRPC 微服务分层
- JWT 统一鉴权 + 角色权限隔离
- OSS 预签名直传/直读，私有 Bucket 安全存储
- Eino 框架 AI 对话，意图识别 + MySQL 数据驱动
- 双前端独立部署，职责清晰
