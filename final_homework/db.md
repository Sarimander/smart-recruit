# 数据库设计

数据库：MySQL 8  
库名：`smart_recruit`

## ER 关系

```
users (1) ----< jobs (HR发布)
users (1) ---- (1) candidate_profiles (候选人档案)
jobs + users(candidate) ----< applications (投递)
users(HR) ----< ai_chat_messages (AI对话)
```

---

## 1. users 用户表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AI | 主键 |
| username | VARCHAR(64) UNIQUE | 用户名 |
| password_hash | VARCHAR(255) | bcrypt 哈希 |
| role | VARCHAR(16) | hr / candidate |
| created_at | DATETIME | 创建时间 |

索引：`username` 唯一索引，`role` 普通索引

约束：一个用户只能有一种角色

---

## 2. jobs 岗位表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AI | 主键 |
| hr_id | BIGINT FK | 发布 HR 用户 ID |
| title | VARCHAR(128) | 岗位名称 |
| description | TEXT | 岗位描述 |
| salary | VARCHAR(64) | 薪资范围 |
| location | VARCHAR(128) | 工作地点 |
| status | VARCHAR(16) | active / inactive |
| created_at | DATETIME | 创建时间 |

索引：`hr_id`, `status`

外键：`hr_id -> users.id`

---

## 3. candidate_profiles 候选人档案表

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | BIGINT PK FK | 候选人用户 ID |
| name | VARCHAR(64) | 姓名 |
| phone | VARCHAR(32) | 手机 |
| education | VARCHAR(64) | 最高学历 |
| school | VARCHAR(128) | 学校 |
| experience | TEXT | 工作/项目经历 |
| skills | TEXT | 核心技能 |
| resume_oss_key | VARCHAR(512) | OSS 简历路径 |
| profile_complete | TINYINT(1) | 档案是否完整 |
| updated_at | DATETIME | 更新时间 |

外键：`user_id -> users.id`

`profile_complete` 规则：所有必填字段非空且 `resume_oss_key` 非空

---

## 4. applications 投递表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AI | 主键 |
| job_id | BIGINT FK | 岗位 ID |
| candidate_id | BIGINT FK | 候选人 ID |
| status | VARCHAR(16) | pending 等 |
| applied_at | DATETIME | 投递时间 |

索引：`job_id`, `candidate_id`  
唯一约束建议：`(job_id, candidate_id)` 防止重复投递

外键：`job_id -> jobs.id`, `candidate_id -> users.id`

---

## 5. ai_chat_messages AI 对话表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AI | 主键 |
| hr_id | BIGINT FK | HR 用户 ID |
| role | VARCHAR(16) | user / assistant |
| content | TEXT | 消息内容 |
| created_at | DATETIME | 创建时间 |

索引：`hr_id`, `created_at`

外键：`hr_id -> users.id`

---

## 初始化 SQL

```sql
CREATE DATABASE IF NOT EXISTS smart_recruit DEFAULT CHARSET utf8mb4;
USE smart_recruit;
```

服务启动时 GORM AutoMigrate 会自动建表。
