/** stpro v1.0                                           **/
/** The implementation of a simple protocol based on TCP.**/
/** Copyright 2017. Author:by-zhang@github.com           **/
package stpro

import (
	"bufio"
	"errors"
	"net"
	"reflect"
)

type Client struct {
	RAddr *net.TCPAddr
	Conn  *net.TCPConn
}

var enPacketedData []byte
var enPacketChan = make(chan bool)

func NewClient(v interface{}) (Client, error) {
	vTyp = reflect.TypeOf(v)
	vVal = reflect.ValueOf(v)
	cli := Client{}
	err := cli.checkField()
	if err != nil {
		return cli, err
	}
	err = cli.checkMethod()
	if err != nil {
		return cli, err
	}
	return cli, nil
}

func (cli *Client) checkField() error {
	Addr, err := commonFieldCheck()
	if err != nil {
		return err
	}
	cli.RAddr = Addr
	return nil
}

func (cli *Client) checkMethod() error {
	err := commonMethodCheck()
	if err != nil {
		return err
	}
	for i := 0; i < moduleNum; i++ {
		tempMethod := vTyp.Method(i)
		if moduleNames[tempMethod.Name[1:]] == nil {
			err := errors.New("init error: methods are defined incorrectly.")
			return err
		}
		if tempMethod.Type.In(1) != byteSliceType {
			err := errors.New("init error: method `" + tempMethod.Name + "` has wrong type of params.")
			return err
		}
		mapNameFunc[tempMethod.Name[1:]] = tempMethod.Func
	}
	return nil
}

func (cli *Client) dial() error {
	if cli.Conn == nil {
		conn, err := net.DialTCP("tcp", nil, cli.RAddr)
		if err != nil {
			return err
		}
		cli.Conn = conn
	}
	return nil
}

func (cli *Client) Send(typ byte, data []byte) error {
	go cli.enPacket(typ, data)
	err := cli.dial()
	<-enPacketChan
	if err != nil {
		return err
	}
	cli.send()
	finishReply := make(chan bool)
	taskChan := make(chan []byte, 1)
	go cli.Process(taskChan, finishReply)
	go dePacket(cli.Conn, &taskChan)
	<-finishReply
	return nil
}

func (cli *Client) Process(taskChan chan []byte, finishReply chan bool) {
	for recvBuffer := range taskChan {
		packetType := recvBuffer[0]
		data := recvBuffer[1:]
		mapNameFunc[originMap[packetType]].Call([]reflect.Value{vVal, reflect.ValueOf(data)})
		finishReply <- true
	}
}

func (cli *Client) enPacket(typ byte, data []byte) {
	enPacketedData = enPacket(typ, data)
	enPacketChan <- true
}

func (cli *Client) send() {
	bufferWriter := bufio.NewWriter(cli.Conn)
	bufferWriter.Write(enPacketedData)
	bufferWriter.Flush()
}
