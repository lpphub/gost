# gost

Go 微服务工具包。

## 模块

| 模块 | 说明 |
|------|------|
| `config` | 配置管理（基于 viper，支持环境变量覆盖） |
| `logger` | 结构化日志（基于 zerolog + lumberjack，支持 trace 注入） |
| `httpx` | HTTP 工具（Gin JSON 响应封装、业务错误类型、pprof） |
| `jwt` | JWT 鉴权（HS256，access/refresh 双令牌） |
| `dbx` | 数据库工具（泛型仓库、基于 context 的事务管理） |
| `otel` | OpenTelemetry 初始化（OTLP gRPC 导出 traces & metrics） |
| `contrib/logx` | 日志插件（Gin 中间件、GORM Logger、Redis Hook） |
| `contrib/otelx` | 追踪插件（Gin 中间件、GORM/Redis 自动埋点） |

## 使用

```go
// Gin 请求日志中间件
r.Use(logx.GinTraceLog())

// GORM SQL 日志（慢查询告警）
db.Logger = logx.NewGormLogger()

// Gin 链路追踪中间件
r.Use(otelx.GinMiddleware("my-app"))

// GORM / Redis 自动埋点
db = otelx.WrapDB(db)
rdb = otelx.WrapRDB(rdb)
```

## 安装

```shell
go get github.com/lpphub/gost
```
