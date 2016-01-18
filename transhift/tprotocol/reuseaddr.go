package tprotocol

import (
	"net"
	"syscall"
	"os"
	"errors"
	"strconv"
)

func NewReusableAddrConn(proto, addr string) (conn net.Conn, err error) {
	const FilePrefix = "port."
	sockAddr, soType, err := getSockAddr(proto, addr)
	if err != nil {
		return
	}
	fd, err := syscall.Socket(soType, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			syscall.Close(fd)
		}
	}()
	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return
	}
	if err = syscall.Bind(fd, sockAddr); err != nil {
		return
	}
	if err = syscall.Listen(fd, syscall.SOMAXCONN); err != nil {
		return
	}
	file := os.NewFile(uintptr(fd), FilePrefix + strconv.Itoa(os.Getpid()))
	if conn, err = net.FileConn(file); err != nil {
		return
	}
	err = file.Close()
	return
}

func getSockAddr(proto, addr string) (sa syscall.Sockaddr, soType int, err error) {
	var addr4 [4]byte
	var addr6 [16]byte

	ip, err := net.ResolveTCPAddr(proto, addr)
	if err != nil {
		return
	}
	switch proto {
	case "tcp4":
		if ip.IP != nil {
			copy(addr4[:], ip.IP[12:16])
		}
		return &syscall.SockaddrInet4{
			Port: ip.Port,
			Addr: addr4,
		}, syscall.AF_INET, nil
	case "tcp6":
		if ip.IP != nil {
			copy(addr6[:], ip.IP)
		}
		return &syscall.SockaddrInet6{
			Port: ip.Port,
			Addr: addr6,
		}, syscall.AF_INET6, nil
	}
	err = errors.New("only tcp4 and tcp6 allowed")
	return
}
