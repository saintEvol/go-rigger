package rigger

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/sirupsen/logrus"
	"time"
)

const allApplicationTopSupName = "@rg$0"

func init() {
	Register(allApplicationTopSupName, SupervisorBehaviourProducer(func() SupervisorBehaviour {
		return &applicationTopSup{}
	}))
}
type applicationTopSup struct {

}

func (a *applicationTopSup) OnRestarting(ctx actor.Context) {
}

func (a *applicationTopSup) OnStarted(ctx actor.Context, args interface{}) error {
	registeredProcess[allApplicationTopSupName] = ctx.Self()

	logrus.Tracef("started: %v", ctx.Self())
	//if pid, exists := GetPid(allApplicationTopSupName); exists {
	//	logrus.Tracef("top sup: %v", pid)
	//} else {
	//	logrus.Error("faild to get top sup pid\r\n")
	//}
	return nil
}

func (a *applicationTopSup) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (a *applicationTopSup) OnStopping(ctx actor.Context) {
}

func (a *applicationTopSup) OnStopped(ctx actor.Context) {
}

func (a *applicationTopSup) OnGetSupFlag(ctx actor.Context) (supFlag SupervisorFlag, childSpecs []*SpawnSpec) {
	supFlag.WithinDuration = 3 * time.Second
	supFlag.MaxRetries = 10
	supFlag.Decider = func(reason interface{}) actor.Directive {
		return actor.RestartDirective
	}
	supFlag.StrategyFlag = OneForOne

	return
}

