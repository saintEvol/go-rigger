package normal_starting

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/saintEvol/go-rigger/rigger"
)

const gatewayServerName = "gatewayServer"

func init() {
	var gwProducer rigger.GeneralServerBehaviourProducer = func() rigger.GeneralServerBehaviour {
		return &gatewayServer{}
	}
	rigger.Register(gatewayServerName, gwProducer)
}

// 网关服务器
type gatewayServer struct {
	
}

func (g *gatewayServer) OnRestarting(ctx actor.Context) {
}

func (g *gatewayServer) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (g *gatewayServer) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (g *gatewayServer) OnStopping(ctx actor.Context) {
}

func (g *gatewayServer) OnStopped(ctx actor.Context) {
}

func (g *gatewayServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	return nil
}

