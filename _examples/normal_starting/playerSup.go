package normal_starting

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/saintEvol/go-rigger/rigger"
)


const playerServerSupName = "playerServerSup"

func init() {
	var producer rigger.SupervisorBehaviourProducer = func() rigger.SupervisorBehaviour {
		return &playerServerSup{}
	}
	rigger.Register(playerServerSupName, producer)
}

type playerServerSup struct {

}

func (p *playerServerSup) OnRestarting(ctx actor.Context) {
}

func (p *playerServerSup) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (p *playerServerSup) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (p *playerServerSup) OnStopping(ctx actor.Context) {
}

func (p *playerServerSup) OnStopped(ctx actor.Context) {
}

func (p *playerServerSup) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	supFlag.StrategyFlag = rigger.SimpleOneForOne // 将子进程(玩家进程)变为动态进程
	childSpecs = append(childSpecs, rigger.DefaultSpawnSpec(playerServerName))

	return
}

