package rigger

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/cluster"
	"github.com/AsynkronIT/protoactor-go/cluster/automanaged"
	"github.com/AsynkronIT/protoactor-go/remote"
	"github.com/labstack/gommon/log"
	"time"
)

var (
	// 预定义变量,表示进程是全局进程且在当前节点启动
	CurrentNodeLocation = Location{
		Host: "localhost",
		Port: 0,
	}

	allNodes = make(map[string]*ClusterNode)
	currentNode string
	// TODO 支持动态添加
	processName2Node = make(map[string]string)
	globalProcessManagingServerCli *GlobalProcessManagingServerGrainClient
	globalProcessManagingServerPid *actor.PID // 全局管理进程的进程id
)

func SetCurrentNode(name string)  {
	currentNode = name
}

// 设置初始节点
func SetNodes(nodes ...*ClusterNode) error {
	for _, node := range nodes {
		if _, exists := allNodes[node.Name]; exists {
			return ErrNodeExists
		}
		allNodes[node.Name] = node
	}

	return nil
}

// 设置集群,调用前应该先设置节点
func SetCluster(clusterName string, managingPort int)  {
	if clusterInstance != nil {
		return
	}

	// how long before the grain poisons itself
	//timeout := 10 * time.Minute
	if root == nil {
		root = actor.NewActorSystem()
	}

	globalProcessManagingServerKind := cluster.NewKind(GlobalProcessManagingServerKindName, actor.PropsFromProducer(func() actor.Actor {
		return &GlobalProcessManagingServerActor{}
	}))
	var addressArr []string
	if len(allNodes) > 0 {
		for _, node := range allNodes {
			addressArr = append(addressArr, node.Location.String())
		}
	}
	provider := automanaged.NewWithConfig(2 * time.Second, managingPort, addressArr...)
	config := remote.Configure("localhost", 0)
	clusterConfig := cluster.Configure(clusterName, provider, config, globalProcessManagingServerKind)
	clusterInstance = cluster.New(root, clusterConfig)
	GlobalProcessManagingServerFactory(func() GlobalProcessManagingServer {
		return &globalProcessManagingServerGrain{}
	})
}

// 将进程名注册到节点
// TODO 进程启动后,然后可以注册为全局进程
func RegisterGlobal(name string, node string) error {
	if _, exists := processName2Node[name]; exists {
		return ErrGlobalNameExists
	}

	processName2Node[name] = node
	return nil
}
func startCluster()  {
	if clusterInstance !=nil {
		clusterInstance.Start()
	}

	if clusterInstance == nil {
		return
	}

	globalProcessManagingServerCli = GetGlobalProcessManagingServerGrainClient(clusterInstance, GlobalProcessManagingServerKindName)
	join()
}

func join()  {
	// 加入节点
	for true {
		if resp, err := globalProcessManagingServerCli.Join(&JoinRequest{Node: currentNode}); err == nil {
			globalProcessManagingServerPid = resp.Pid
			log.Infof("success join node")
			return
		} else {
			log.Errorf("error when join, wait 3 seconds to try again, error: %s ", err.Error())
			<- time.After(1 * time.Second)
		}
	}
}

// 该名字是否属于当前节点
func belongThisNode(name string) bool {
	if node, exists := processName2Node[name]; exists {
		return node == currentNode
	} else {
		return true
	}
}