package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"runtime"
	"syscall"

	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/spf13/cobra"
)

const VERSION = "1.05"

var targetSettingFile string

func main() {
	RootCmd.Flags().StringVarP(&targetSettingFile, "load", "l", "./consul-bridge.yml", "target setting file")
	RootCmd.AddCommand(VersionCmd)
	err := RootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

var VersionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "打印当前版本号",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("consul-bridge version: v%s(%s/%s)", VERSION, runtime.GOOS, runtime.GOARCH))
	},
}

// RootCmd 根命令
var RootCmd = &cobra.Command{
	Use:   "consul-bridge",
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

		config := api.DefaultConfig()
		config.Address = setting.ConsulAddress //consul地址
		client, err := api.NewClient(config)
		if err != nil {
			panic(err)
		}
		for _, agent := range setting.Agents {
			if agent.ServiceIP == "" {
				agent.ServiceIP = "127.0.0.1"
			}
			if agent.Using == "" {
				agent.Using = "http"
			}
			if agent.Using == "http" {
				go bridgeWithHttp(client, agent)
			} else if agent.Using == "tcp" {
				go bridgeWithTCP(client, agent)
			} else {
				panic("using must be one of [http, tcp]")
			}
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

// Setting 设置模型
type Setting struct {
	ConsulAddress string        `yaml:"consulAddress"`
	Agents        []ConsulAgent `yaml:"agents"`
}

// ConsulAgent consul代理信息
type ConsulAgent struct {
	ServiceName     string `yaml:"name"`  // 服务名称
	Using           string `yaml:"using"` //	使用的协议
	ServiceIP       string `yaml:"ip"`    //	服务ip
	ServicePort     int    `yaml:"port"`  //	服务端口
	RedirectAddress string `yaml:"to"`    //	重定向地址
}

// 代理udp请求

// 将配置中的服务注册到consul中并启动本地服务监听对应的请求
// 由于存在http和https问题,改为使用tcp进行转发
// bridgeWithHttp 以http方式进行代理
func bridgeWithHttp(client *api.Client, agent ConsulAgent) {
	err := RegistToConsul(client, agent)
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
		var schema = "http"
		if r.URL.Scheme != "" {
			schema = r.URL.Scheme
		}
		url := &url.URL{Host: agent.RedirectAddress, Scheme: schema, Path: r.URL.Path, RawQuery: r.URL.RawQuery}
		reqURL := url.String()
		fmt.Println(url)
		req, err := http.NewRequest(r.Method, reqURL, strings.NewReader(string(body)))
		if err != nil {
			fmt.Print("http.NewRequest ", err.Error())
			return
		}
		for k, v := range r.Header {
			if k == "Host" {
				req.Header.Set("Host", agent.RedirectAddress)
			} else {
				req.Header.Set(k, v[0])
			}
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
		Addr:    fmt.Sprintf("%s:%d", agent.ServiceIP, agent.ServicePort),
		Handler: mux,
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return
	}
}

// 将配置中的服务注册到consul中并启动本地服务监听对应的请求
func bridgeWithTCP(client *api.Client, agent ConsulAgent) {
	err := RegistToConsul(client, agent)
	defer func(agent *api.Agent, serviceID string) {
		_ = agent.ServiceDeregister(serviceID)
	}(client.Agent(), agent.ServiceName)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", agent.ServiceIP, agent.ServicePort))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(listener)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("建立连接错误:%v\n", err)
			continue
		}
		go func(connection net.Conn) {
			defer func(connection net.Conn) {
				_ = connection.Close()
			}(connection)
			proxyConnection, err := net.Dial("tcp", agent.RedirectAddress)
			if err != nil {
				fmt.Printf("连接%v失败:%v\n", agent.RedirectAddress, err)
				return
			}
			ExitChan := make(chan bool, 2)
			proxyTo := func(source net.Conn, dest net.Conn, Exit chan bool) {
				_, err := io.Copy(dest, source)
				if err != nil && err != io.EOF {
					fmt.Printf("往%v发送数据失败:%v\n", agent.RedirectAddress, err)
				}
				ExitChan <- true
			}
			// 转发请求
			go proxyTo(connection, proxyConnection, ExitChan)
			go proxyTo(proxyConnection, connection, ExitChan)
			<-ExitChan
			<-ExitChan
			_ = proxyConnection.Close()
		}(conn)
	}
}

// RegistToConsul 注册到consul
func RegistToConsul(client *api.Client, agent ConsulAgent) error {
	reg := api.AgentServiceRegistration{}
	reg.ID = agent.ServiceName
	reg.Name = agent.ServiceName  //注册service的名字
	reg.Address = agent.ServiceIP //注册service的ip
	reg.Port = agent.ServicePort  //注册service的端口
	reg.Tags = []string{"primary"}

	check := api.AgentServiceCheck{}
	check.TTL = "5s"
	check.TLSSkipVerify = true
	//check.HTTP = fmt.Sprintf("http://%s:%d/actuator/health", agent.ServiceIP, agent.ServicePort) //设置检查使用的url
	//reg.Check = &check
	return client.Agent().ServiceRegister(&reg)
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
