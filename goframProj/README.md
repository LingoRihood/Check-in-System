# checkin-h5-app（前后端同级）

本目录下包含：
- `frontend/`：基于 prototype.html 开发的 Vue3 + TS + Vite H5 前端

> 说明：后端目前只实现了 **注册 / 登录 / 刷新 token / 获取当前用户信息**。  
> “签到 / 积分 / 补签 / 奖励”等业务接口在后端代码里尚未实现，所以前端为了保证能跑起来，先用 **localStorage** 模拟积分与签到规则；  
> 一旦后端补齐对应接口，前端只需要把 points store 的本地逻辑替换为真实 API 调用即可。

---

## 1) 后端启动

### 1.1 准备 MySQL

后端默认配置（见 `backend/manifest/config/config.yaml`）：
- host: `127.0.0.1:3306`
- db: `demo`
- user: `gf`
- password: `GfPass!123`

在 MySQL 里创建库并建表：

```sql
CREATE DATABASE IF NOT EXISTS demo DEFAULT CHARACTER SET utf8mb4;
USE demo;
SOURCE backend/manifest/sql/userinfo.sql;
```

### 1.2 启动后端

```bash
cd backend
go mod tidy
go run main.go
```

后端地址默认：`http://localhost:8000`

---

## 2) 前端启动

```bash
cd frontend
npm i
npm run dev
```

前端开发服务器：`http://localhost:5173`

前端默认通过 Vite 代理访问后端：`/api/v1/*` -> `http://localhost:8000/api/v1/*`

---

## 3) 已对接的后端接口（能跑通）

- `POST /api/v1/users` 注册
- `POST /api/v1/auth/login` 登录
- `POST /api/v1/auth/refresh` 刷新 token
- `GET /api/v1/users/me` 获取当前用户信息（Authorization: Bearer）

---

## 4) 已实现的业务规则（前端本地模拟）

- 每天签到：+1 积分
- 当月连续签到：3 天 +5，7 天 +10，15 天 +20
- 当月满签：+100（演示：签满即发）
- 当月补签：每次消耗 100 积分，每月最多 3 次
- 首页展示：总积分 / 本月积分 / 连签天数 / 剩余补签次数 / 当月签到日历
- 积分详情页：按月查看明细记录，支持筛选（全部/获取/消耗）

---

## 5) 后端若要补齐接口（建议）

建议新增：
- `GET /points/summary`：总积分、本月积分、连签、剩余补签
- `GET /points/records?month=YYYY-MM`
- `POST /checkin`：今日签到
- `POST /checkin/makeup`：补签 { date: 'YYYY-MM-DD' }

前端结构已经为 API 对接留好了 axios 与 token 刷新能力。
