package rigger

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"time"
)
const riggerManagingServerName = "@riggerManagingServerName"

func spawnLocalApplications(spec *SpawnLoacalApplicationSpec) (*SpawnLocalApplicationResp, error) {
	if pid, ok := GetPid(riggerManagingServerName); ok {
		if r, err := root.Root.RequestFuture(pid, spec, 10 * time.Second).Result(); err != nil {
			return nil, err
		} else {
			return r.(*SpawnLocalApplicationResp), nil
		}
	} else {
		return nil, ErrNotRegister(riggerManagingServerName)
	}
}

func init() {
	Register(riggerManagingServerName, GeneralServerBehaviourProducer(func() GeneralServerBehaviour {
		return &riggerManagingServer{}
	}))
}

type riggerManagingServer struct {

}

func (r riggerManagingServer) OnRestarting(ctx actor.Context) {
}

func (r riggerManagingServer) OnStarted(ctx actor.Context, args interface{}) {
}

func (r riggerManagingServer) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (r riggerManagingServer) OnStopping(ctx actor.Context) {
}

func (r riggerManagingServer) OnStopped(ctx actor.Context) {
}

func (r riggerManagingServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	switch msg := message.(type) {
	case *SpawnLoacalApplicationSpec:
		if msg.LaunchConfigPath != "" {
			// 读取配置文件
			readLaunchConfig(msg.LaunchConfigPath)
			// 生成启动树
			parseConfig()
			printProcessTree(startingTasks)
			// 读取应用配置
			readAppConfig(msg.ApplicationConfigPath)
			startApplications()
			return &SpawnLocalApplicationResp{}
		} else {
			// 读取应用配置
			readAppConfig(msg.ApplicationConfigPath)
			if app, err := startApplication(msg.ApplicationId); err == nil {
				setRunningApplication(msg.ApplicationId, app)
				return &SpawnLocalApplicationResp{}
			} else {
				return &SpawnLocalApplicationResp{Error: err.Error()}
			}
		}
	}

	return nil
}
