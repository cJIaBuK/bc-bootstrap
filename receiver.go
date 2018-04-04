package main

import (
	"time"
	"fmt"
	"net"
	"io"
	"os"
)

func CheckError(err error) {
	if err != nil {
		erreof := io.EOF
		if err == erreof {
			fmt.Println("Ввод прерван")
		} else {
			fmt.Println("Err: ", err)
		}
		var input string
		fmt.Scanln(&input)
		os.Exit(1)
	}
}

func sendAnswer(raddr *net.UDPAddr) {
	laddr, err := net.ResolveUDPAddr("udp", ":0")
	CheckError(err)
	raddr.Port = 54901
//	fmt.Println("Отправка на ", raddr)
	conn, err := net.DialUDP("udp", laddr, raddr)
	CheckError(err)
	conn.Write([]byte("я тоже"))
	conn.Close()
}

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":30900")
	CheckError(err) 
	listener, err := net.ListenUDP("udp", addr)
	CheckError(err)
//	defer listener.Close()
	fmt.Println("Запущен на адресе: ", addr.String())
	
	for {
		
		buf := make([]byte, 8192)
		n, addrSend, err := listener.ReadFromUDP(buf)
		CheckError(err)
		if addrSend.IP.String() != "127.0.0.1" {
			go sendAnswer(addrSend)
			fmt.Println("Принято с: ", addrSend.String(), " : ", string(buf[0:n]))
			tm1 := time.Now().UTC()
			const longForm = "2006-01-02 15:04:05.0000000 +0000 UTC"
			tm0, err := time.Parse(longForm, string(buf[0:n])) 
			CheckError(err)
			dur := tm1.Sub(tm0)
			fmt.Println("Нашелся за: ", dur)
		}	
	}
}