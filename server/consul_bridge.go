package main

import (
	"github.com/ns-cn/goter"
)

var targetFile = goter.NewCmdFlagString("./consul-bridge.yml", "load", "l", "target setting file")

func main() {
	root := goter.NewRootCmd("consul-bridge", "consul bridge between test and prod", VERSION)
	root.AddCommand(cmdProxy.Bind(&targetFile))
	_ = root.Execute()
}
