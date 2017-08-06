/** stpro v1.0                                           **/
/** The implementation of a simple protocol based on TCP.**/
/** Copyright 2017. Author:by-zhang@github.com           **/
package stpro

import (
	"bufio"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net"
)

func dePacket(conn net.Conn, taskChan *chan []byte) {
	state := 0x00
	length := uint16(0)
	crc16 := uint16(0)
	var recvBuffer []byte
	cursor := uint16(0)
	reader := bufio.NewReader(conn)
	for {
		recvByte, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("closed: tcp client %s\n", conn.RemoteAddr().String())
				close(*taskChan)
				return
			} else {
				continue
			}

		}
		switch state {
		case 0x00:
			if recvByte == 0xFF {
				state = 0x01
			}
		case 0x01:
			if recvByte == 0xFF {
				state++
			} else {
				state = 0x00
			}
		case 0x02:
			length += uint16(recvByte) * 256
			state++
		case 0x03:
			length += uint16(recvByte)
			recvBuffer = make([]byte, length)
			state++
		case 0x04:
			recvBuffer[cursor] = recvByte
			cursor++
			if cursor == length {
				state++
			}
		case 0x05:
			crc16 += uint16(recvByte) * 256
			state++
		case 0x06:
			crc16 += uint16(recvByte)
			state++
		case 0x07:
			if recvByte == 0xFF {
				state++
			} else {
				state = 0x00
			}
		case 0x08:
			if recvByte == 0xFE {
				err := checkDataCrc(recvBuffer, crc16)
				if err == nil {
					*taskChan <- recvBuffer
				}
			}
			state = 0x00
			length = uint16(0)
			crc16 = uint16(0)
			cursor = uint16(0)
			recvBuffer = nil
		}
	}
}

func checkDataCrc(buffer []byte, crc16 uint16) (err error) {
	if (crc32.ChecksumIEEE(buffer)>>16)&0xFFFF == uint32(crc16) {
		err = nil
	} else {
		err = errors.New("err:data check failed")
	}
	return
}

func enPacket(packetType byte, sendBytes []byte) []byte {
	tempSlice := make([]byte, len(sendBytes)+1)
	tempSlice[0] = packetType
	copy(tempSlice[1:], sendBytes)
	packetLength := len(tempSlice) + 8
	result := make([]byte, packetLength)
	result[0] = 0xFF
	result[1] = 0xFF
	result[2] = byte(uint16(len(tempSlice)) >> 8)
	result[3] = byte(uint16(len(tempSlice)) & 0xFF)
	copy(result[4:], tempSlice)
	sendCrc := crc32.ChecksumIEEE(tempSlice)
	result[packetLength-4] = byte(sendCrc >> 24)
	result[packetLength-3] = byte(sendCrc >> 16 & 0xFF)
	result[packetLength-2] = 0xFF
	result[packetLength-1] = 0xFE
	return result
}
