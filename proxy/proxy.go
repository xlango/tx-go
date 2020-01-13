package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

//路由规则，将groupId相同的连接请求路由到相同节点

type Rounter struct {
	GroupId string //事务组Id
	Addr    string //服务节点IP：Port
}

var lock sync.Mutex
var trueList []string
var ip string
var list string

func main() {
	flag.StringVar(&ip, "l", ":9897", "-l=0.0.0.0:9897 指定服务监听的端口")
	flag.StringVar(&list, "d", "127.0.0.1:1789,127.0.0.1:1788", "-d=127.0.0.1:1789,127.0.0.1:1788 指定后端的IP和端口,多个用','隔开")
	flag.Parse()
	trueList = strings.Split(list, ",")
	if len(trueList) <= 0 {
		fmt.Println("后端IP和端口不能空,或者无效")
		os.Exit(1)
	}
	server()
}

func server() {
	lis, err := net.Listen("tcp", ip)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer lis.Close()
	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("建立连接错误:%v\n", err)
			continue
		}
		fmt.Println(conn.RemoteAddr(), conn.LocalAddr())
		go handle(conn)
	}
}

func handle(sconn net.Conn) {
	defer sconn.Close()
	ip, ok := getIP()
	if !ok {
		return
	}
	dconn, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Printf("连接%v失败:%v\n", ip, err)
		return
	}
	ExitChan := make(chan bool, 1)
	go func(sconn net.Conn, dconn net.Conn, Exit chan bool) {
		_, err := io.Copy(dconn, sconn)
		fmt.Printf("往%v发送数据失败:%v\n", ip, err)
		ExitChan <- true
	}(sconn, dconn, ExitChan)
	go func(sconn net.Conn, dconn net.Conn, Exit chan bool) {
		_, err := io.Copy(sconn, dconn)
		fmt.Printf("从%v接收数据失败:%v\n", ip, err)
		ExitChan <- true
	}(sconn, dconn, ExitChan)
	<-ExitChan
	dconn.Close()
}

func getIP() (string, bool) {
	lock.Lock()
	defer lock.Unlock()

	if len(trueList) < 1 {
		return "", false
	}
	ip := trueList[0]
	trueList = append(trueList[1:], ip)
	return ip, true
}
