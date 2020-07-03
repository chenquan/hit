# 分布式缓存
## 1. 缓存算法
默认使用了LRU(最近最久未使用)缓存算法
## 2. 使用
### 2.1 编译
方法一:自动编译未可执行文件,生成的文件在`$GOPATH/bin`中,确保将`$GOPATH/bin`其添加到您的中`$PATH`
```shell script
go get -u github.com/chenquan/hit/cmd/hit
```
方法二:手动编译
```shell script
git clone http://github.com/chenquan/hit # 或 git@github.com:chenquan/hit.git
cd hit/cmd/hit
go build -o hit # windows:go build -o hit.exe
```
查看帮助:
```shell script
hit -h # windows:hit.exe -h
```
```shell script
Usage of hit:
  -addr string
        Protoc server port (default "http://localhost:8001")
  -api
        Start a api server? (default true)
  -apiAddr string
        Server api port (default "9999")
  -groups value
        Group cache names (default default,tmp)
  -peerAddrs value
        Peer node addrs (default http://localhost:8001,http://localhost:8002,http://localhost:8003)
```

### 2.2 部署
**单机单例:**
```shell script
hit -peerAddrs=""
```
运行成功:
```shell script
2020/07/03 08:39:10 [Hit] group: default
2020/07/03 08:39:10 [Hit] group: tmp
2020/07/03 08:39:10 [Hit] api server is running at http://localhost:9999
2020/07/03 08:39:10 [Hit] running at http://localhost:8001
```

**单机多例:**
```shell script
hit
```
```shell script
hit -port=http://localhost:8002 -api=0
```
```shell script
hit -port=http://localhost:8003 -api=0
```

测试(PostMan):
```
http://localhost:9999/api?key=Sam&group=default
```
返回:
```shell script
567
```

**分布式:**
手动部署:
部署到三台机器中:
- 机器A:IP:192.168.1.11(开放对外API访问接口)
- 机器B:IP:192.168.1.12
- 机器C:IP:192.168.1.13

机器A:
```shell script
hit -addr=http://192.168.1.11:8001 -apiPort=9999 -peerAddrs="http://192.168.1.11:8001,http://192.168.1.12:8001,http://192.168.1.13:8001" 
```

机器B:
```shell script
hit -addr=http://192.168.1.12:8001  -api=false -peerAddrs="http://192.168.1.11:8001,http://192.168.1.12:8001,http://192.168.1.13:8001" 

```
机器C:
```shell script
hit -addr=http://192.168.1.13:8001  -api=false -peerAddrs="http://192.168.1.11:8001,http://192.168.1.12:8001,http://192.168.1.13:8001" 
```
测试(PostMan):
```
http://192.168.1.11:9999/api?key=Sam&group=default
```
返回:
```shell script
567
```

docker-compose部署:
```shell script
cd hit
docker build -t hit .
cd test
docker-compose up -d
```
测试(PostMan):
```
http://localhost/api?key=Sam&group=default
```
返回:
```shell script
567
```