package normal_starting

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/saintEvol/go-rigger/rigger"
	"github.com/sirupsen/logrus"
)

const playerBroadcastServerName = "playerBroadcastServer"
func init() {
	rigger.Register(playerBroadcastServerName, rigger.RouterGroupBehaviourProducer(func() rigger.RouterGroupBehaviour {
		return &playerBroadcastServer{}
	}))
}
type playerBroadcastServer struct {

}

func (p playerBroadcastServer) OnGetType() rigger.RouterType {
	logrus.Trace("OnGetType of braodcast")
	return rigger.BroadcastType
}

func (p playerBroadcastServer) OnGetRoutee() []*actor.PID {
	return nil
}
