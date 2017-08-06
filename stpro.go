/** stpro v1.0                                           **/
/** The implementation of a simple protocol based on TCP.**/
/** Copyright 2017. Author:by-zhang@github.com           **/
package stpro

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"
)

const (
	MAP_NAME  = "Pmap"
	HOST_NAME = "Phost"
)

type Server struct {
	LAddr    *net.TCPAddr
	Listener *net.TCPListener
}

var (
	vTyp          = *new(reflect.Type)
	vVal          = *new(reflect.Value)
	mapTypeString = reflect.TypeOf(new(map[byte]string)).Elem().String()
	byteSliceType = reflect.TypeOf(new([]byte)).Elem()
	moduleNum     = 0
	originMap     = map[byte]string{}
	moduleNames   = map[string]*byte{}
	mapNameFunc   = map[string]reflect.Value{}
)

func (srv *Server) initialize(v interface{}) error {
	vTyp = reflect.TypeOf(v)
	vVal = reflect.ValueOf(v)
	err := srv.checkField()
	if err != nil {
		return err
	}
	err = srv.checkMethod()
	if err != nil {
		return err
	}
	fmt.Println("- Server has been initialized successfully.")
	return nil
}

func (srv *Server) launch() error {
	listener, err := net.ListenTCP("tcp", srv.LAddr)
	if err != nil {
		return errors.New("launch error: server listened failed.")
	}
	defer listener.Close()
	srv.Listener = listener
	fmt.Println("- Server has been launched successfully.")
	fmt.Println("- Server is listening..")
	for {
		conn, err := srv.Listener.Accept()
		if err != nil {
			err = errors.New("launch error: server accepted failed from " + conn.RemoteAddr().String())
			return err
		}
		go srv.handle(conn)
	}
	return nil
}

func (srv *Server) handle(conn net.Conn) {
	defer conn.Close()
	taskChan := make(chan []byte, 10)
	go srv.process(taskChan, conn)
	dePacket(conn, &taskChan)
	return
}

func (srv *Server) process(taskChan chan []byte, conn net.Conn) {
	for recvBuffer := range taskChan {
		packetType := recvBuffer[0]
		data := recvBuffer[1:]
		tempReturn := mapNameFunc[originMap[packetType]].Call([]reflect.Value{vVal, reflect.ValueOf(data)})
		ret := enPacket(packetType, tempReturn[0].Interface().([]byte))
		bufferWriter := bufio.NewWriter(conn)
		bufferWriter.Write(ret)
		bufferWriter.Flush()
	}
}

func (srv *Server) checkField() error {
	Addr, err := commonFieldCheck()
	if err != nil {
		return err
	}
	srv.LAddr = Addr
	return nil
}

func commonFieldCheck() (*net.TCPAddr, error) {
	_, flag := vTyp.FieldByName(HOST_NAME)
	if !flag {
		err := errors.New("init error: field `" + HOST_NAME + "` not found.")
		return nil, err
	}
	fieldStruct, flag := vTyp.FieldByName(MAP_NAME)
	if !flag {
		err := errors.New("init error: field `" + MAP_NAME + "` not found.")
		return nil, err
	}
	mapType := fieldStruct.Type
	if mapType.Kind() != reflect.Map || mapType.String() != mapTypeString {
		err := errors.New("init error: the type of `Pmap` field is supposed to be " + mapTypeString + ".")
		return nil, err
	}
	hostVal := vVal.FieldByName(HOST_NAME).Interface()
	host := string(hostVal.(string))
	rawHost := strings.TrimSpace(host)
	if len(rawHost) < 1 {
		err := errors.New("init error: value of `" + HOST_NAME + "` has not been assigned.")
		return nil, err
	}
	Laddr, err := net.ResolveTCPAddr("tcp", rawHost)
	if err != nil {
		err = errors.New("init error: " + err.Error() + ".")
		return nil, err
	}
	mapVal := vVal.FieldByName(MAP_NAME).Interface()
	originMap = map[byte]string(mapVal.(map[byte]string))
	moduleNum = len(originMap)
	for index, value := range originMap {
		if moduleNames[value] != nil {
			err := errors.New("init error: values of `Pmap` repeated.")
			return nil, err
		}
		moduleNames[value] = &index
	}
	return Laddr, nil
}

func (srv *Server) checkMethod() error {
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
		if tempMethod.Type.In(1) != byteSliceType || tempMethod.Type.Out(0) != byteSliceType {
			err := errors.New("init error: method `" + tempMethod.Name + "` has wrong type of params or values.")
			return err
		}
		mapNameFunc[tempMethod.Name[1:]] = tempMethod.Func
	}
	return nil
}

func commonMethodCheck() error {
	methodNum := vTyp.NumMethod()
	if methodNum != moduleNum {
		err := errors.New("init error: methods are not defined completely.")
		return err
	}
	return nil
}

func New(v interface{}) error {
	server := &Server{}
	err := server.initialize(v)
	if err != nil {
		return err
	}
	err = server.launch()
	if err != nil {
		return err
	}
	return nil
}
