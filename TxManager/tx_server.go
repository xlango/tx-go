package main

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"net"
	"os"
)

var typeMap map[string][]string
var isEndMap map[string]bool
var countMap map[string]int
var channelGroup map[string][]net.Conn

func init() {
	typeMap = make(map[string][]string)
	isEndMap = make(map[string]bool)
	countMap = make(map[string]int)
	channelGroup = make(map[string][]net.Conn)
}

func main() {
	service := ":7777"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError1(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError1(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

type Msg struct {
	GroupId string
	Type    string
	Command string
	TxCount int
	IsEnd   bool
}

func handleClient(conn net.Conn) {
	//defer conn.Close()

	for {
		b := make([]byte, 1024)
		n, err := conn.Read(b)
		if err!=nil {
			//客户端端口连接（异常断开或完成事务）
			logs.Info("连接已断开",err)
			return
		}
		receiveMsg := make([]byte, n)
		receiveMsg = b[:n]
		fmt.Println(string(receiveMsg))
		msg := Msg{}
		json.Unmarshal(receiveMsg, &msg)
		fmt.Printf("gid:%v,type:%v,command:%v,count:%v,isEnd:%v \n", msg.GroupId, msg.Type, msg.Command, msg.TxCount, msg.IsEnd)

		if msg.Command == "create" {
			//创建事务组
			typeMap[msg.GroupId] = make([]string, 0)
			channelGroup[msg.GroupId] = make([]net.Conn, 0)
		} else if msg.Command == "add" {
			//加入事务组
			typeMap[msg.GroupId] = append(typeMap[msg.GroupId], msg.Type)
			fmt.Printf("%T===%v===%d", typeMap[msg.GroupId], typeMap[msg.GroupId], len(typeMap[msg.GroupId]))

			channelGroup[msg.GroupId] = append(channelGroup[msg.GroupId], conn)

			if msg.IsEnd {
				isEndMap[msg.GroupId] = true
				countMap[msg.GroupId] = msg.TxCount
			}

			rsMsg := Msg{
				GroupId: msg.GroupId,
			}
			if isEndMap[msg.GroupId] && countMap[msg.GroupId] == len(typeMap[msg.GroupId]) {
				if contains(typeMap[msg.GroupId], "rollback") {
					rsMsg.Command = "rollback"
					rsbytes, err := json.Marshal(&rsMsg)
					checkError1(err)
					sendResult(msg.GroupId, rsbytes)
				} else {
					rsMsg.Command = "commit"
					rsbytes, err := json.Marshal(&rsMsg)
					checkError1(err)
					sendResult(msg.GroupId, rsbytes)
				}
			}
		}

	}

}

func sendResult(groupId string, rs []byte) {
	for _, conn := range channelGroup[groupId] {
		conn.Write(rs)
		//conn.Close()
	}


	//通知commit/rollback后解除该事务
	delete(typeMap, groupId)
	delete(isEndMap, groupId)
	delete(countMap, groupId)


	//for _, conn := range channelGroup[groupId] {
	//	conn.Close()
	//}

	delete(channelGroup, groupId)

}

func checkError1(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

//判断[]string 切片中是否包含string
func contains(s []string, v string) bool {
	for _, e := range s {
		if e == v {
			return true
		}
	}

	return false
}
