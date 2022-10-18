# consul-bridge
consul-bridge: 架设多个consul环境之间的桥梁

## 解决问题

- [x] 开发时需要本地开发调试，但本地服务又存在服务依赖问题
- [x] 本地配置不足以完整运行整套微服务
- [x] 需要接入其他环境做测试、debug等
- [x] 每次接入其他环境就需要更改大量的配置参数 

## 实现方式
- [x] consul-bridge根据配置文件注册多个微服务实例到指定的consul(或本机的consul)
- [x] 通过本地监听端口实现对微服务请求做转发,接入其他测试环境

## 下载
下载链接: [windows](./built/consul_bridge_win.exe)、[linux](./built/consul_bridge_linux)、[mac](./built/consul_bridge_darwin)
或使用brew方式安装
```
brew tap ns-cn/ttools && brew install consul-bridge
```

## 使用

#### 修改配置文件
设置转发规则yaml文件(默认是[consul-bridge.yml](./consul-bridge.yml))
```yaml
consulAddress: 127.0.0.1:8500
agents:
  - { name: "baidu", "port": 8080, to: "www.baidu.com:80"}
  - { name: "micro-service-1", using: "http", "port": 10010, to: "remote-micro-service-1:10010"}
  - { name: "mysql", using: "tcp", "port": 3306, to: "remote-server:3306", ignore: true}
  - { name: "redis", using: "tcp", "port": 6379, to: "remote-server:6379", ignore: true}
  - { name: "rabbitmq", using: "tcp", "port": 5672, to: "remote-server:5672", ignore: true}
  - { name: "rabbitmq-ui", using: "http", "port": 15672, to: "remote-server:15672"}
```
其中

| 配置名               |                  含义                   |         值说明          |
|:------------------|:-------------------------------------:|:--------------------:|
| **consulAddress** |              本地consul地址               |                      |
| **agents**        |           转发规则, 可根据实际情况配置多组           |                      |
| **name**          |             注册到consul的服务名             |                      |
| **using**(可选)     |                端口监听方式                 | 当前支持:http/tcp,默认http |
| **ip**(可选)        | 本地服务的Host,如需与局域网共同使用consul做服务联调,需修改该值 |   可选,默认为127.0.0.1    |
| **port**          |                服务监听端口                 |                      |
| **to**            |                目标转发地址                 |                      |
| **ignore**        |       是否忽略(不注册到consul,默认false)        |     可选,默认值为false     |
#### 启动
> 注: 可使用参数```--load```或```-l```指定配置文件
```shell
# windows 使用默认文件(./consul_bridge.yml)加载
./consul_bridge.exe
# windows 指定配置文件加载
./consul_bridge.exe -l ./other_setting.yml
./consul_bridge.exe --load ./other_setting.yml


# linux 使用默认文件(./consul_bridge.yml)加载
./consul_bridge_linux
# windows 指定配置文件加载
./consul_bridge_linux -l ./other_setting.yml
./consul_bridge_linux --load ./other_setting.yml


# darwin 使用默认文件(./consul_bridge.yml)加载
./consul_bridge_darwin
# windows 指定配置文件加载
./onsul_bridge_darwin -l ./other_setting.yml
./onsul_bridge_darwin --load ./other_setting.yml
```