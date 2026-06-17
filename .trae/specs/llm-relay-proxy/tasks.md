# Tasks

- [x] Task 1: Go 后端项目初始化
  - [x] 初始化 Go module 并引入依赖（gin, gorm, sqlite driver）
  - [x] 创建项目目录结构（cmd/, internal/, web/）
  - [x] 创建 main.go 入口文件，启动 Gin 服务

- [x] Task 2: 数据库层实现
  - [x] 定义 Config 数据模型（GORM Model）
  - [x] 实现 SQLite 数据库初始化和自动迁移
  - [x] 实现 Config 的 CRUD 数据访问层（repository）

- [x] Task 3: 模型映射配置管理 API
  - [x] 实现 GET /api/configs（列表）
  - [x] 实现 POST /api/configs（创建）
  - [x] 实现 PUT /api/configs/:id（更新）
  - [x] 实现 DELETE /api/configs/:id（删除）
  - [x] 实现 POST /api/configs/:id/test（连通性测试）

- [x] Task 4: OpenAI 兼容转发接口
  - [x] 实现 POST /v1/chat/completions 请求解析
  - [x] 实现根据 external_model 查找配置逻辑
  - [x] 实现请求转发（替换 model 字段，设置 Authorization header）
  - [x] 实现流式响应（SSE）透传
  - [x] 实现非流式响应透传

- [x] Task 5: Go 后端 - 静态资源内嵌与服务
  - [x] 使用 embed 内嵌 web/dist 目录
  - [x] 配置 Gin 静态文件路由，/admin 路径返回前端 SPA
  - [x] 配置 SPA fallback（Vue/React router 的 history mode）

- [x] Task 6: React 前端项目初始化
  - [x] 使用 Vite 创建 React + TypeScript 项目（web/ 目录）
  - [x] 安装 antd 和 axios 依赖
  - [x] 配置 Vite 代理（开发时转发 /api 到后端）
  - [x] 配置 Vite build 输出到 web/dist/

- [x] Task 7: 前端管理页面实现
  - [x] 实现 Config 列表表格组件（antd Table）
  - [x] 实现创建/编辑 Config 的 Modal 表单（antd Form + Modal）
  - [x] 实现删除 Config 的确认弹窗（antd Popconfirm）
  - [x] 实现连通性测试功能（调用后端测试接口并展示结果）
  - [x] 实现 API 请求封装（axios 实例）

- [x] Task 8: 集成验证
  - [x] 构建前端（npm run build）
  - [x] 构建后端（go build），确认内嵌静态资源生效
  - [x] 端到端验证：启动服务 → 访问管理页面 → 创建配置 → curl 调用 /v1/chat/completions

# Task Dependencies
- Task 2 依赖 Task 1
- Task 3 依赖 Task 2
- Task 4 依赖 Task 2
- Task 5 依赖 Task 6（需要前端构建产物目录存在）
- Task 6 独立，可与 Task 1-4 并行
- Task 7 依赖 Task 6
- Task 8 依赖所有前置任务
