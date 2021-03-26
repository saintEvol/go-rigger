package normal_starting

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/saintEvol/go-rigger/rigger"
	"github.com/sirupsen/logrus"
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
	logrus.Tracef("started: %v", ctx.Self())
	return nil
}

func (g *gatewayServer) OnPostStarted(ctx actor.Context, args interface{}) {
	logrus.Tracef("post Started: %v", ctx.Self())
}

func (g *gatewayServer) OnStopping(ctx actor.Context) {
}

func (g *gatewayServer) OnStopped(ctx actor.Context) {
}

func (g *gatewayServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	return nil
}

