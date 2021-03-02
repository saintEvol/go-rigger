package normal_starting

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/saintEvol/go-rigger/rigger"
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
	return rigger.BroadcastType
}

func (p playerBroadcastServer) OnGetRoutee() []*actor.PID {
	return nil
}
