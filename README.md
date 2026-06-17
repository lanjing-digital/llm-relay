# LLM Relay

大模型中转站，对外暴露 OpenAI 兼容接口，外部调用方统一使用 `deepseek-chat` 作为模型标识，内部根据管理后台配置将请求转发到真正的目标模型和 API 地址。

前端管理页面使用 React + Antd 构建，构建产物内嵌到 Go 二进制中，单文件部署。

## 功能特性

- OpenAI 兼容的 `/v1/chat/completions` 转发接口，支持流式（SSE）和非流式
- 管理后台可视化配置模型映射（增删改查 + 连通性测试）
- SQLite 持久化存储，零外部依赖
- 前端构建产物内嵌到 Go 二进制，单文件部署
- 支持自定义上游 Base URL（兼容 `/v1`、`/v3` 等版本路径）

## 技术栈

| 层 | 技术 |
|------|------|
| 后端 | Go 1.25+、Gin、GORM、SQLite |
| 前端 | React 19、Antd 6、Vite、TypeScript |
| 部署 | 单二进制文件（embed 内嵌前端） |

## 项目结构

```
llm_relay/
├── cmd/relay/main.go              # 后端入口
├── internal/
│   ├── db/db.go                   # SQLite 初始化 + 自动迁移
│   ├── model/config.go            # Config 数据模型
│   ├── repository/config_repo.go  # Config CRUD 数据访问层
│   └── handler/
│       ├── config_handler.go      # 配置管理 API + 连通性测试
│       └── proxy_handler.go       # OpenAI 兼容转发 + URL 拼接 + SSE 透传
├── web/                           # React 前端
│   ├── src/
│   │   ├── App.tsx                # 管理页面主组件
│   │   ├── components/
│   │   │   └── config-form-modal.tsx
│   │   ├── api/
│   │   │   ├── client.ts          # axios 实例
│   │   │   └── configs.ts         # 配置 API 封装
│   │   └── types/config.ts        # 类型定义
│   ├── vite.config.ts
│   └── package.json
├── web_embed.go                   # embed 内嵌 web/dist
├── go.mod
└── data.db                        # SQLite 数据库（运行时自动生成）
```

## 环境要求

- Go 1.25+
- Node.js 18+
- npm 9+（或 pnpm / yarn）

## 快速开始

### 1. 克隆项目

```bash
git clone <your-repo-url> llm_relay
cd llm_relay
```

### 2. 构建前端

```bash
cd web
npm install
npm run build
cd ..
```

构建产物输出到 `web/dist/`，会被 Go 通过 `embed` 内嵌到二进制中。

### 3. 构建后端

```bash
go build -o llm-relay ./cmd/relay
```

### 4. 运行

```bash
./llm-relay
```

默认监听 `8080` 端口，数据库文件 `data.db` 会在当前目录自动创建。

### 5. 访问管理页面

浏览器打开：http://localhost:8080/admin

## 开发模式

开发时前后端分开启动，前端通过 Vite 代理转发 API 请求到后端。

**终端 1 - 启动后端：**

```bash
go run ./cmd/relay
```

**终端 2 - 启动前端：**

```bash
cd web
npm run dev
```

前端开发服务器地址：http://localhost:5173/admin/

Vite 已配置代理，`/api` 和 `/v1` 请求会自动转发到 `http://localhost:8080`。

## 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `LLM_RELAY_PORT` | 服务监听端口 | `8080` |

示例：

```bash
LLM_RELAY_PORT=9090 ./llm-relay
```

### 配置管理 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/configs` | 列出所有配置 |
| POST | `/api/configs` | 创建配置 |
| PUT | `/api/configs/:id` | 更新配置 |
| DELETE | `/api/configs/:id` | 删除配置 |
| POST | `/api/configs/:id/test` | 测试上游连通性 |

