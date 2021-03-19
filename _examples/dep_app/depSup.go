package dep_app

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/saintEvol/go-rigger/rigger"
)

const depSupName = "depSup"

func init() {
	rigger.Register(depSupName, rigger.SupervisorBehaviourProducer(func() rigger.SupervisorBehaviour {
		return &depSup{}
	}))
}
type depSup struct {

}

func (d depSup) OnRestarting(ctx actor.Context) {
}

func (d depSup) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (d depSup) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (d depSup) OnStopping(ctx actor.Context) {
}

func (d depSup) OnStopped(ctx actor.Context) {
}

func (d depSup) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	childSpecs = []*rigger.SpawnSpec{
		rigger.DefaultSpawnSpec(depSerName),
	}

	return
}

