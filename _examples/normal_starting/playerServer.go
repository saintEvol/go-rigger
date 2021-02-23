package normal_starting

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/saintEvol/go-rigger/rigger"
)

const playerServerName = "playerServer"

func init() {
	var producer rigger.GeneralServerBehaviourProducer = func() rigger.GeneralServerBehaviour {
		return &playerServer{}
	}
	// 通过注册start fun动态给玩家进程命名
	rigger.RegisterStartFun(playerServerName, producer,
		func(parent actor.SpawnerContext, props *actor.Props, args interface{}) (pid *actor.PID, err error) {
			name := genPlayerProcessName(args.(uint64))
			return parent.SpawnNamed(props, name)
	})
}

func genPlayerProcessName(id uint64) string {
	return fmt.Sprintf("player_%d", id)
}

type playerServer struct {

}

func (p *playerServer) OnRestarting(ctx actor.Context) {
}

func (p *playerServer) OnStarted(ctx actor.Context, args interface{}) {
	fmt.Printf("player started: %d\r\n", args)
}

func (p *playerServer) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (p *playerServer) OnStopping(ctx actor.Context) {
}

func (p *playerServer) OnStopped(ctx actor.Context) {
}

func (p *playerServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	return nil
}
