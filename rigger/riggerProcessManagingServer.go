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
	//globalProcessManagingServerClient *GlobalProcessManagingServerGrainClient
	//globalProcessManagingServerPid *actor.PID // 全局管理进程的进程id
}

func (r *riggerProcessManagingServer) OnRestarting(ctx actor.Context) {
}

func (r *riggerProcessManagingServer) OnStarted(ctx actor.Context, args interface{}) error {
	root.Root.WithSpawnMiddleware(registerNamedProcessMiddleware)
	registeredProcess[riggerProcessManagingServerName] = ctx.Self()
	riggerProcessManagingServerPid = ctx.Self()

	if globalProcessManagingServerPid != nil {
		ctx.Watch(globalProcessManagingServerPid)
	}

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
		r.onProcessDown(msg.Who)

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
	if !belongThisNode(info.name) {
		if _, err :=  globalProcessManagingServerCli.Register(&RegisterGlobalProcessRequest{Name: info.name, Pid: info.pid}); err != nil {
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
	if _, exists := registeredProcess[name]; exists {
		return ErrLocalNameExists
	}

	registeredProcess[name] = pid
	ctx.Watch(pid)

	return nil
}

func (r *riggerProcessManagingServer) getRemotePid(ctx actor.Context, name string) (*actor.PID, error) {
	if resp, err := globalProcessManagingServerCli.GetPid(&GetPidRequest{Name: name}); err == nil {
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

func (r *riggerProcessManagingServer) onProcessDown(pid *actor.PID){
	// TODO 优化
	// 是否是全局管理进程
	if globalProcessManagingServerPid != nil && globalProcessManagingServerPid.Address == pid.Address && globalProcessManagingServerPid.Id == pid.Id {
		globalProcessManagingServerCli = nil
		globalProcessManagingServerPid = nil
		join()
		return
	}

	name := parseProcessName(pid.Id)
	delete(registeredProcess, name)
}

