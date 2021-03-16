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
	broadcastPid *actor.PID
}

func (p *playerServer) OnRestarting(ctx actor.Context) {
}

func (p *playerServer) OnStarted(ctx actor.Context, args interface{}) error {
	fmt.Printf("player started: %d\r\n", args)
	// 将自己加入广播
	if pid, ok := rigger.GetPid(playerBroadcastServerName); ok {
		p.broadcastPid = pid
		rigger.AddRoutee(ctx, pid, ctx.Self())
	}

	return nil
}

func (p *playerServer) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (p *playerServer) OnStopping(ctx actor.Context) {
}

func (p *playerServer) OnStopped(ctx actor.Context) {
	rigger.RemoveRoutee(ctx, p.broadcastPid, ctx.Self())
	fmt.Printf("now stopped!\r\n")
}

func (p *playerServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	switch msg := message.(type) {
	case *Broadcast:
		fmt.Printf("receive broadcast, content:%s\r\n", msg.Content)
	}
	return nil
}
