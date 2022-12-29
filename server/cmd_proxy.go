package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/ns-cn/goter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var cmdProxy = &goter.Command{Cmd: &cobra.Command{
	Use:   "proxy",
	Short: "consul bridge between test and prod",
	Run: func(cmd *cobra.Command, args []string) {
		var setting Setting
		yamlFile, err := os.ReadFile(targetFile.Value)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			return
		}
		err = yaml.Unmarshal(yamlFile, &setting)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		setting.InitDefaults()
		setting.PrettyPrint()
		exitChan := make(chan os.Signal)
		signal.Notify(exitChan, os.Interrupt, os.Kill, syscall.SIGTERM)
		go exitHandle(setting, exitChan)

		config := api.DefaultConfig()
		config.Address = setting.ConsulAddress //consul地址
		client, err := api.NewClient(config)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			return
		}
		for _, agent := range setting.Agents {
			if strings.ToLower(agent.Using) == "http" {
				go bridgeWithHttp(client, *agent)
			} else if strings.ToLower(agent.Using) == "tcp" {
				go bridgeWithTCP(client, *agent)
			} else {
				fmt.Println("ERROR: \"using must be one of [http, tcp]\"")
				return
			}
		}
		select {}
	},
}}

// 处理系统的推出信号, 注销consul中的服务
func exitHandle(setting Setting, exitChan chan os.Signal) {
	for {
		select {
		case <-exitChan:
			config := api.DefaultConfig()
			config.Address = setting.ConsulAddress //consul地址
			client, err := api.NewClient(config)   //创建客户端
			if err != nil {
				log.Fatal(err)
			}
			for _, agent := range setting.Agents {
				if !agent.Ignore {
					fmt.Println("deregister service:", agent.ServiceName)
					err := client.Agent().ServiceDeregister(agent.ServiceName)
					if err != nil {
						log.Println(err)
					}
				}
			}
			os.Exit(0)
		}
	}
}
