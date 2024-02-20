# xrpc 说明
[English](README_en.md) | 中文

xrpc是用来自动生成rpc服务接口信息代码的工具

## Getting Started
1. 下载xrpc到本地
```git
go install github.com/yc90s/xrpc/cmd/xrpc@latest
```

2. 编写一个接口文件`hello.service`如下, 它定义了一个服务`HelloService`, 包含两个接口
```
package main

service HelloService {
    Hello(string) (string, error)
    Add(*string, int) (*string, error)
}
```

3. 生成代码. 执行下面的命令会在当前目录生成一个hello.service.go文件, 里面定义了`HelloService`服务的rpc接口信息
```
xrpc -out ./ hello.service
```
* `out` 指定输出目录
* `hello.service` 是定义接口的文件, 支持通配符, 如`*.service`

## 接口文件的语法
接口文件最大程度的贴近golang原生语法

1. 定义包名
每个service都需要以`package`开头, `package`后面就是模块的包名, 支持数字、字母和下划线.
```
package main
```

2. 导入其他包
可以用`import`关键字导入依赖的其他包, 每个`import`只能导入一个包.
```
import "github.com/yc90s/xrpc/examples/protobuf/pb"
import "fmt"
```

3. 定义服务
一个接口文件可以定义多个服务, `service`关键字表示定义一个服务, 后接服务名.
```
service HelloService {

}

service WorldService {

}
```

4. 定义服务的接口
每个服务都可以定义多个接口, 每个接口支持任意个参数, 0个或者2个返回值, 如果有2个返回值, 第二个返回值类型为error.
```
service HelloService {
    Hello(string) (string, error)
    Add(*string, int) (*string, error)
    Hi()
    Sum() (int, error)
}
```