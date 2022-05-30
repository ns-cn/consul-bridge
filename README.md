# consul-bridge
consul client's bridge between test and prod

## 解决问题
> 开发时本地consul和测试环境的consul注册信息不统一,导致本地和其他环境**没法联调**或需要**修改很多配置文件**才能联调<br/>
> 通过在本地consul注册对应环境的应用节点,监听目标端口并转发实际的请求实现本地环境接入测试环境

## 下载
下载链接: [windows](./built/consul_bridge_win.exe)、[linux](./built/consul_bridge_linux)、[mac](./built/consul_bridge_darwin)

## 使用

#### 修改配置文件
设置转发规则yaml文件(默认是[consul-bridge.yml](./consul-bridge.yml))
```yaml
consulAddress: 127.0.0.1:8500
agents:
  - { name: "baidu", "port": 8080, to: "www.baidu.com"}
  - { name: "qq", ip: "127.0.0.1", "port": 8081, to: "www.qq.com:8888" }
```
其中

| 配置名 |                  含义                   |       值说明       |
|:----|:-------------------------------------:|:---------------:|
| **consulAddress** |              本地consul地址               |                 |
| **agents** |           转发规则, 可根据实际情况配置多组           |                 |
| **name** |             注册到consul的服务名             |                 |
| **ip**(可选) | 本地服务的Host,如需与局域网共同使用consul做服务联调,需修改该值 | 可选,默认为127.0.0.1 |
| **port** |                服务监听端口                 |                 |
| **to** |                目标转发地址                 |                 |
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