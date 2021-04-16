package rigger

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
)

const riggerProcessManagingServerName = "riggerProcessManagingServer"

func init() {
	Register(riggerProcessManagingServerName, GeneralServerBehaviourProducer(func() GeneralServerBehaviour {
		return &riggerProcessManagingServer{}
	}))
}
type riggerProcessManagingServer struct {

}

func (r *riggerProcessManagingServer) OnRestarting(ctx actor.Context) {
}

func (r *riggerProcessManagingServer) OnStarted(ctx actor.Context, args interface{}) error {
	root.Root.WithSpawnMiddleware(registerNamedProcessMiddleware)
	registeredProcess[riggerProcessManagingServerName] = ctx.Self()
	riggerProcessManagingServerPid = ctx.Self()
	return nil
}

func (r *riggerProcessManagingServer) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (r *riggerProcessManagingServer) OnStopping(ctx actor.Context) {
}

func (r *riggerProcessManagingServer) OnStopped(ctx actor.Context) {
}

func (r *riggerProcessManagingServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	switch msg := message.(type) {
	case *registerNamedPid:
		r.registerNamedProcess(ctx, msg)
		ctx.Respond(nil)
	case *actor.Terminated:
		r.onProcessDown(msg.Who)

	}
	return nil
}

func (r *riggerProcessManagingServer) registerNamedProcess(ctx actor.Context, info *registerNamedPid)  {
	fmt.Printf("real treate register kind: %s \r\n", info.name)
	registeredProcess[info.name] = info.pid
	ctx.Watch(info.pid)
}

func (r *riggerProcessManagingServer) onProcessDown(pid *actor.PID){
	name := parseProcessName(pid.Id)
	delete(registeredProcess, name)
}
