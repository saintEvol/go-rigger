package rigger

import (
	"errors"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/remote"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
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

var (
	//topSupPid = (*actor.PID)(nil) // 顶级监控进程, 监控rigger下所有的应用
)

/*
rigger 管理服务
*/
type riggerManagingServer struct {
	topSupPid *actor.PID
}

func (r *riggerManagingServer) OnRestarting(ctx actor.Context) {
}

func (r *riggerManagingServer) OnStarted(ctx actor.Context, args interface{}) error {
	if pid, exists := GetPid(allApplicationTopSupName); exists {
		r.topSupPid = pid
		return nil
	} else {
		return errors.New("top sup pid not exists")
	}
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
		return r.launchLoacalApplications(ctx, msg)
	}

	return nil
}

// 根据配置, 启动本地的应用
func (r *riggerManagingServer) launchLoacalApplications(ctx actor.Context, spec *SpawnLoacalApplicationSpec) proto.Message {
	if spec.LaunchConfigPath != "" {
		// 读取配置文件
		readLaunchConfig(spec.LaunchConfigPath)
		// 生成启动树
		parseConfig()
		printProcessTree(startingTasks)
		// 读取应用配置
		readAppConfig(spec.ApplicationConfigPath)
		r.startApplications(ctx)
		return &SpawnLocalApplicationResp{}
	} else {
		// 读取应用配置
		readAppConfig(spec.ApplicationConfigPath)
		if app, err := StartChildSync(ctx, r.topSupPid, DefaultSpawnSpec(spec.ApplicationId), startTimeOut); err == nil {
			setRunningApplication(spec.ApplicationId, app)
			return &SpawnLocalApplicationResp{}
		} else {
			return &SpawnLocalApplicationResp{Error: err.Error()}
		}
	}
}

func (r *riggerManagingServer) startApplications(ctx actor.Context)  {
	for _, node := range startingTasks {
		r.startNode(ctx, node)
	}
}

func (r *riggerManagingServer) startNode(ctx actor.Context, node *StartingNode) {
	if node.location != nil {
		return
	}

	// 启动应用
	if app, err := r.startApplicationNode(ctx, node); err == nil {
		// 将启动的应用存起来
		setRunningApplication(node.name, app)
		//startRest(app, filterLocalNode(node.children))
	} else {
		log.Panicf("error when start application: %s", node.name)
	}
}

func (r *riggerManagingServer) startApplicationNode(ctx actor.Context, node *StartingNode) (*actor.PID, error) {
	if node.parent != nil {
		log.Panicf("application should not have Parent:%s", node.name)
	}
	spawnSpec := node.spawnSpec
	// producer先不判断了,因为生成时,已经判断过了
	// remote
	if node.remote == nil {
		return StartChildSync(ctx, r.topSupPid, spawnSpec, startTimeOut)
		// return startApplicationSpec(spawnSpec)
	} else {
		re := remote.NewRemote(root, remote.Configure(node.remote.host, node.remote.port))
		re.Start()
		return StartChildSync(ctx, r.topSupPid, spawnSpec, startTimeOut)
	}
}

