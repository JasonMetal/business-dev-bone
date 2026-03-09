#business-dev-bone

**纯基础结构**：无业务、无 MySQL/Redis/K8s，只保留 HTTP 服务、配置、优雅关闭和最小路由，用来**在此基础上开发别的应用**。

## 结构（只保留基础）

```
business-dev-bone/
├── go.mod                      # 模块配置文件
├── cmd/server/main.go          # 入口：core.NewApp("app").Run()
├── configs/dev/app.yaml        # 开发配置（端口、日志等）
├── pkg/                        # pkg包 
├── internal/
│   ├── core/                   # 骨架业务：app、server、router、config、options
│   └── framework/              #  app、server、options、middleware（因 Go 禁止跨 module 引用 internal）
└── README.md
```
 

## 以下内容，按需新增

- 所有业务路由、controller、service、store
- MySQL、Redis、K8s、JWT、Analytics、定时任务等配置与逻辑
- 业务错误码、业务 API 文档、Swagger
 

## 如何用来开发新应用

1. **在本仓库内直接改**
   - 在 `internal/core/options/options.go` 里增加你的配置项（如数据库、业务开关）。
   - 在 `internal/core/router.go` 里挂你的路由和 handler。
   - 需要数据层时，可仿照主项目增加 `store` 接口与实现，在 `server.go` 里初始化并注入。

2. **复制出去当新项目**
   - 复制整个 `business-dev-bone` 目录到新仓库。
   - 把 `cmd/server/main.go` 里 `NewApp("app")` 的 `"app"` 改成你的应用名（同时把 `configs/**/app.yaml` 改成同名 yaml）。
   - 新项目需能引用原平台的 `internal/pkg` 和 `pkg/component-base`（例如通过 go mod replace 指向原仓库，或把依赖的包拷贝到新仓库）。

## 构建与运行

**独立构建**：

```bash
cdbusiness-dev-bone
go mod tidy
go build -o server.exe ./cmd/server
./server.exe -c configs/dev/app.yaml
```

或按环境读配置：`./server.exe -e dev`（读 `configs/dev/app.yaml`）。

**在平台仓库内构建**（若business-dev-bone 放在平台下）：

```bash
cd platform
go build -obusiness-dev-bone/server.exe ./business-dev-bone/cmd/server
```

默认 HTTP 端口 8888，提供 `/healthz`、`/version`、`/ping`、`/metrics`、`/debug/pprof`。
