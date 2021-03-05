package rigger

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
)

type riggerRestServer struct {

}

func (r *riggerRestServer) OnRestarting(ctx actor.Context) {

}

func (r *riggerRestServer) OnStarted(ctx actor.Context, args interface{}) {
}

func (r *riggerRestServer) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (r *riggerRestServer) OnStopping(ctx actor.Context) {
}

func (r *riggerRestServer) OnStopped(ctx actor.Context) {
}

func (r *riggerRestServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	return nil
}

func (r *riggerRestServer) startRestServer()  {

}

