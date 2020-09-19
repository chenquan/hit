# 分布式缓存
## 1. 缓存算法
默认使用了LRU(最近最久未使用)缓存算法
## 2. 使用

### 2.1 服务端编译
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
Usage of hit.exe:
  -path string
        配置文件地址 (default "hit.toml")

```

### 2.2 部署
配置:hit.toml
```toml
Endpoints=["localhost:2379"]
LeaseTtl=10
DialTimeout=5
NodeAddr="localhost"
NodeName="node1"
Protocol="http"
Port="2020"
```
**单机单例:**
```shell script
hit
```
运行成功:
```shell script
2020-09-19 22:48:20.066321 I | 注册 name: hit/node1 addr: http://localhost:2020
2020-09-19 22:48:20.690561 I | 续租成功节点:node1.
```

**单机多例:**
```shell script
hit -path=test/hit-1.toml
```
```shell script
hit -path=test/hit-2.toml

```
```shell script
hit -path=test/hit-3.toml
```