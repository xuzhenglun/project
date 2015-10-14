package GPS_Recv

import (
	"fmt"
	"net"
	"strconv"
)

type ServerUdp struct {
	net.UDPAddr
	Info chan []byte
}

func (s ServerUdp) Listen() {
	addr, err := net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(s.Port))
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	for {
		s.udpHandler(conn)
	}
}

func (s ServerUdp) udpHandler(conn *net.UDPConn) {
	var buf [512]byte
	_, addr, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		fmt.Printf("Error: %s \n", err)
	}
	//fmt.Println(buf)
	_, err = conn.WriteToUDP([]byte("200"), addr)
	if err != nil {
		fmt.Printf("Error: %s \n", err)
	}
	s.Info <- buf[0:len(buf)]
}
