package main

import (
	"fmt"
	"github.com/liushuochen/gotable"
	"strconv"
)

// Setting 设置模型
type Setting struct {
	ConsulAddress string         `yaml:"consulAddress"`
	Agents        []*ConsulAgent `yaml:"agents"`
}

func (setting Setting) InitDefaults() {
	for _, agent := range setting.Agents {
		if agent.ServiceIP == "" {
			agent.ServiceIP = "127.0.0.1"
		}
		if agent.Using == "" {
			agent.Using = "http"
		}
	}
}

// ConsulAgent consul代理信息
type ConsulAgent struct {
	ServiceName     string `yaml:"name"`   // 服务名称
	Using           string `yaml:"using"`  // 使用的协议
	ServiceIP       string `yaml:"ip"`     // 服务ip
	ServicePort     int    `yaml:"port"`   // 服务端口
	RedirectAddress string `yaml:"to"`     // 重定向地址
	Ignore          bool   `yaml:"ignore"` // 是否注册到consul,默认为false
}

func (setting *Setting) PrettyPrint() {
	fmt.Printf("consul: %s\n", setting.ConsulAddress)
	table, err := gotable.Create("服务名称", "代理方式(默认http)", "本地端口", "目标地址", "是否忽略(不注册到consul,默认false)")
	if err != nil {
		fmt.Println("Create table failed: ", err.Error())
		return
	}
	for _, agent := range setting.Agents {
		_ = table.AddRow([]string{agent.ServiceName, agent.Using, strconv.Itoa(agent.ServicePort),
			agent.RedirectAddress, strconv.FormatBool(agent.Ignore)})
	}
	fmt.Println(table)
}
