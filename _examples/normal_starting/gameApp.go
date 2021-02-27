package normal_starting

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/saintEvol/go-rigger/rigger"
)

const gameAppName = "gameApp"

func init()  {
	var appProducer rigger.ApplicationBehaviourProducer = func() rigger.ApplicationBehaviour {
		return &gameApp{}
	}
	rigger.Register(gameAppName, appProducer)
}
// gameApp
type gameApp struct {

}

// Interface: ApplicationBehaviour
// 当进程重启时会被回调, 不过Application目前不会被重启
func (g *gameApp) OnRestarting(ctx actor.Context) {
}

// 进程启动时的回调, 此回调成功返回后,监控进程才会认为此进程已经成功启动,因此不宜在此回调中进行较费时的操作
func (g *gameApp) OnStarted(ctx actor.Context, args interface{}) {
	fmt.Printf("%s Started\r\n", gameAppName)
}

// 进程启动时回调, 不过在OnStarted之后,可以在此进行一些比较耗时的初始化工作
func (g *gameApp) OnPostStarted(ctx actor.Context, args interface{}) {
}

// 进程即将停止时回调
func (g *gameApp) OnStopping(ctx actor.Context) {
}

// 进程被停止后的回调
func (g *gameApp) OnStopped(ctx actor.Context) {
}

// 获取对子进程的监控标志/参数及其子进程规范,以便 go-rigger对子进程进行启动和管理
func (g *gameApp) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	// 只有一个子进程,也就是整个游戏的主监控树
	childSpecs = append(childSpecs, rigger.DefaultSpawnSpec(gameSupName))
	return
}


