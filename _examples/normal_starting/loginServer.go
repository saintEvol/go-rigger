package normal_starting

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/saintEvol/go-rigger/rigger"
	"github.com/sirupsen/logrus"
)

const loginServerName = "loginServer"

func init() {
	var loginProd rigger.GeneralServerBehaviourProducer = func() rigger.GeneralServerBehaviour {
		return &loginServer{}
	}
	rigger.Register(loginServerName, loginProd)
}

type loginServer struct {
	
}

func (l *loginServer) OnRestarting(ctx actor.Context) {
}

func (l *loginServer) OnStarted(ctx actor.Context, args interface{}) error {
	logrus.Tracef("started: %v", ctx.Self())
	return nil
}

func (l *loginServer) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (l *loginServer) OnStopping(ctx actor.Context) {
}

func (l *loginServer) OnStopped(ctx actor.Context) {
}

func (l *loginServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	return nil
}

