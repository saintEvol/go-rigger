package rigger

import (
	"errors"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/remote"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"time"
)

const riggerManagingServerName = "@riggerManagingServerName"
var riggerManagingServerPid *actor.PID = nil
var riggerProcessManagingServerPid *actor.PID = nil
// 已经注册的进程,此MAP只能由applicationTopSup在OnStarted时对自己进行注册时和riggerManagingServer进程对其它进程进行注册时修改
var registeredProcess = make(map[string]*actor.PID)

type registerNamedPid struct {
	name string
	pid *actor.PID
	//isGlobal bool
}

type getRemotePid struct {
	name string // 进程名
}

// 启动本地应用
func spawnLocalApplications(spec *SpawnLoacalApplicationSpec) (*SpawnLocalApplicationResp, error) {
	// 确保rigger管理服务启动了
	if riggerManagingServerPid != nil {
		if r, err := root.Root.RequestFuture(riggerManagingServerPid, spec, 10 * time.Second).Result(); err != nil {
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
	riggerManagingServerPid = ctx.Self()
	//registeredProcess[riggerManagingServerName] = ctx.Self()
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
		err := r.startApplicationRecursively(ctx, spec.ApplicationId, make(map[string]bool), startTimeOut)
		if err == nil {
			return &SpawnLocalApplicationResp{}
		} else {

			return &SpawnLocalApplicationResp{Error: err.Error()}
		}
	}
}

/*
递归的启动应用,也即同时启动该应用的所有依赖应用
启动过程中会检查应用是否存在循环依赖
*/
func (r *riggerManagingServer) startApplicationRecursively(ctx actor.Context, appOrSpec interface{},
	handling map[string]bool, timeout time.Duration) error {
	var app string
	var spec *SpawnSpec
	switch s := appOrSpec.(type) {
	case string:
		app = s
		spec = SpawnSpecWithKind(app)
	case *SpawnSpec:
		app = s.Kind
		spec = s
	}

	deps := getDependence(app)
	handling[app] = true
	for _, dep := range deps {
		// 检查是否循环依赖
		if _, exists := handling[dep]; app == dep || exists {
			return errors.New(fmt.Sprintf("find loop dependent: %s <<=>> %s", app, dep))
		}

		// 依赖的是否是应用
		if info, exists := getRegisterInfo(dep); exists {
			if _, ok := info.producer.(ApplicationBehaviourProducer); !ok {
				return errors.New(fmt.Sprintf("%s not an application", dep))
			}
		} else {
			return errors.New(fmt.Sprintf("has no register info: %s", dep))
		}

		if _, started := GetRunningApplication(dep); !started {
			// 依赖没启动,先启动依赖
			if err := r.startApplicationRecursively(ctx, dep, handling, timeout); err != nil {
				return err
			}
		}
	}
	delete(handling, app)
	// 所有依赖启动完成, 启动应用本身
	if pid, err := StartChildSync(ctx, r.topSupPid, spec, timeout); err == nil {
		setRunningApplication(app, pid)
		return nil
	} else {
		return err
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
	if err := r.startApplicationNode(ctx, node); err != nil {
		log.Panicf("error when start application: %s, error: %s", node.kind, err.Error())
	}
}

func (r *riggerManagingServer) startApplicationNode(ctx actor.Context, node *StartingNode) error {
	if node.parent != nil {
		log.Panicf("application should not have Parent:%s", node.kind)
	}
	spawnSpec := node.spawnSpec
	// producer先不判断了,因为生成时,已经判断过了
	// remote
	if node.remote != nil {
		// TODO 多应用时是否会重复启动remote
		re := remote.NewRemote(root, remote.Configure(node.remote.Host, node.remote.Port))
		re.Start()
	}
	return r.startApplicationRecursively(ctx, spawnSpec, make(map[string]bool), startTimeOut)
}