### Config 数据模型

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | integer | 自增主键 |
| `name` | string | 配置名称（自定义） |
| `external_model` | string | 对外暴露的模型标识（如 `deepseek-chat`） |
| `target_model` | string | 转发到上游时的真实模型名 |
| `target_base_url` | string | 上游 API 地址 |
| `target_api_key` | string | 上游 API 密钥 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

### target_base_url 拼接规则

| 输入 Base URL | 实际转发地址 |
|------|------|
| `https://api.openai.com/v1` | `https://api.openai.com/v1/chat/completions` |
| `https://ark.cn-beijing.volces.com/api/v3` | `https://ark.cn-beijing.volces.com/api/v3/chat/completions` |
| `https://api.example.com` | `https://api.example.com/v1/chat/completions` |
| `https://api.example.com/v1/chat/completions` | 原样使用 |

## 使用示例

### 1. 通过 API 创建配置

```bash
curl -X POST http://localhost:8080/api/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "豆包线路",
    "external_model": "deepseek-chat",
    "target_model": "doubao-seed-2.0-lite",
    "target_base_url": "https://ark.cn-beijing.volces.com/api/v3",
    "target_api_key": "your-api-key"
  }'
```

### 2. 测试连通性

```bash
curl -X POST http://localhost:8080/api/configs/1/test
```

### 3. 调用转发接口（非流式）

外部调用方使用 `deepseek-chat` 作为 `model`，服务自动替换为真实模型并转发：

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-chat",
    "messages": [
      {
        "role": "user",
        "content": "你好，简单介绍一下你自己。"
      }
    ],
    "stream": false
  }'
```

### 4. 调用转发接口（流式）

```bash
curl -N http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-chat",
    "messages": [
      {
        "role": "user",
        "content": "请用三句话介绍 Golang。"
      }
    ],
    "stream": true
  }'
```

### 5. 查看当前配置

```bash
curl http://localhost:8080/api/configs
```

## 与 OpenAI SDK 兼容性

本项目完全兼容 OpenAI SDK，只需将 `base_url` 指向本服务即可：

**Python (openai)：**

```python
from openai import OpenAI

client = OpenAI(
    api_key="any-string",
    base_url="http://localhost:8080/v1",
)

response = client.chat.completions.create(
    model="deepseek-chat",
    messages=[{"role": "user", "content": "Hello!"}],
)
print(response.choices[0].message.content)
```

**Node.js (openai)：**

```javascript
import OpenAI from 'openai'

const client = new OpenAI({
  apiKey: 'any-string',
  baseURL: 'http://localhost:8080/v1',
})

const response = await client.chat.completions.create({
  model: 'deepseek-chat',
  messages: [{ role: 'user', content: 'Hello!' }],
})
console.log(response.choices[0].message.content)
```

## 部署

### 单文件部署

```bash
# 构建前端
cd web && npm install && npm run build && cd ..

# 构建后端（内嵌前端）
CGO_ENABLED=1 go build -o llm-relay ./cmd/relay

# 运行
./llm-relay
```

> SQLite 驱动依赖 CGO，构建时需要 `CGO_ENABLED=1` 且本机有 C 编译器（gcc）。

### 交叉编译

由于依赖 CGO（SQLite），交叉编译需要对应平台的 C 编译器。例如在 macOS 上编译 Linux 版本：

```bash
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc go build -o llm-relay ./cmd/relay
```

## 常见问题

### 启动时端口被占用

```bash
# 查看占用进程
lsof -nP -iTCP:8080 -sTCP:LISTEN

# 结束进程后重试
kill <PID>
```

或使用自定义端口：

```bash
LLM_RELAY_PORT=9090 ./llm-relay
```

### 转发返回 404

检查管理后台中该配置的 `target_base_url` 是否正确。Base URL 应到版本号为止（如 `/v1`、`/v3`），不要包含 `/chat/completions` 后缀（也可以包含，系统会原样使用）。

### CGO 相关编译错误

确保已安装 C 编译器：

- macOS：`xcode-select --install`
- Ubuntu/Debian：`apt install gcc`
- CentOS：`yum install gcc`
