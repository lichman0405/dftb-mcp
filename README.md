# DFTB+ Optimization MCP Server

这是一个基于 Model Context Protocol (MCP) 的 DFTB+ 几何优化服务，提供高性能的分子结构优化计算。

## 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   TypeScript    │    │      Go         │    │     DFTB+       │
│     MCP         │◄──►│    Service      │◄──►│   Calculator    │
│    Server       │    │                 │    │                 │
│   (Port 3000)   │    │   (Port 8080)   │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 组件说明

1. **TypeScript MCP Server** - MCP协议服务器，处理与客户端的通信
2. **Go Service** - 高性能计算服务，负责DFTB+计算的实际执行
3. **DFTB+ Calculator** - 量子化学计算引擎

## 功能特性

- 🚀 **高性能** - 使用Go语言实现计算密集型任务
- 🔧 **易于集成** - 基于MCP协议，易于与各种AI助手集成
- 📁 **多格式支持** - 支持CIF文件格式的输入输出
- ⚡ **并发处理** - 支持多个计算请求并发执行
- 🛡️ **错误处理** - 完善的错误处理和日志记录
- 🐳 **容器化** - 完整的Docker容器化部署方案
- 📊 **监控支持** - 可选的Prometheus + Grafana监控
- 🚀 **Redis缓存** - 高性能缓存机制，显著提升重复请求性能
- 📚 **API文档** - 完整的OpenAPI 3.0规范和Swagger UI
- 🤖 **LLM集成** - 支持Gemini CLI、Claude Code、Cline等主流LLM平台

## 快速开始

### 前置要求

- Docker >= 20.10
- Docker Compose >= 2.0
- 至少4GB可用内存

### 1. 克隆项目

```bash
git clone <repository-url>
cd dftbopt-mcp
```

### 2. 验证Docker环境

确保Docker和Docker Compose已正确安装：

```bash
# 验证Docker安装
docker --version

# 验证Docker Compose安装
docker-compose --version
```

**注意**：DFTB+软件现在会自动在容器构建过程中安装，无需手动准备二进制文件。

### 3. 启动基础服务

```bash
# 启动核心服务
docker-compose up -d

# 查看服务状态
docker-compose ps
```

### 4. 验证服务

```bash
# 检查Go服务健康状态
curl http://localhost:8080/health

# 检查TypeScript服务健康状态
curl http://localhost:3000/health

# 获取服务信息
curl http://localhost:8080/info
```

## 详细配置

### 完整部署（包含监控和缓存）

```bash
# 启动所有服务
docker-compose --profile nginx --profile redis --profile postgres --profile monitoring up -d

# 查看所有服务
docker-compose ps
```

### 环境变量配置

创建 `.env` 文件：

```env
# Go Service配置
PORT=8080
WORK_DIR=/app/work
DFTB_PATH=/usr/local/bin/dftb+
MAX_REQUESTS=10
TIMEOUT=300
DEBUG=false
CLEANUP=true

# TypeScript MCP配置
NODE_ENV=production
GO_SERVICE_URL=http://go-service:8080

# 数据库配置（可选）
POSTGRES_DB=dftbopt
POSTGRES_USER=dftbopt
POSTGRES_PASSWORD=your_secure_password

# Redis配置（可选）
REDIS_URL=redis://redis:6379
```

## API使用

### 优化分子结构

```bash
curl -X POST http://localhost:8080/api/v1/optimize \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "test-001",
    "structure_file": "data_chemical\n# 简单的CIF文件内容...",
    "method": "GFN2-xTB",
    "fmax": 0.001,
    "original_filename": "molecule.cif"
  }'
```

### 查询计算状态

```bash
curl http://localhost:8080/api/v1/status/test-001
```

### 服务信息

```bash
curl http://localhost:8080/info
```

## API文档

### Swagger UI

访问交互式API文档：

```
http://localhost:3000/api-docs
```

### OpenAPI JSON

获取API规范JSON格式：

```bash
curl http://localhost:3000/api-docs.json
```

### 可用端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/api/v1/tools` | GET | 获取可用工具列表 |
| `/api/v1/mcp` | POST | MCP协议请求处理 |
| `/api/v1/optimize` | POST | 直接优化请求 |
| `/api-docs` | GET | Swagger UI界面 |
| `/api-docs.json` | GET | OpenAPI规范JSON |

## LLM集成

### 支持的平台

- **Gemini CLI** - 通过MCP服务器配置
- **Claude Code** - 原生MCP支持
- **Cline** - 通过插件系统集成

### Gemini CLI 集成

配置文件 `~/.config/gemini/config.json`：

```json
{
  "mcp_servers": {
    "dftbopt": {
      "command": "node",
      "args": ["dist/server.js"],
      "env": {
        "NODE_ENV": "production",
        "PORT": "3000",
        "GO_SERVICE_URL": "http://localhost:8080",
        "REDIS_URL": "redis://localhost:6379",
        "ENABLE_CACHE": "true"
      }
    }
  }
}
```

使用示例：

```bash
gemini "Optimize this benzene molecule using GFN2-xTB with fmax=0.01"
```

### Claude Code 集成

配置文件 `~/.config/claude/mcp.json`：

```json
{
  "mcpServers": {
    "dftbopt": {
      "command": "node",
      "args": ["dist/server.js"],
      "env": {
        "NODE_ENV": "production",
        "PORT": "3000",
        "GO_SERVICE_URL": "http://localhost:8080",
        "REDIS_URL": "redis://localhost:6379",
        "ENABLE_CACHE": "true"
      }
    }
  }
}
```

使用示例：

```javascript
const result = await tools.dftb_optimize({
  structure_file: "base64-encoded-cif-content",
  method: "GFN2-xTB",
  fmax: 0.01,
  original_filename: "benzene.cif"
});
```

