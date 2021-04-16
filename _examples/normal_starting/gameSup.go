package normal_starting

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/saintEvol/go-rigger/rigger"
	"github.com/sirupsen/logrus"
)

const gameSupName = "gameSup"

func init() {
	var supProducer rigger.SupervisorBehaviourProducer = func() rigger.SupervisorBehaviour {
		return &gameSup{}
	}
	rigger.Register(gameSupName, supProducer)
}

// gameSup, 游戏的总监控进程
type gameSup struct {

}

// Interface: SupervisorBehaviours
func (g *gameSup) OnRestarting(ctx actor.Context) {
}

func (g *gameSup) OnStarted(ctx actor.Context, args interface{}) error {
	logrus.Tracef("strted: %v", ctx.Self())
	return nil
}

func (g *gameSup) OnPostStarted(ctx actor.Context, args interface{}) {
	logrus.Tracef("post Started: %v", ctx.Self())
}

func (g *gameSup) OnStopping(ctx actor.Context) {
}

func (g *gameSup) OnStopped(ctx actor.Context) {
}

func (g *gameSup) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	// 监控进程会依次同步启动下列进程,
	childSpecs = append(childSpecs, rigger.SpawnSpecWithKind(gatewayServerName))
	childSpecs = append(childSpecs, rigger.SpawnSpecWithKind(loginServerName))
	childSpecs = append(childSpecs, rigger.SpawnSpecWithKind(playerManagingServerName))
	childSpecs = append(childSpecs, rigger.SpawnSpecWithKind(playerBroadcastServerName))
	childSpecs = append(childSpecs, rigger.SpawnSpecWithKind(playerServerSupName))
	return
}

