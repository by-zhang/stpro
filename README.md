# stpro 一个基于tcp协议实现的简洁通信框架
# a skeleton for communication based on TCP

## 特性

* 引入go包即可使用
* 实现了crc校验，保证数据传输的完整性与正确性
* 调用方式简单明了

## 快速开始

### 1. 引入

```
    import "stpro"
    
```
### 2. server 端

```
    /** 三步搭建服务端
        1 定义任意名称struct的数据结构，必须包含Pmap、Phost两个
          字段，其中Phost为服务端ip+port拼接的字符串，Pmap为自定
          义数据包类型与数据包名称的映射。
        2 实例化对象为字段赋值，实现对应已定义`包名称`的数据包处
          理方法，方法名必为"P[包名称]",如type包的处理方法为Ptype
          。方法中请定义数据处理逻辑,输入输入皆为[]byte类型。
        3 stpro.New()传入实例化的对象，如无报错则服务端开始监听，
          并按照你所定义的逻辑处理数据包，返回响应数据。
    **/
    package main

    import (
        "fmt"
        "stpro"
    )

    type Server struct {
	    Phost string
	    Pmap  map[uint8]string
    }

    func (m Server) Ptype(in []byte) (out []byte) {
	    fmt.Printf("客户端发来type包:%s\n", in)
        /** process... **/
    	bytes := []byte("hello1")
	    return bytes
    }

    func (m Server) Pname(in []byte) (out []byte) {
	    fmt.Printf("客户端发来name包:%s\n", in)
        /** process... **/
        bytes := []byte("hello2")
	    return bytes
    }

    func main() {
	    m := Model{
		    Phost: ":9091",
		    Pmap:  make(map[uint8]string),
	    }
    	m.Pmap[0x01] = "type"
    	m.Pmap[0x02] = "name"
    	err := stpro.New(m)
    	if err != nil {
	    	fmt.Println(err)
	    }
   }

```

### 3.client端

```
    /**
        三部搭建客户端
        1 数据结构同服务端。
        2 P[type]方法是发送对应包后接收到响应数据的处理方法。
        3 实例化对象，并调用Send(type byte, content []byte)方
          法发送数据到客户端，接收到的数据后会自定按照上述定
          义方法处理。
    **/
    package main

    import (
        "fmt"
        "stpro"
    )

    type Client struct {
    	Phost string
	    Pmap  map[byte]string
    }

    func (c Client) Ptype(in []byte) {
	    fmt.Printf("收到了type包的回复:%s\n", in)
    }

    func (c Client) Pname(in []byte) {
	    fmt.Printf("收到了name包的回复:%s\n", in)
    }

    func main() {
	    client, err := stpro.NewClient(Client{
		    Phost: "192.168.1.106:9091",
		    Pmap: map[byte]string{
            0x01: "type",
            0x02: "name",
		    },
	    })

	    if err != nil {
		    fmt.Println(err)
		    return
	    }

	    err = client.Send(0x02, []byte("jintianzhenhao"))
	    if err != nil {
		    fmt.Println(err)
		    return
	    }

	    err = client.Send(0x01, []byte("jintianzhenhao3333"))
	    if err != nil {
		    fmt.Println(err)
		    return
	    }
    }
```
