# gost

Go 微服务工具包。

## 模块

### 核心

| 模块 | 说明 |
|------|------|
| `config` | 配置管理（基于 viper，支持环境变量覆盖） |
| `log` | 结构化日志（基于 zerolog + lumberjack，支持 trace 注入） |
| `otel` | OpenTelemetry 初始化（OTLP gRPC 导出，支持 Prometheus 拉取 metrics） |
| `httpx` | HTTP 工具（Gin JSON 响应封装、业务错误类型、pprof） |
| `jwt` | JWT 鉴权（HS256，access/refresh 双令牌） |
| `dbx` | 数据库工具（泛型仓库、基于 context 的事务管理） |

### 适配器

| 模块 | 说明 |
|------|------|
| `adapters/log` | 日志适配器（Gin 中间件、GORM Logger、Redis Hook） |
| `adapters/otel` | 追踪适配器（Gin 中间件、GORM/Redis 自动埋点） |

## 使用

推荐按以下顺序初始化各组件：

```go
import (
    "github.com/lpphub/gost/config"
    "github.com/lpphub/gost/log"
    "github.com/lpphub/gost/otel"
    "github.com/lpphub/gost/dbx"
    "github.com/lpphub/gost/httpx"

    logadapt "github.com/lpphub/gost/adapters/log"
    oteladapt "github.com/lpphub/gost/adapters/otel"
)

// 1. 加载配置（最前置）
cfg, _ := config.LoadFile[AppConfig]("./config.yml")

// 2. 初始化日志（紧随配置，确保后续组件日志可输出）
log.Init(
    log.WithLevel(log.DebugLevel),
    log.WithOutputFile("logs/app.log"),
)

// 3. 初始化 OpenTelemetry
otel.Init(
    otel.WithService("my-app"),
    otel.WithOTLPEndpoint("otel-collector:4317"),
    otel.WithPrometheus(),
)

// 4. 创建 Gin 引擎、注册中间件
r := gin.Default()

// 5. 追踪中间件（先注入 span，日志才能读到 trace 信息）
r.Use(oteladapt.GinTelemetry("my-app"))

// 6. 日志中间件（从 context 读取 trace，记录完整请求）
r.Use(logadapt.GinRequestLog(
    logadapt.WithSkipPaths("/health", "/metrics"),
))

// 7. Prometheus metrics 端点
oteladapt.RegisterMetricsEndpoint(r, "/metrics")

// 8. GORM：先装日志，再装追踪
db.Logger = logadapt.NewGormLogger()
db = oteladapt.DBTelemetry(db)

// 9. Redis：先装日志 hook，再装追踪 hook
rdb.AddHook(logadapt.NewRedisLogger())
rdb = oteladapt.RedisTelemetry(rdb)

// 10. 启动
httpx.StartPprof(httpx.WithPprofPort(6060))
r.Run(":8080")
```

## 安装

```shell
go get github.com/lpphub/gost
```
