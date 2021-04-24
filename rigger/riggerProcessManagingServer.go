package rigger

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
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
	//registeredProcess[riggerProcessManagingServerName] = ctx.Self()
	_ = registeredProcess.add(riggerProcessManagingServerName, ctx.Self(), true)
	riggerProcessManagingServerPid = ctx.Self()

	if globalManagerGatewayCli != nil {
		join(ctx)
	}
	root.Root.WithSpawnMiddleware(registerNamedProcessMiddlewareRoot)
	return nil
}

func (r *riggerProcessManagingServer) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (r *riggerProcessManagingServer) OnStopping(ctx actor.Context) {
}

func (r *riggerProcessManagingServer) OnStopped(ctx actor.Context) {
	if globalProcessManagingServerPid != nil {
		ctx.Unwatch(globalProcessManagingServerPid)
	}
}

func (r *riggerProcessManagingServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	switch msg := message.(type) {
	case *registerNamedPid:
		logrus.Warnf("register msg, name:%s", msg.name)
		if err := r.registerNamedProcess(ctx, msg); err != nil {
			return &Error{ErrStr: err.Error()}
		} else {
			return nil
		}
	case *getRemotePid:
		if pid, err := r.getRemotePid(ctx, msg.name); err == nil {
			return pid
		} else {
			return nil
		}
	case *actor.Terminated:
		r.onProcessDown(ctx, msg.Who)

	}
	return nil
}

// TODO 因为需要全局注册,所以注册有可能失败
func (r *riggerProcessManagingServer) registerNamedProcess(ctx actor.Context, info *registerNamedPid) error {
	fmt.Printf("real treate register kind: %s \r\n", info.name)
	if err := r.registerLocal(ctx, info.name, info.pid); err != nil {
		return err
	}

	// 注册全局进程
	if !isLocalName(info.name) {
		if _, err := globalManagerGatewayCli.Register(&RegisterGlobalProcessRequest{Name: info.name, Pid: info.pid}); err != nil {
			// TODO 错误处理
			logrus.Errorf("error when regisger global pid, name: %s", info.name)
			return err
		} else {
			logrus.Tracef("success register global pid: %s", info.name)
			return nil
		}
	} else {
		return nil
	}
}

func (r *riggerProcessManagingServer) registerLocal(ctx actor.Context, name string, pid *actor.PID) error {
	// TODO 因为使用名字启动时protoactor会检查名字唯一性,因此是否可以不再检查?
	if err := registeredProcess.add(name, pid, false); err == nil {
		ctx.Watch(pid)
		return nil
	} else {
		return err
	}
}

func (r *riggerProcessManagingServer) getRemotePid(ctx actor.Context, name string) (*actor.PID, error) {
	if resp, err := globalManagerGatewayCli.GetPid(&GetPidRequest{Name: name}); err == nil {
		pid := resp.Pid
		if err := r.registerLocal(ctx, name, pid); err == nil {
			return pid, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (r *riggerProcessManagingServer) onProcessDown(ctx actor.Context, pid *actor.PID){
	// TODO 优化
	// 是否是全局管理进程
	if globalProcessManagingServerPid != nil && globalProcessManagingServerPid.Address == pid.Address && globalProcessManagingServerPid.Id == pid.Id {
		globalManagerGatewayCli = nil
		globalProcessManagingServerPid = nil
		// 重新获取全局管理进程
		globalManagerGatewayCli = GetGlobalManagingGatewayGrainClient(clusterInstance, GlobalManagingGatewayKindName)
		join(ctx)
		return
	}

	name := parseProcessName(pid.Id)
	registeredProcess.remove(name)
}

