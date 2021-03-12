package rigger

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"time"
)

const riggerAppName = "@riggerAppName"

func init() {
	Register(riggerAppName, ApplicationBehaviourProducer(func() ApplicationBehaviour {
		return &riggerApp{}
	}))
}

type riggerApp struct {

}

func (r riggerApp) OnRestarting(ctx actor.Context) {
}

func (r riggerApp) OnStarted(ctx actor.Context, args interface{}) {
}

func (r riggerApp) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (r riggerApp) OnStopping(ctx actor.Context) {
}

func (r riggerApp) OnStopped(ctx actor.Context) {
}

func (r riggerApp) OnGetSupFlag(ctx actor.Context) (supFlag SupervisorFlag, childSpecs []*SpawnSpec) {
	supFlag.WithinDuration = 3 * time.Second
	supFlag.MaxRetries = 10
	supFlag.Decider = func(reason interface{}) actor.Directive {
		return actor.RestartDirective
	}
	supFlag.StrategyFlag = OneForOne

	childSpecs = []*SpawnSpec{
		DefaultSpawnSpec(riggerManagingServerName),
	}
	return
}
