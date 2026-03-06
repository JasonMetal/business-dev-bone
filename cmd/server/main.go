package main

import (
	"business-dev-bone/internal/core"
)

func main() {
	// basename 用于配置文件名与 env 前缀，新项目可改为自己的名字
	core.NewApp("app").Run()
}
