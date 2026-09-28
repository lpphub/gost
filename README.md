# gost

Go 微服务工具包。

## 模块

### 核心

| 模块 | 说明 |
| ------ | ------ |
| `config` | 配置管理（基于 viper，支持环境变量覆盖） |
| `log` | 结构化日志（基于 zerolog，日志 sink 通过 WithWriter 可插拔，支持 trace 注入） |
| `otel` | OpenTelemetry 初始化（默认只本地录制 span 以支持日志 trace_id；配了 OTLP 地址才开始导出，导出器/地址/协议/采样全走标准环境变量，gRPC 等导出器用 `RegisterTraceExporter`/`RegisterMetricReader` 注册） |
| `httpx` | HTTP 工具（Gin JSON 响应封装、业务错误类型、pprof） |
| `jwt` | JWT 鉴权（HS256，access/refresh 双令牌） |
| `dbx` | 数据库工具（泛型仓库、基于 context 的事务管理） |

### 框架集成

| 模块 | 说明 |
| ------ | ------ |
| `contrib/log` | 日志集成（Gin 中间件、GORM Logger、Redis Hook） |
| `contrib/obs` | 可观测性集成（Gin 中间件组合、GORM/Redis 追踪埋点） |

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
    "github.com/lpphub/gost/contrib/obs"
)

// 1. 加载配置（最前置）
cfg, _ := config.LoadFile[AppConfig]("./config.yml")

// 2. 初始化日志（紧随配置，确保后续组件日志可输出）
log.Init(
    log.WithLevel(log.DebugLevel),
    log.WithWriter(os.Stdout), // 扩展点：任意 io.Writer——控制台/文件/VictoriaLogs/FluentBit
)

// 3. 初始化 OpenTelemetry（导出/采样全走标准环境变量；什么都不配时只录制、不导出）
if err := otel.Init("my-app"); err != nil {
    panic(err)
}
defer func() { _ = otel.Shutdown(context.Background()) }()

// 4. 创建 Gin 引擎、注册中间件
r := gin.Default()

// 5. 追踪 + 请求日志中间件（obs 负责顺序：先注入 span，日志才能读到 trace）
r.Use(obs.Gin("my-app",
    contriblog.WithSkipPaths("/health", "/metrics"),
)...)

// 6. GORM：日志 + 追踪
db.Logger = contriblog.NewGORMLogger(contriblog.GORMLogCfg{})
db = obs.DB(db)

// 7. Redis：日志 hook + 追踪 hook
rdb.AddHook(contriblog.NewRedisLogger(contriblog.RedisLogCfg{}))
rdb = obs.Redis(rdb)

// 8. 启动
httpx.StartPprof(httpx.WithPprofPort(6060))
r.Run(":8080")
```

`obs.Gin` 的顺序是硬约束：`GinRequestLog` 从请求 ctx 里读 span，所以追踪中间件必须先跑。GORM/Redis 侧的日志与追踪互不依赖，先后随意——它们日志里的 `trace_id` 来自 HTTP 请求 ctx 的透传。

## 自定义导出器（例如 gRPC）

`otel` 只内置 OTLP/HTTP 导出器，gRPC 导出器不进 gost 依赖。想让「改环境变量就能切协议」的项目，注册一次即可：

```go
import (
    "context"

    "github.com/lpphub/gost/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    metricsdk "go.opentelemetry.io/otel/sdk/metric"
    tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

// 放在自己包的 init 里，例如 telemetry/otel_grpc.go
func init() {
    otel.RegisterTraceExporter(otel.ProtocolGRPC, func(ctx context.Context) (tracesdk.SpanExporter, error) {
        return otlptracegrpc.New(ctx)
    })
    otel.RegisterMetricReader(otel.ProtocolGRPC, func(ctx context.Context) (metricsdk.Reader, error) {
        exp, err := otlpmetricgrpc.New(ctx)
        if err != nil {
            return nil, err
        }
        return metricsdk.NewPeriodicReader(exp), nil
    })
}
```

之后就是纯环境变量切换：`OTEL_EXPORTER_OTLP_PROTOCOL=grpc`（或 `http/protobuf`）。

traces 与 metrics 各有一张表（`RegisterTraceExporter` / `RegisterMetricReader`），**按需注册**：只想要 traces 用 `console` 就只注册 traces 那一条，metrics 不受影响。内置 HTTP 是两张表里的默认条目（不需要注册，也可以用 `ProtocolHTTP` 作 key 覆盖）。表的 key 既可以是 OTLP 协议（由 `OTEL_EXPORTER_OTLP_PROTOCOL` 选中），也可以是 `OTEL_TRACES_EXPORTER` / `OTEL_METRICS_EXPORTER` 的取值（例如 `console`，命中后不需要 endpoint）：

```go
otel.RegisterTraceExporter("console", func(context.Context) (tracesdk.SpanExporter, error) {
    return stdouttrace.New(stdouttrace.WithPrettyPrint())
})
// OTEL_TRACES_EXPORTER=console
```

`console` / `prometheus` 这类实现不在 gost 依赖里，由使用方自己 `go get` 并注册；gost 只提供「名字 → 导出器」的机制。

环境变量值是非法或未注册的，`Init` 一开头就报错（不静默降级、也不等到真要建导出器时才报）。

## 环境变量

`otel` 自己只读 6 个判定变量；其余（headers/压缩/超时/证书等连接参数、`OTEL_SERVICE_NAME`、`OTEL_RESOURCE_ATTRIBUTES`、`OTEL_TRACES_SAMPLER`）由 SDK 与导出器自动读取。服务名优先级：`Init` 参数 > `OTEL_SERVICE_NAME` / `OTEL_RESOURCE_ATTRIBUTES`（后者经 SDK 的 `resource.Default()`，进程内只读一次），都没有则用 SDK 默认的 `unknown_service:*`。

| 变量 | 取值 | 行为 |
| --- | --- | --- |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | 任意 | 出现即视为「要导出」，作为 traces/metrics 的兜底端点 |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | 任意 | 同上，traces 专用 |
| `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` | 任意 | 同上，metrics 专用 |
| `OTEL_TRACES_EXPORTER` | 空 / `otlp` | 按 endpoint 导出；直接写 `otlp` 却没配 endpoint 会报错 |
| | `none` | 关闭导出；span 仍本地录制，日志 trace_id 照常可用 |
| | 已注册的名字（如 `console`） | 用 `RegisterTraceExporter` 注册的实现；不需要 endpoint |
| | 其它 | 报错，不静默降级 |
| `OTEL_METRICS_EXPORTER` | 空 / `otlp` / `none` / 已注册的名字 | 同上（metrics 侧注册函数是 `RegisterMetricReader`） |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | 空 / `http/protobuf` | 内置 OTLP/HTTP |
| | 已注册的值（如 `grpc`） | 用注册的导出器，见上一节 |
| | 其它 | 报错 |

没有配置任何 endpoint 就等同于不需要导出：不创建导出器、不报错；此时 span 仍会本地录制，日志里的 trace_id 照常可用。

```go
if err := otel.Init("my-app"); err != nil {   // 最简形式；Init("") 则走 OTEL_SERVICE_NAME
    panic(err)
}
defer func() { _ = otel.Shutdown(context.Background()) }()
```

`Shutdown` 幂等，未初始化时调用返回 nil；重复 `Init` 会用新的一对 provider 覆盖包内记录，旧的那对不会自动关闭（进程级组件按约定只初始化一次）。

## 安装

```shell
go get github.com/lpphub/gost
```
