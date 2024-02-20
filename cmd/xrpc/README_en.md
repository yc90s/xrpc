# xrpc demo
English | [中文](README.md)

xrpc is a tool used to automatically generate RPC service interface information codes.

## Getting Started
1. Download xrpc
```git
go install github.com/yc90s/xrpc/cmd/xrpc@latest
```

2. Write the interface file `hello.service` as follows, which defines a service `HelloService` that includes two interfaces
```
package main

service HelloService {
    Hello(string) (string, error)
    Add(*string, int) (*string, error)
}
```

3. Generate code. Execute the following command to generate a hello. service. go file in the current directory, which defines the rpc interface information of the `HelloService` service
```
xrpc -out ./ hello.service
```
* `out` output directory
* `hello.service` service file pattern, supports wildcard characters, such as`*.service`

## Grammar
Maximizing the closeness of interface files to Golang's native syntax.

1. Define package name
Each service needs to start with `package`, followed by the module's package name, which supports numbers, letters, and underscores.
```
package main
```

2. import package
You can use the `import` keyword to import other dependent packages, and each `import` can only import one package.
```
import "github.com/yc90s/xrpc/examples/protobuf/pb"
import "fmt"
```

3. Define Services
An interface file can define multiple services, with the keyword `service` indicating the definition of a service, followed by the service name.
```
service HelloService {

}

service WorldService {

}
```

4. Define the interface of the service
Each service can define multiple interfaces, each interface supports any number of parameters, 0 or 2 return values. If there are 2 return values, the second return value type is error.
```
service HelloService {
    Hello(string) (string, error)
    Add(*string, int) (*string, error)
    Hi()
    Sum() (int, error)
}
```