package core

import (
	"business-dev-bone/internal/core/config"
	"business-dev-bone/internal/core/options"
	"business-dev-bone/internal/framework/app"
)

const description = `通用 HTTP 服务骨架，无业务逻辑，可在此基础上开发新应用。`

// NewApp 创建仅含基础结构的 CLI 应用，basename 决定配置文件名（如 app.yaml）和 env 前缀
func NewApp(basename string) *app.App {
	opts := options.NewOptions()
	return app.NewApp("App",
		basename,
		app.WithOptions(opts),
		app.WithDescription(description),
		app.WithDefaultValidArgs(),
		app.WithRunFunc(run(opts)),
	)
}

func run(opts *options.Options) app.RunFunc {
	return func(basename string) error {
		cfg, err := config.CreateFromOptions(opts)
		if err != nil {
			return err
		}
		return Run(cfg)
	}
}

// Run 启动 HTTP 服务与优雅关闭
func Run(cfg *config.Config) error {
	srv, err := createServer(cfg)
	if err != nil {
		return err
	}
	return srv.prepareRun().run()
}