### Cline 集成

插件配置 `~/.config/cline/plugins/dftbopt.yaml`：

```yaml
name: dftbopt
version: "1.0.0"
type: mcp
config:
  command: node
  args: 
    - dist/server.js
  env:
    NODE_ENV: production
    PORT: "3000"
    GO_SERVICE_URL: "http://localhost:8080"
    REDIS_URL: "redis://localhost:6379"
    ENABLE_CACHE: "true"
```

### 自然语言使用示例

```
# 基本优化
Optimize this benzene molecule using DFTB+ with GFN2-xTB method and fmax=0.01.

[CIF content here]

# 高级配置
I need to perform a geometry optimization on a protein-ligand complex. 
Please use the dftb_optimize tool with GFN1-xTB method for faster computation, 
fmax=0.05 for moderate accuracy. The structure is in CIF format.

[CIF content here]

# 比较研究
Please optimize this structure using both GFN1-xTB and GFN2-xTB methods. 
Use fmax=0.01 for both calculations and compare the results.

[CIF content here]
```

### 详细集成指南

完整的LLM集成指南请参考：`typescript-mcp/llm-integration-examples.md`

## 开发指南

### 本地开发环境

#### 1. 安装依赖

**TypeScript服务：**
```bash
cd typescript-mcp
npm install
```

**Go服务：**
```bash
cd go-service
go mod download
```

#### 2. 启动开发服务器

**TypeScript服务：**
```bash
cd typescript-mcp
npm run dev
```

**Go服务：**
```bash
cd go-service
go run cmd/server/main.go -debug=true
```

### 代码结构

```
dftbopt-mcp/
├── typescript-mcp/          # TypeScript MCP服务器
│   ├── src/
│   │   ├── server/          # MCP服务器实现
│   │   ├── tools/           # MCP工具定义
│   │   ├── client/          # Go服务客户端
│   │   └── types/           # TypeScript类型定义
│   ├── package.json
│   └── tsconfig.json
├── go-service/              # Go计算服务
│   ├── cmd/server/         # 主程序入口
│   ├── internal/
│   │   ├── api/           # API处理器
│   │   ├── dftb/          # DFTB+计算逻辑
│   │   ├── parser/        # CIF文件解析器
│   │   └── types/         # Go类型定义
│   └── go.mod
├── docker-compose.yml      # Docker编排配置
└── README.md
```

## 监控和日志

### 日志查看

```bash
# 查看TypeScript服务日志
docker-compose logs -f typescript-mcp

# 查看Go服务日志
docker-compose logs -f go-service

# 查看所有服务日志
docker-compose logs -f
```

### 监控仪表板

如果启用了监控配置，可以通过以下地址访问：

- **Grafana**: http://localhost:3001 (admin/admin)
- **Prometheus**: http://localhost:9090

### 性能指标

系统提供以下监控指标：

- 请求处理时间
- 并发请求数量
- 内存使用情况
- DFTB+计算成功率
- 错误率统计

## 故障排除

### 常见问题

1. **DFTB+可执行文件未找到**
   ```
   Error: DFTB+ executable not found
   ```
   解决方案：重新构建镜像，DFTB+会在构建过程中自动安装：
   ```bash
   docker-compose build --no-cache
   docker-compose up -d
   ```

2. **内存不足**
   ```
   Error: allocation failed
   ```
   解决方案：增加Docker可用内存或减少并发请求数量。

3. **端口冲突**
   ```
   Error: port is already allocated
   ```
   解决方案：修改docker-compose.yml中的端口映射或停止占用端口的服务。

4. **镜像构建失败**
   ```
   Error: conda install failed
   ```
   解决方案：确保网络连接正常，或尝试使用不同的镜像源：
   ```bash
   # 清理Docker缓存
   docker system prune -a
   # 重新构建
   docker-compose build --no-cache
   ```

### 调试模式

启用调试模式以获取更详细的日志信息：

```bash
# Go服务调试
docker-compose run --rm go-service ./main -debug=true

# TypeScript服务调试
docker-compose run --rm typescript-mcp npm run dev
```

## 性能优化

### 建议配置

- **CPU**: 4核心以上
- **内存**: 8GB以上
- **存储**: SSD硬盘，至少10GB可用空间

### 优化建议

1. **并发控制**: 根据系统资源调整 `MAX_REQUESTS` 参数
2. **超时设置**: 根据分子大小调整 `TIMEOUT` 参数
3. **文件清理**: 启用自动清理功能避免磁盘空间不足
4. **缓存策略**: 使用Redis缓存重复计算结果

## 安全考虑

### 网络安全

- 使用Nginx反向代理提供SSL/TLS加密
- 限制外部访问，仅暴露必要端口
- 使用Docker网络隔离服务

### 数据安全

- 不在日志中记录敏感信息
- 定期清理临时文件
- 使用非root用户运行容器

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 许可证

本项目采用MIT许可证。详见 [LICENSE](LICENSE) 文件。

## 联系方式

- 项目主页: [GitHub Repository]
- 问题反馈: [GitHub Issues]
- 技术支持: [Support Email]

## 更新日志

### v1.1.0
- **新增Redis缓存机制** - 显著提升重复请求性能
- **新增OpenAPI文档** - 完整的API规范和Swagger UI界面
- **新增LLM集成指南** - 支持Gemini CLI、Claude Code、Cline等平台
- **优化Docker配置** - 支持Redis缓存和监控服务
- **更新文档** - 完善项目说明和使用指南

### v1.0.0
- 初始版本发布
- 支持基本的DFTB+几何优化
- 实现MCP协议接口
- 完整的Docker容器化部署
