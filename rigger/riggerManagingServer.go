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

/*
rigger 管理服务
*/
type riggerManagingServer struct {

}

func (r *riggerManagingServer) OnRestarting(ctx actor.Context) {
}

func (r *riggerManagingServer) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (r *riggerManagingServer) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (r *riggerManagingServer) OnStopping(ctx actor.Context) {
}

func (r *riggerManagingServer) OnStopped(ctx actor.Context) {
}

func (r *riggerManagingServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	switch msg := message.(type) {
	case *SpawnLoacalApplicationSpec:
		return r.launchLoacalApplications(msg)
	}

	return nil
}

// 根据配置, 启动本地的应用
func (r *riggerManagingServer) launchLoacalApplications(spec *SpawnLoacalApplicationSpec) proto.Message {
	if spec.LaunchConfigPath != "" {
		// 读取配置文件
		readLaunchConfig(spec.LaunchConfigPath)
		// 生成启动树
		parseConfig()
		printProcessTree(startingTasks)
		// 读取应用配置
		readAppConfig(spec.ApplicationConfigPath)
		startApplications()
		return &SpawnLocalApplicationResp{}
	} else {
		// 读取应用配置
		readAppConfig(spec.ApplicationConfigPath)
		if app, err := startApplication(spec.ApplicationId); err == nil {
			setRunningApplication(spec.ApplicationId, app)
			return &SpawnLocalApplicationResp{}
		} else {
			return &SpawnLocalApplicationResp{Error: err.Error()}
		}
	}
}