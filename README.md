"# tx-go" 

<h3>简介</h3>

    golang编写，client、server使用socket通信，实现代理数据源在commit时wait
    ，通过事务管理器通知后实行本地事务提交或回滚。
    
<h3>简述</h3>

    1.开启全局事务，开启事务节点创建全局事务GroupId    
    2.服务见远程调用传递GroupId,本地事务携带GroupId注册至事务管理器的同一个事务组  
    3.代理Datasource执行本地事务，实现commit接口，commit前wait   
    4.事务管理器根据所有本地事务执行情况计算出所有分支事务commit/rollback     
    5.事务管理器通知分支事务commit/rollback  

<h3>逻辑架构图</h3>
![img](https://github.com/xx132917/tx-go/blob/master/image/txarchitecture.png)

<h3>解决问题</h3> 
    
    1.wait时有分支事务宕机：server发现socket连接中断，通知所有分支事务回滚
    2.wait时server宕机：本地事务发现socket连接中断，本地执行回滚
    3.server在通知分支事务commit/rollback时有分支事务宕机：
    4.server集群方案：加上层proxy（proxy可多节点），proxy保存groupId和节点的关系，
      根据groupId将同一个groupId的连接路由到同一个节点
      如果不使用代理，可将数据放入redis中共享，多节点访问同一个redis中数据
      （redis单线程，线程安全）
  
