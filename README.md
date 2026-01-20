# Browser-Auto

自然语言驱动的浏览器自动化操作文档生成系统。通过 AI 解析用户意图，自动执行浏览器操作并生成操作文档。

## 功能特性

- **自然语言理解**：使用 LLM 解析用户描述，自动规划操作步骤
- **浏览器自动化**：基于 Playwright 执行点击、输入、导航等操作
- **多种认证方式**：支持 Cookie、表单登录、SSO 等认证
- **文档生成**：自动截图并生成 HTML/Markdown/PDF 格式文档
- **Web 界面**：提供友好的任务创建和管理界面

## 快速开始

### 前置要求

- Go 1.21+
- Node.js 18+
- Playwright 浏览器

### 安装依赖

```bash
# 后端依赖
go mod download

# 前端依赖
cd frontend && npm install

# 安装 Playwright 浏览器
npx playwright install chromium
```

### 启动服务

```bash
# 启动后端 (端口 8080)
make run
# 或
go run cmd/server/main.go

# 启动前端 (端口 3000)
cd frontend && npm run dev
```

### 访问界面

- 前端界面：http://localhost:3000
- 创建任务：http://localhost:3000/create
- API 文档：http://localhost:8080/health

## 使用指南

### 1. 通过 Web 界面创建任务

1. 访问 http://localhost:3000/create
2. 填写以下信息：
   - **目标 URL**：要操作的网站地址
   - **任务描述**：用自然语言描述要完成的操作
   - **LLM 配置**：选择 AI 模型和参数
   - **认证设置**：配置网站登录方式
   - **输出选项**：选择文档格式

3. 点击"开始执行"等待任务完成

### 2. 通过 API 创建任务

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "description": "进入项目列表，点击第一个项目",
    "target_url": "https://example.com",
    "auth": {
      "type": "cookie",
      "cookies": [
        {"name": "session", "value": "xxx", "domain": ".example.com", "path": "/"}
      ]
    },
    "llm": {
      "provider": "custom",
      "model": "deepseek-v3.1-terminus",
      "endpoint": "http://your-llm-endpoint",
      "api_key": "your-api-key"
    },
    "output": {
      "formats": ["html"]
    }
  }'
```

### 3. 查询任务状态

```bash
# 获取任务详情
curl http://localhost:8080/api/v1/tasks/{task_id}

# 获取任务列表
curl http://localhost:8080/api/v1/tasks
```

## 认证配置

### Cookie 认证

从浏览器获取 Cookie 后配置：

```json
{
  "type": "cookie",
  "cookies": [
    {
      "name": "cookie_name",
      "value": "cookie_value",
      "domain": ".example.com",
      "path": "/"
    }
  ]
}
```

**获取 Cookie 方法**：
1. 在浏览器中登录目标网站
2. 打开开发者工具 (F12) → Application → Cookies
3. 复制关键 Cookie（如 jwt、session 等）

### 表单登录

```json
{
  "type": "form",
  "username": "your_username",
  "password": "your_password"
}
```

### 无认证

```json
{
  "type": "none"
}
```

## LLM 配置

### 自定义 LLM

```json
{
  "provider": "custom",
  "model": "model-name",
  "endpoint": "http://your-llm-endpoint",
  "api_key": "your-api-key",
  "temperature": 0.7,
  "max_tokens": 4096
}
```

### OpenAI

```json
{
  "provider": "openai",
  "model": "gpt-4o",
  "api_key": "sk-xxx"
}
```

## 任务描述编写技巧

### 推荐写法

详细描述每一步操作：

```
1. 点击顶部导航的"xxx"菜单
2. 点击左侧菜单的"xxx"
3. 点击"xxx"子菜单
4. 点击"xxx"按钮
5. ......
6. 点击保存按钮
```

### 不推荐写法

描述过于简单，AI 难以理解具体操作：

```
xxxx
```

## API 参考

### 创建任务

```
POST /api/v1/tasks
```

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| description | string | 是 | 任务描述 |
| target_url | string | 是 | 目标网址 |
| auth | object | 否 | 认证配置 |
| llm | object | 是 | LLM 配置 |
| output | object | 否 | 输出配置 |

### 查询任务

```
GET /api/v1/tasks/{id}
```

**响应**：

| 字段 | 说明 |
|------|------|
| id | 任务 ID |
| status | 状态：pending/running/completed/failed |
| result | 执行结果（包含文档和截图） |
| error | 错误信息 |

### 任务列表

```
GET /api/v1/tasks
```

## 项目结构

```
browser-auto/
├── cmd/
│   └── server/          # 服务入口
├── internal/
│   ├── api/             # HTTP API 处理
│   ├── browser/         # Playwright 浏览器控制
│   ├── domain/          # 领域模型
│   ├── llm/             # LLM 客户端
│   ├── orchestrator/    # 任务编排
│   ├── planner/         # AI 规划器
│   └── docgen/          # 文档生成
├── frontend/            # Next.js 前端
├── configs/             # 配置文件
└── deployments/         # 部署配置
```

## 开发命令

```bash
# 编译
make build

# 运行测试
make test

# 代码检查
make lint

# 清理构建产物
make clean
```

## 常见问题

### Q: Cookie 认证失败？

确保每个 Cookie 都包含 `domain` 和 `path` 字段：
```json
{"name": "jwt", "value": "xxx", "domain": ".example.com", "path": "/"}
```

### Q: 任务一直处于 running 状态？

1. 检查日志：`tail -f /tmp/browser-auto.log`
2. 可能是页面元素未找到，尝试更详细的任务描述

### Q: LLM 调用失败？

检查 LLM 配置的 endpoint 和 api_key 是否正确。

## License

MIT
