package main

import (
	"github.com/ns-cn/goter"
)

func main() {
	root := goter.NewRootCmd("consul-bridge", "consul bridge between test and prod", VERSION)
	settingFileFlag := goter.CmdFlagString{P: &targetSettingFile, Value: "./consul-bridge.yml", Name: "load", Shorthand: "l", Usage: "target setting file"}
	root.AddCommand(cmdConsulBridge.Bind(settingFileFlag))
	_ = root.Execute()
}
