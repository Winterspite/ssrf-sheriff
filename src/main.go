package main

import (
	"github.com/winterspite/ssrf-sheriff/src/handler"
	"go.uber.org/fx"
)

func init() {
	initEnv()
}

func main() {
	fx.New(opts()).Run()
}

func opts() fx.Option {
	return fx.Options(
		fx.Provide(
			handler.NewLogger,
			handler.NewSSRFSheriffRouter,
			handler.NewServerRouter,
			handler.NewHTTPServer,
		),
		fx.Invoke(handler.StartFilesGenerator, handler.StartServer),
	)
}
