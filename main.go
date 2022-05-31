package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hashicorp/consul/api"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var targetSettingFile string

func main() {
	RootCmd.Flags().StringVarP(&targetSettingFile, "load", "l", "./consul-bridge.yml", "target setting file")
	err := RootCmd.Execute()
	if err != nil {
		return
	}
}

var RootCmd = &cobra.Command{
	Use:   "condge",
	Short: "consul bridge between test and prod",
	Long:  `consul bridge between test and prod`,
	Run: func(cmd *cobra.Command, args []string) {
		var setting Setting
		yamlFile, err := ioutil.ReadFile(targetSettingFile)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		err = yaml.Unmarshal(yamlFile, &setting)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(setting)
		exitChan := make(chan os.Signal)
		signal.Notify(exitChan, os.Interrupt, os.Kill, syscall.SIGTERM)
		go exitHandle(setting, exitChan)
		for _, agent := range setting.Agents {
			go bridge(setting.ConsulAddress, agent)
		}
		select {}
	},
}

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
				fmt.Println("deregister service:", agent.ServiceName)
				err := client.Agent().ServiceDeregister(agent.ServiceName)
				if err != nil {
					log.Fatal(err)
				}
			}
			os.Exit(0)
		}
	}

}

type Setting struct {
	ConsulAddress string        `yaml:"consulAddress"`
	Agents        []ConsulAgent `yaml:"agents"`
}

type ConsulAgent struct {
	ServiceName     string `yaml:"name"`
	ServiceIp       string `yaml:"ip"`
	ServicePort     int    `yaml:"port"`
	RedirectAddress string `yaml:"to"`
}

// 将配置中的服务注册到consul中并启动本地服务监听对应的请求
func bridge(consulAddress string, agent ConsulAgent) {
	if agent.ServiceIp == "" {
		agent.ServiceIp = "127.0.0.1"
	}
	config := api.DefaultConfig()
	config.Address = consulAddress //consul地址
	reg := api.AgentServiceRegistration{}
	reg.ID = agent.ServiceName
	reg.Name = agent.ServiceName  //注册service的名字
	reg.Address = agent.ServiceIp //注册service的ip
	reg.Port = agent.ServicePort  //注册service的端口
	reg.Tags = []string{"primary"}

	check := api.AgentServiceCheck{}
	check.TTL = "5s"
	check.TLSSkipVerify = true
	//check.HTTP = fmt.Sprintf("http://%s:%d/actuator/health", agent.ServiceIp, agent.ServicePort) //设置检查使用的url
	//reg.Check = &check

	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Agent().ServiceRegister(&reg)
	defer func(agent *api.Agent, serviceID string) {
		_ = agent.ServiceDeregister(serviceID)
	}(client.Agent(), agent.ServiceName)
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/actuator/health", healthCheck)
	redirect := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/actuator/health" {
			healthCheck(w, r)
			println("health checking")
			return
		}
		cli := &http.Client{}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Print("io.ReadFull(r.Body, body) ", err.Error())
		}
		url := &url.URL{Host: agent.RedirectAddress, Scheme: "http", Path: r.URL.Path, RawQuery: r.URL.RawQuery}
		reqUrl := url.String()
		fmt.Println(url)
		req, err := http.NewRequest(r.Method, reqUrl, strings.NewReader(string(body)))
		if err != nil {
			fmt.Print("http.NewRequest ", err.Error())
			return
		}
		for k, v := range r.Header {
			req.Header.Set(k, v[0])
		}
		res, err := cli.Do(req)
		if err != nil {
			fmt.Print("cli.Do(req) ", err.Error())
			return
		}
		defer res.Body.Close()
		for k, v := range res.Header {
			w.Header().Set(k, v[0])
		}
		io.Copy(w, res.Body)
	}
	mux.HandleFunc("/", redirect)
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", agent.ServiceIp, agent.ServicePort),
		Handler: mux,
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return
	}
}

// consul健康状态检查, 以做不做服务检查
func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Println("health check")
	result := healthCheckResult{Status: "UP"}
	bytes, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}
	return
}

type healthCheckResult struct {
	Status string `json:"status"`
}