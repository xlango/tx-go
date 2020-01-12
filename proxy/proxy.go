package main

//路由规则，将groupId相同的连接请求路由到相同节点

type Rounter struct {
	GroupId string //事务组Id
	Addr    string //服务节点IP：Port
}

func main() {

}
