package dep_app

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/saintEvol/go-rigger/rigger"
)

const AnotherAppName = "depApp"

func init() {
	rigger.Register(AnotherAppName, rigger.ApplicationBehaviourProducer(func() rigger.ApplicationBehaviour {
		return &depApp{}
	}))
}

type depApp struct {

}

func (a *depApp) OnRestarting(ctx actor.Context) {
}

func (a *depApp) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}


func (a *depApp) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (a *depApp) OnStopping(ctx actor.Context) {
}

func (a *depApp) OnStopped(ctx actor.Context) {
}

func (a *depApp) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	childSpecs = []*rigger.SpawnSpec{
		rigger.DefaultSpawnSpec(depSupName),
	}

	return
}
