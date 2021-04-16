package cluster

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/cluster"
	"github.com/AsynkronIT/protoactor-go/cluster/consul"
	"github.com/AsynkronIT/protoactor-go/remote"
)

func Start()  {
	sys := actor.NewActorSystem()
	remoteConfig := remote.Configure("127.0.0.1", 9090)
	if provider, err := consul.New(); err == nil {
		clusterConfig := cluster.Configure("mycluster", provider, remoteConfig)
		clu := cluster.New(sys, clusterConfig)
		clu.Start()
	}
}
