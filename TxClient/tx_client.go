package main

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"net"
	"os"
)

func main() {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", "localhost:7777")
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	msg1 := Msg{
		GroupId: "group1",
		Type:    "commit",
		Command: "create",
		TxCount: 2,
	}

	msg2 := Msg{
		GroupId: "group1",
		Type:    "rollback",
		Command: "add",
		TxCount: 2,
	}

	msg3 := Msg{
		GroupId: "group1",
		Type:    "commit",
		Command: "add",
		TxCount: 2,
		IsEnd:   true,
	}

	start := 2
	if start == 1 {
		bytes1, err := json.Marshal(&msg1)
		checkError(err)

		_, err = conn.Write(bytes1)
		checkError(err)

		bytes2, err := json.Marshal(&msg2)
		checkError(err)

		_, err = conn.Write(bytes2)
		checkError(err)
	} else {
		bytes3, err := json.Marshal(&msg3)
		checkError(err)

		_, err = conn.Write(bytes3)
		checkError(err)
	}

	for {
		//result, err := ioutil.ReadAll(conn)
		b := make([]byte, 1024)
		n, err := conn.Read(b)
		if err != nil {
			logs.Error(err)
			return
		}
		reseiveMsg := make([]byte, n)
		reseiveMsg = b[:n]
		msg := Msg{}
		json.Unmarshal(reseiveMsg, &msg)
		if msg.Command == "commit" || msg.Command == "rollback" {
			fmt.Println(msg.Command)
			conn.Close()
			return
		}

	}
}

type Msg1 struct {
	GroupId string
	Type    string
	Command string
	TxCount int
	IsEnd   bool
}

func checkError1(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		//os.Exit(1)
	}
}
