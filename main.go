package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

// #include <sys/syscall.h>
import "C"

const (
	payload_sz = 32
	chunk_sz   = 1024
	local_port = 5674
)

func main() {
	if len(os.Args) < 4 {
		usage()
	}

	addr, err := net.ResolveUDPAddr("", os.Args[3])
	chk(err)

	iter, err := strconv.Atoi(os.Args[2])
	chk(err)
	iter *= 1E6

	switch os.Args[1] {
	case "write":
		write(addr, iter)
	case "writeToUDP":
		writeToUDP(addr, iter)
	case "sendTo":
		sendTo(addr, iter)
	case "sendMsg":
		sendMsg(addr, iter)
	case "sendMMsg":
		sendMMsg(addr, iter)
	case "All":
		write(addr, iter)
		time.Sleep(time.Second * 2)
		writeToUDP(addr, iter)
		time.Sleep(time.Second * 2)
		sendTo(addr, iter)
		time.Sleep(time.Second * 2)
		sendMsg(addr, iter)
		time.Sleep(time.Second * 2)
		sendMMsg(addr, iter)
	default:
		usage()
	}
}

func usage() {
	log.Fatal("Usage: %s write|writeToUDP|sendTo|sendMsg|sendMMsg|All  iteration_count*1M  target_ip:port", os.Args[0])
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

func write(addr *net.UDPAddr, i int) {
	log.Printf("Start `write` test with %d iteration\n", i)

	conn, err := net.DialUDP(addr.Network(), nil, addr)
	chk(err)

	payload := make([]byte, payload_sz)
	for ; i > 0; i-- {
		_, err := conn.Write(payload)
		chk(err)
	}
}

func writeToUDP(addr *net.UDPAddr, i int) {
	log.Printf("Start `writeToUDP` test with %d iteration\n", i)

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	chk(err)

	payload := make([]byte, payload_sz)
	for ; i > 0; i-- {
		_, err := conn.WriteToUDP(payload, addr)
		chk(err)
	}
}

func sendTo(addr *net.UDPAddr, i int) {
	log.Printf("Start `sendTo` test with %d iteration\n", i)

	laddr := &syscall.SockaddrInet4{Port: local_port, Addr: [4]byte{net.IPv4zero[0], net.IPv4zero[1], net.IPv4zero[2], net.IPv4zero[3]}}
	raddr := &syscall.SockaddrInet4{Port: addr.Port, Addr: [4]byte{addr.IP[0], addr.IP[1], addr.IP[2], addr.IP[3]}}

	fd := connectUDP(laddr, raddr)

	payload := make([]byte, payload_sz)
	for ; i > 0; i-- {
		err := syscall.Sendto(fd, payload, syscall.MSG_DONTWAIT, raddr)
		chk(err)
	}
}

func sendMsg(addr *net.UDPAddr, i int) {
	log.Printf("Start `sendMsg` test with %d iteration\n", i)

	laddr := &syscall.SockaddrInet4{Port: local_port, Addr: [4]byte{net.IPv4zero[0], net.IPv4zero[1], net.IPv4zero[2], net.IPv4zero[3]}}
	raddr := &syscall.SockaddrInet4{Port: addr.Port, Addr: [4]byte{addr.IP[0], addr.IP[1], addr.IP[2], addr.IP[3]}}

	fd := connectUDP(laddr, raddr)

	payload := make([]byte, payload_sz)
	for ; i > 0; i-- {
		err := syscall.Sendmsg(fd, payload, nil, raddr, syscall.MSG_DONTWAIT)
		chk(err)
	}
}

type MMsghdr struct {
	Msg syscall.Msghdr
	cnt int
}

func sendMMsg(addr *net.UDPAddr, i int) {
	i = i / chunk_sz
	log.Printf("Start `sendMMsg` test with %d iteration\n", i)

	laddr := &syscall.SockaddrInet4{Port: local_port, Addr: [4]byte{net.IPv4zero[0], net.IPv4zero[1], net.IPv4zero[2], net.IPv4zero[3]}}
	raddr := &syscall.SockaddrInet4{Port: addr.Port, Addr: [4]byte{addr.IP[0], addr.IP[1], addr.IP[2], addr.IP[3]}}

	fd := connectUDP(laddr, raddr)

	msgcnt := chunk_sz
	var msgArr [chunk_sz]MMsghdr
	for j := 0; j < msgcnt; j++ {
		p := make([]byte, payload_sz)
		for k := 0; k < payload_sz; k++ {
			p[k] = byte(k)
		}

		var iov syscall.Iovec
		iov.Base = (*byte)(unsafe.Pointer(&p[0]))
		iov.SetLen(len(p))

		var msg syscall.Msghdr
		msg.Iov = &iov
		msg.Iovlen = 1

		msgArr[j] = MMsghdr{msg, 0}
	}

	for ; i > 0; i-- {
		//_, _, e1 := syscall.Syscall6(C.SYS_sendmmsg, uintptr(fd), uintptr(unsafe.Pointer(&msgArr[0])), uintptr(msgcnt), uintptr(syscall.MSG_DONTWAIT), 0, 0)
		_, _, e1 := syscall.Syscall6(C.SYS_sendmmsg, uintptr(fd), uintptr(unsafe.Pointer(&msgArr[0])), uintptr(msgcnt), 0, 0, 0)
		if e1 != 0 {
			panic("error on sendmmsg")
		}
	}
}

func connectUDP(laddr, raddr syscall.Sockaddr) int {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	chk(err)

	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	chk(err)

	err = syscall.Bind(fd, laddr)
	chk(err)

	err = syscall.Connect(fd, raddr)
	chk(err)

	return fd
}
