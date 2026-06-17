# LLM Relay Proxy Spec

## Why
搭建一个大模型中转站，对外暴露 OpenAI 兼容接口，外部调用方统一使用 `deepseek-chat` 作为模型标识，内部根据管理后台配置将请求转发到真正的目标模型和 API 地址。前端管理页面使用 React + Antd 实现。

## What Changes
- 新建 Go + Gin 后端服务，实现 OpenAI 兼容的 `/v1/chat/completions` 转发接口
- 新建 React + Antd 前端管理页面，管理模型映射配置
- 使用 SQLite 持久化存储模型映射配置
- 前端构建产物由后端内嵌静态资源提供服务

## Impact
- Affected specs: 全新项目，无已有影响
- Affected code: 全项目新建

## ADDED Requirements

### Requirement: OpenAI 兼容的 Chat Completions 接口
系统 SHALL 提供 `POST /v1/chat/completions` 接口，请求和响应格式完全兼容 OpenAI Chat Completions API。

#### Scenario: 使用 deepseek-chat 模型调用
- **WHEN** 调用方发送请求，其中 `model` 字段值为 `deepseek-chat`
- **THEN** 系统根据配置查找 `deepseek-chat` 对应的真实目标模型和 API 地址
- **AND** 系统将请求中的 `model` 字段替换为真实目标模型名
- **AND** 系统将请求转发到目标 API 地址
- **AND** 系统将目标 API 的响应流式原样返回（支持 SSE streaming）

#### Scenario: 请求模型未配置
- **WHEN** 调用方发送请求，其中 `model` 字段值未在系统中配置
- **THEN** 系统返回 400 错误，提示模型未配置

#### Scenario: 流式响应
- **WHEN** 调用方发送请求，其中 `stream` 字段为 `true`
- **THEN** 系统向目标 API 发起流式请求，并将 SSE 流逐块转发给调用方

### Requirement: 模型映射配置管理接口
系统 SHALL 提供 RESTful API 管理模型映射配置，包括增删改查操作。

#### Scenario: 列出所有配置
- **WHEN** 管理员请求 `GET /api/configs`
- **THEN** 系统返回所有模型映射配置的 JSON 数组

#### Scenario: 创建配置
- **WHEN** 管理员请求 `POST /api/configs` 提供配置信息（名称、上游模型、目标地址等）
- **THEN** 系统创建新的映射配置并返回创建结果

#### Scenario: 更新配置
- **WHEN** 管理员请求 `PUT /api/configs/:id` 提供更新后的配置信息
- **THEN** 系统更新对应配置并返回更新结果

#### Scenario: 删除配置
- **WHEN** 管理员请求 `DELETE /api/configs/:id`
- **THEN** 系统删除对应配置并返回成功

#### Scenario: 测试连通性
- **WHEN** 管理员请求 `POST /api/configs/:id/test`
- **THEN** 系统向目标 API 发送测试请求，返回连通性测试结果

### Requirement: 模型映射配置数据模型
每条模型映射配置 SHALL 包含以下字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | integer | 自增主键 |
| name | string | 配置名称（用户自定义） |
| external_model | string | 对外暴露的模型标识（如 deepseek-chat） |
| target_model | string | 转发到上游时的真实模型名 |
| target_base_url | string | 上游 API 地址 |
| target_api_key | string | 上游 API 密钥 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### Requirement: 前端管理页面
系统 SHALL 提供基于 React + Antd 构建的 SPA 管理页面，包含以下功能：

#### Scenario: 配置列表页
- **WHEN** 管理员访问管理页面
- **THEN** 系统展示所有模型映射配置的表格，包含名称、外部模型、目标模型、目标地址、操作按钮（编辑、删除、测试）

#### Scenario: 创建/编辑配置
- **WHEN** 管理员点击新建或编辑按钮
- **THEN** 系统弹出表单对话框，包含所有可编辑字段
- **AND** 提交后刷新列表

#### Scenario: 删除配置
- **WHEN** 管理员点击删除按钮
- **THEN** 系统弹出确认提示，确认后删除并刷新列表

#### Scenario: 测试连通性
- **WHEN** 管理员点击测试按钮
- **THEN** 系统调用后端测试接口，展示连通性测试结果（成功/失败及详情）

### Requirement: 前端静态资源服务
系统 SHALL 将前端构建产物内嵌到 Go 二进制中，由 Gin 提供静态资源服务，管理页面通过 `/admin` 路径访问。

#### Scenario: 访问管理页面
- **WHEN** 用户浏览器访问 `/admin`
- **THEN** 系统返回 React SPA 的 index.html

### Requirement: 服务配置
系统 SHALL 支持通过环境变量或配置文件配置服务端口。

#### Scenario: 配置服务端口
- **WHEN** 设置环境变量 `LLM_RELAY_PORT` 或配置文件中指定端口
- **THEN** 系统在指定端口启动 HTTP 服务

#### Scenario: 默认端口
- **WHEN** 未指定端口
- **THEN** 系统默认在 8080 端口启动
