# API 接口文档

Base URL: `http://localhost:8080/api`

统一响应格式：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

认证方式：除公开接口外，请求头需携带 `Authorization: Bearer <token>`

---

## 1. 认证模块

### POST /auth/register

注册（HR 或候选人）

请求体：
```json
{
  "username": "hr01",
  "password": "123456",
  "role": "hr"
}
```

`role` 取值：`hr` | `candidate`

### POST /auth/login

登录

请求体：
```json
{
  "username": "hr01",
  "password": "123456"
}
```

响应 data：
```json
{
  "token": "jwt-token",
  "user_id": 1,
  "username": "hr01",
  "role": "hr"
}
```

---

## 2. 岗位模块（公开）

### GET /jobs

游客可访问，分页获取公开岗位列表

Query：`page`, `page_size`

### GET /jobs/:id

获取岗位详情

---

## 3. HR 岗位管理

权限：HR + JWT

### GET /hr/jobs

获取当前 HR 发布的岗位

### POST /hr/jobs

创建岗位

请求体：
```json
{
  "title": "Go 后端工程师",
  "description": "负责微服务开发",
  "salary": "15k-25k",
  "location": "杭州"
}
```

### PUT /hr/jobs/:id

更新岗位（仅能修改自己的）

### DELETE /hr/jobs/:id

删除岗位（仅能删除自己的）

---

## 4. 候选人档案

权限：Candidate + JWT

### GET /user/profile

获取结构化档案

### PUT /user/profile

更新档案

请求体：
```json
{
  "name": "张三",
  "phone": "13800000000",
  "education": "本科",
  "school": "浙江大学",
  "experience": "3年后端开发经验",
  "skills": "Go, MySQL, gRPC"
}
```

---

## 5. 简历 OSS

### GET /user/resume/upload-url

权限：Candidate + JWT

Query：`filename=resume.pdf`

响应 data：
```json
{
  "upload_url": "https://...",
  "oss_key": "resumes/1/xxx.pdf",
  "expire_seconds": 900
}
```

前端使用 `PUT upload_url` 直传 OSS，然后调用确认接口。

### POST /user/profile/resume

确认简历上传

请求体：
```json
{
  "oss_key": "resumes/1/xxx.pdf"
}
```

### GET /hr/resume/download-url

权限：HR + JWT

Query：`candidate_id`, `oss_key`

---

## 6. 投递

### POST /user/applications

权限：Candidate + JWT

请求体：
```json
{
  "job_id": 1
}
```

错误码说明：
- `412`：档案未完善或未上传简历
- `409`：重复投递

---

## 7. HR 候选人台账

### GET /hr/candidates

权限：HR + JWT

Query：`page`, `page_size`, `job_id`（可选）

---

## 8. AI 对话

权限：HR + JWT

### POST /hr/ai/chat

```json
{
  "message": "我的岗位总共有多少投递？"
}
```

### GET /hr/ai/history

获取当前 HR 的全部对话历史

---

## 错误码

| HTTP | 说明 |
|------|------|
| 400 | 参数错误 |
| 401 | 未登录或 Token 无效 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 409 | 冲突（用户名已存在/重复投递） |
| 412 | 前置条件不满足（档案不完整） |
| 500 | 服务器内部错误 |
