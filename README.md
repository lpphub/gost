# gost

Go 微服务工具包。

## 模块

### 核心

| 模块 | 说明 |
| ------ | ------ |
| `config` | 配置管理（基于 viper，支持环境变量覆盖） |
| `log` | 结构化日志（基于 zerolog，日志 sink 通过 WithWriter 可插拔，支持 trace 注入） |
| `otel` | OpenTelemetry 初始化（OTLP gRPC/HTTP 导出，默认本地录制 span 以支持日志 trace_id） |
| `httpx` | HTTP 工具（Gin JSON 响应封装、业务错误类型、pprof） |
| `jwt` | JWT 鉴权（HS256，access/refresh 双令牌） |
| `dbx` | 数据库工具（泛型仓库、基于 context 的事务管理） |

### 框架集成

| 模块 | 说明 |
|------|------|
| `contrib/log` | 日志集成（Gin 中间件、GORM Logger、Redis Hook） |
| `contrib/otel` | 追踪集成（Gin 中间件、GORM/Redis 自动埋点） |

## 使用

推荐按以下顺序初始化各组件：

```go
import (
    "context"
    "os"

    "github.com/lpphub/gost/config"
    "github.com/lpphub/gost/log"
    "github.com/lpphub/gost/otel"
    "github.com/lpphub/gost/dbx"
    "github.com/lpphub/gost/httpx"

    contriblog "github.com/lpphub/gost/contrib/log"
    contribotel "github.com/lpphub/gost/contrib/otel"
)

// 1. 加载配置（最前置）
cfg, _ := config.LoadFile[AppConfig]("./config.yml")

// 2. 初始化日志（紧随配置，确保后续组件日志可输出）
log.Init(
    log.WithLevel(log.DebugLevel),
    log.WithWriter(os.Stdout), // 扩展点：任意 io.Writer——控制台/文件/VictoriaLogs/FluentBit
)

// 3. 初始化 OpenTelemetry（导出/采样配置走标准环境变量，见 otel.Init 文档）
if err := otel.Init(
    otel.WithService("my-app"),
); err != nil {
    panic(err)
}
defer func() { _ = otel.Shutdown(context.Background()) }()

// 4. 创建 Gin 引擎、注册中间件
r := gin.Default()

// 5. 追踪中间件（先注入 span，日志才能读到 trace 信息）
r.Use(contribotel.GinTelemetry("my-app"))

// 6. 日志中间件（从 context 读取 trace，记录完整请求）
r.Use(contriblog.GinRequestLog(
    contriblog.WithSkipPaths("/health", "/metrics"),
))

// 7. GORM：先装日志，再装追踪
db.Logger = contriblog.NewGORMLogger(contriblog.GORMLogCfg{})
db = contribotel.DBTelemetry(db)

// 8. Redis：先装日志 hook，再装追踪 hook
rdb.AddHook(contriblog.NewRedisLogger(contriblog.RedisLogCfg{}))
rdb = contribotel.RedisTelemetry(rdb)

// 9. 启动
httpx.StartPprof(httpx.WithPprofPort(6060))
r.Run(":8080")
```

## 安装

```shell
go get github.com/lpphub/gost
```
