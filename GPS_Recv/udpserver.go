package GPS_Recv

import (
	"log"
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
		log.Printf("Error: %s \n", err)
	}
	var hash int = 0
	for _, d := range buf[11:] {
		hash += int(d)
	}
	gethash, _ := strconv.Atoi(string(buf[5:10]))

	if string(buf[0:5]) == "MAGIC" && hash == gethash {
		_, err = conn.WriteToUDP([]byte("200"), addr)
		if err != nil {
			log.Printf("Error: %s \n", err)
		}
	} else {
		_, err = conn.WriteToUDP([]byte("400"), addr)
		if err != nil {
			log.Printf("Error: %s \n", err)
		}
		log.Printf("Inv UDP %s From %v:%v,Ignored\n", string(buf[0:]), addr.IP, addr.Port)
		return
	}
	log.Printf("Recv GPRMC From %v:%v\n%s\n", addr.IP, addr.Port, string(buf[0:]))
	s.Info <- buf[11:len(buf)]
}
