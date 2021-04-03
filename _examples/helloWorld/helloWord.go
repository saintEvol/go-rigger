package helloWorld

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/saintEvol/go-rigger/rigger"
	"time"
)


func Start()  {
	_ = rigger.Start(appName, "")
}

/*
hello world 应用
*/
// 应用名,应用标识
const appName = "helloWorldApp"
func init() {
	rigger.Register(appName, rigger.ApplicationBehaviourProducer(func() rigger.ApplicationBehaviour {
		return &helloWorldApp{}
	}))

	// 依赖 rigger-amqp应用, 保证在启动helloWroldApp 前,会先启动 rigger-amqp
	// rigger.DependOn("rigger-amqp")
}
type helloWorldApp struct {

}

func (h *helloWorldApp) OnRestarting(ctx actor.Context) {
}

// 如果返回的错误非空,则表示启动失败,则会停止启动后续进程
func (h *helloWorldApp) OnStarted(ctx actor.Context, args interface{}) error {
	fmt.Print("start hello world\r\n")
	return nil
}

func (h *helloWorldApp) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (h *helloWorldApp) OnStopping(ctx actor.Context) {
}

func (h *helloWorldApp) OnStopped(ctx actor.Context) {
}

func (h *helloWorldApp) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	// 一对一
	supFlag.StrategyFlag = rigger.OneForOne
	// 最多尝试10次
	supFlag.MaxRetries = 10
	// 最多尝试1秒
	supFlag.WithinDuration = 1 * time.Second
	// 任何原因下都重启
	supFlag.Decider = func(reason interface{}) actor.Directive {
		return actor.RestartDirective
	}

	// 将helloWorldSup 设置为应用的子进程
	childSpecs = []*rigger.SpawnSpec {
		rigger.DefaultSpawnSpec(helloWorldSupName),
	}

	return
}

/*
helloWorld 的监控进程,负责对业务子进程的监控及重启
此进程不处理任何业务
*/
const helloWorldSupName = "helloWorldSup"

func init() {
	rigger.Register(helloWorldSupName, rigger.SupervisorBehaviourProducer(func() rigger.SupervisorBehaviour {
		return &helloWorldSup{}
	}))
}
type helloWorldSup struct {
	
}

func (h helloWorldSup) OnRestarting(ctx actor.Context) {
}

// 如果返回非空,则表示启动失败,会停止后续进程的启动
func (h helloWorldSup) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (h helloWorldSup) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (h helloWorldSup) OnStopping(ctx actor.Context) {
}

func (h helloWorldSup) OnStopped(ctx actor.Context) {
}

func (h helloWorldSup) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	childSpecs = []*rigger.SpawnSpec{
		// 配置一个子进程(业务进程)
		rigger.DefaultSpawnSpec(helloWorldServerName),
	}
	return
}

const helloWorldServerName = "helloWorldServer"

/*
helloWorld服务器, 负责响应外部消息,完成业务处理
*/
func init() {
	rigger.Register(helloWorldServerName, rigger.GeneralServerBehaviourProducer(func() rigger.GeneralServerBehaviour {
		return &helloWordServer{}
	}))
}
type helloWordServer struct {
	
}

func (h *helloWordServer) OnRestarting(ctx actor.Context) {
}

func (h *helloWordServer) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (h *helloWordServer) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (h *helloWordServer) OnStopping(ctx actor.Context) {
}

func (h *helloWordServer) OnStopped(ctx actor.Context) {
}

// 消息处理, 框架会根据需要将返回值回复给请求进程
func (h *helloWordServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	return nil
}

