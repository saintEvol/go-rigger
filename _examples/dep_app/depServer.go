package dep_app

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/saintEvol/go-rigger/rigger"
	"github.com/sirupsen/logrus"
	"time"
)

func Echo() error {
	pid, _ := rigger.GetPid(DepSerName)
	f := rigger.Root().Root.RequestFuture(pid, &echo{}, 3 * time.Second)
	err := f.Wait()
	return err
}

func init() {
	rigger.Register(DepSerName, rigger.GeneralServerBehaviourProducer(func() rigger.GeneralServerBehaviour {
		return &depServer{}
	}))
}
type echo struct {

}

const DepSerName = "depServer"

type depServer struct {

}

func (d depServer) OnRestarting(ctx actor.Context) {
}

func (d depServer) OnStarted(ctx actor.Context, args interface{}) error {
	logrus.Tracef("started: %v", ctx.Self())
	return nil
}

func (d depServer) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (d depServer) OnStopping(ctx actor.Context) {
}

func (d depServer) OnStopped(ctx actor.Context) {
}

func (d depServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	switch message.(type) {
	case *echo:
		logrus.Tracef("received echo")
		return nil
	}
	return nil
}

