package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
	"flag"
)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Err: ", err)
		fmt.Scanln()
		os.Exit(1)
	}
}

func CheckErrorConnection(err error) (status bool) {			// return "true", if connection not established
	if err != nil {
		fmt.Println("err: ", err)
		return true
	}
	return false
}

func incIP(addr *net.UDPAddr) {				// increment IP address in net.UDPAddr struct
	ipaddr := addr.IP
	for i := 3; i >= 0; i-- {
		if ipaddr[i] == 255 {
			ipaddr[i] = 0
		} else {
			ipaddr[i]++
			break
		}
	}
}

func finder(index int, threads int, cycle bool) {				// Blockchain UDP bootstrap without masternode
	threadID := index
	for {
		time.Sleep(5 * time.Millisecond)
		addr, baseIP, err := createRaddr(index)
		if err == nil && addr == nil { 
			index += threads
			continue 
		}
		if err != nil || addr == nil {
			if cycle {
				index = threadID
				continue
			} else {
				fmt.Println("Перебор в потоке ", threadID, " окончен")
				return
			}
			
		}
		fmt.Println("Удаленный адрес: ", addr.String(), " поток ", threadID)
		tm := time.Now().UTC()
		for addr.IP.Mask(baseIP.Mask).Equal(baseIP.IP.Mask(baseIP.Mask)) {
			time.Sleep(5 * time.Millisecond)
			laddr, err := net.ResolveUDPAddr("udp", ":0")
			CheckError(err)
			conn, err := net.DialUDP("udp", laddr, addr)
			if !CheckErrorConnection(err) {
				conn.SetWriteBuffer(128)
//				fmt.Println("Пробуем подключиться к: ", addr.String(), " поток ", threadID)
				conn.Write([]byte(tm.String()))
				incIP(addr)
			} else {
				incIP(addr)
				continue
			}
			//conn.Close()
		}
		index += threads
	}
}

// createRaddr return connection address and base address of connection
// in -> index: row in ip.txt file
// return nil's, if IP addr in blacklist
// return nil, nil, err if index out of range

func createRaddr(index int) (addr *net.UDPAddr, baseaddr *net.IPNet, err error) {	
	fileip, err := os.OpenFile("ip.txt", os.O_RDONLY, 0666)
	CheckError(err)
	fileBlackList, err := os.OpenFile("blacklist.txt", os.O_RDONLY, 0666)
	CheckError(err)
	bListBuf := bufio.NewReader(fileBlackList)
	buf := bufio.NewReader(fileip)
	var strIP string
	var blackStrIP string
	isBlack := false
	i := 1
	for ; i <= index && err != io.EOF; i++ {
		strIP, err = buf.ReadString('\r')
		if err != io.EOF {
			strIP = strings.TrimRight(strIP, "\r\n")
			strIP = strings.TrimSpace(strIP)
		}
	}
	var errbl error
	for errbl != io.EOF {
		blackStrIP, errbl = bListBuf.ReadString('\r')
		if errbl != io.EOF {
			blackStrIP = strings.TrimRight(blackStrIP, "\r\n")
			blackStrIP = strings.TrimSpace(blackStrIP)
			if strings.Compare(strIP, blackStrIP) == 0 {
				strIP = "black"
				isBlack = true
				break
			}
		}
	}
	if err == io.EOF || i < index {
		return nil, nil, err
	}
	if isBlack {
		return nil, nil, nil
	}
	strIP = strings.TrimRight(strIP, "\r\n")
	strIP = strings.TrimSpace(strIP)
	_, ipAddr, err := net.ParseCIDR(strIP)
	CheckError(err)
	_, ipBase, err := net.ParseCIDR(strIP)
	CheckError(err)
	ipStartSearch := ipAddr
	ipStartSearch.IP[3]++
	return &net.UDPAddr{IP: ipStartSearch.IP, Port: 30900}, ipBase, nil
}

func receiveAnswer() {
	for {
		laddr, err := net.ResolveUDPAddr("udp", ":54901")
		CheckError(err)
		conn, err := net.ListenUDP("udp", laddr)
		CheckError(err)
	//	err = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	//	CheckError(err)
		buf := make([]byte, 8192)
		n, raddr, err := conn.ReadFromUDP(buf)
		CheckError(err)
		if strings.Compare(string(buf[0:n]), "я тоже") == 0 {
			fmt.Println("Соединение с ", raddr.String(), " установлено")
			ltcpaddr, err := net.ResolveTCPAddr("tcp", ":50901")
			CheckError(err)
			go tcpConnectionNode(&net.TCPAddr{IP: raddr.IP, Port: 50900}, ltcpaddr)
		}
		conn.Close()
	}
}

// TODO: CREATE TCP CONNECTION WITH CLIENT

func tcpConnectionNode(raddr *net.TCPAddr, laddr *net.TCPAddr) {
	fmt.Println("TCP CONNECTION")
	laddr, err := net.ResolveTCPAddr("tcp", ":50901")
	CheckError(err)
}


func main() {
	thrPtr := flag.Int("t", 1, "threads")
	cycPtr := flag.Bool("c", false, "cycle")
	flag.Parse()
	flag.PrintDefaults()
	time.Sleep(1 * time.Second)
	threads := *thrPtr
	for i := 1; i <= threads; i++ {
		go finder(i, threads, *cycPtr)
		go receiveAnswer()
	}
	fmt.Scanln()
}
