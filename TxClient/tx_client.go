package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net"
	"os"
)

type Msg struct {
	GroupId string
	Type    string
	Command string
	TxCount int
	IsEnd   bool
}
type TxConnection struct {
	Tx      *sql.Tx
	Msg     Msg
	IsStart bool
}

func (tx *TxConnection) Begin() error {
	fmt.Println("代理commit 开启事务")
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "localhost:7778")
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	msg1 := Msg{
		GroupId: "group1",
		Type:    "",
		Command: "create",
		//TxCount: tx.Msg.TxCount,
	}
	bytes, err := json.Marshal(&msg1)
	checkError(err)

	_, err = conn.Write(bytes)
	checkError(err)

	return err
}

func (tx *TxConnection) Commit() error {
	fmt.Println("代理commit")
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "localhost:7778")
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	msg := Msg{
		GroupId: "group1",
		Type:    tx.Msg.Type,
		Command: "add",
		TxCount: tx.Msg.TxCount,
		IsEnd:   tx.Msg.IsEnd,
	}

	bytes, err := json.Marshal(&msg)
	checkError(err)

	_, err = conn.Write(bytes)
	checkError(err)

	for {
		b := make([]byte, 1024)
		n, err := conn.Read(b)
		if err != nil {
			logs.Error(err)
			return err
		}
		reseiveMsg := make([]byte, n)
		reseiveMsg = b[:n]
		msg := Msg{}
		json.Unmarshal(reseiveMsg, &msg)
		if msg.Command == "commit" {
			//收到事务管理器通知提交
			fmt.Println(msg.Command)
			tx.Tx.Commit()
			conn.Close()
			return err
		} else if msg.Command == "rollback" {
			//收到事务管理器通知回滚
			fmt.Println(msg.Command)
			tx.Tx.Rollback()
			conn.Close()
			return err
		}

	}
	return nil
}

func (tx *TxConnection) Rollback() error {
	fmt.Println("代理rollback")
	return tx.Tx.Rollback()
}

// 插入数据事务
func InsertTx1(db *sql.Tx) error {
	stmt, err := db.Prepare("INSERT INTO user(name, age) VALUES(?, ?);")
	if err != nil {
		log.Fatal(err)
		return err
	}
	res, err := stmt.Exec("python1", 18)
	if err != nil {
		log.Fatal(err)
		return err
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
		return err
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Printf("ID=%d, affected=%d\n", lastId, rowCnt)

	return nil
}

func main() {
	db, err := sql.Open("mysql", "root:123456@/test")
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	txConnection := &TxConnection{
		Tx: tx,
	}
	txConnection.Begin()
	txConnection.Msg.Command = "add"
	txConnection.Msg.TxCount = 2
	txConnection.Msg.IsEnd = false
	if err != nil {
		txConnection.Msg.Type = "rollback"
		log.Fatal(err)
	}
	defer db.Close()

	err = InsertTx1(txConnection.Tx)

	if err != nil {
		txConnection.Msg.Type = "rollback"
	}

	txConnection.Msg.Type = "commit"
	txConnection.Commit()
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
	}
}
