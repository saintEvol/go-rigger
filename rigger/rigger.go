package rigger

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"reflect"
	"time"
)

// 判断PID是否是本地PID
func IsLocalPid(pid *actor.PID) bool {
	return pid.Address == "nonhost"
}

func IsRemotePid(pid *actor.PID) bool {
	return !IsLocalPid(pid)
}

// 启动子进程的命令
type StartChildCmd struct {
	specOrArgs interface{}
}

type SpawnFun func(parent actor.SpawnerContext, props *actor.Props, args interface{}) (pid *actor.PID, err error)
// 启动规范
type SpawnSpec struct {
	Id             string // Id, 框架会根据此ID查询启动 Producer和StartFun
	//Producer     interface{} // props producer,生成一个对应行为模式的实例,也即行为模式工厂
	//Starter      SpawnFun
	Args         interface{}   // 启动参数, 原样传入 Starter, 与LifeCyclePart.OnStareted, 对于SimpleOneForOne, 此字段不生效
	SpawnTimeout time.Duration // 超时时间,如果为0表示不等待,也即异步启动
	// 平静期超时时间,如果在指定时间内没收到任何消息,则会触发TimeroutReceiver回调,此值不为0时,需要实现TimeoutReceiver
	ReceiveTimeout time.Duration
}

func NewDefaultSpawnSpec() *SpawnSpec {
	return &SpawnSpec{
		//Producer:       nil,
		//Starter:        nil,
		Args:           nil,
		SpawnTimeout:   startTimeOut,
		ReceiveTimeout: 0,
		Id:             "",
	}
}

// spawn(子)进程的回应,目前只有利用监控进程手动启动子进程时,会有此回复
// 改成通过prtobuf定义
//type SpawnResponse struct {
//	Err    error
//	Sender *actor.PID // 是谁发起的请求
//	Parent *actor.PID // 父进程ID
//	Pid    *actor.PID //新进程的PID
//}

type noReply int

func (n noReply) Reset() {
}

func (n noReply) String() string {
	return "no reply"
}

func (n noReply) ProtoMessage() {
}

// 表示不需要回复发送者,意味者,处理者需要自己正确处理回复
const NoReply = noReply(1)

type RespondType byte

const  (
	RespondNone   RespondType = iota // 不回复
	RespondOrigin                    // 回复给原始发送进程
	RespondSelf                      // 回复给本进程
)

// 转发结构体,如果处理完消息后返回此值,则会继续将消息转发给指定进程, 且,后续接收进程会根据Responded的值选择是否回复初始发送者
type Forward struct {
	To *actor.PID // 转发给谁
	Message proto.Message // 需要转发的消息
	RespondType
}

func (f Forward) Reset() {
}

func (f Forward) String() string {
	return ""
}

func (f Forward) ProtoMessage() {
}

// 如果起动进程时,ReceiveTimeout为大于0的值,则超时后会触发此回调
type TimeoutReceiver interface {
	OnTimeout(ctx actor.Context)
}

type Stoppable interface {
	Stop()
	StopFuture() *actor.Future
	Poison()
	PoisonFuture() *actor.Future
}

// actor 生命周期接口
type LifeCyclePart interface {
	OnRestarting(ctx actor.Context)
	// 启动时的回调,应该在此回调中进行初始化,不管是正常启动或是重启,都会调用此事件
	OnStarted(ctx actor.Context, args interface{})
	// 初始化完成后执行,在调用前会先通知调用者初始化完成,建议将比较费时的初始化操作放在此回调中进行
	// 以防止初始化太久而导致超时, 对于Supervisro进程来说,调用此方法时,会保证所有子进程都已经启动完成
	OnPostStarted(ctx actor.Context, args interface{})
	OnStopping(ctx actor.Context)
	OnStopped(ctx actor.Context)
}

// pid持有接口
type PidHolder interface {
	GetPid() *actor.PID
	SetPid(pid *actor.PID)
}

type LifeCycleProducer func() LifeCyclePart

type spawnerSetter interface {
	setSpawner(spawner actor.SpawnerContext) spawnerSetter
}

type supervisorSetter interface {
	setSupervisor(strategy actor.SupervisorStrategy) supervisorSetter
}

func fetchSpawner(spawner interface{}) actor.SpawnerContext {

	var result actor.SpawnerContext

	if spawner == nil {
		return actor.NewActorSystem().Root
	}

	switch sp := spawner.(type) {
	case *actor.ActorSystem:
		result = sp.Root
	case actor.SpawnerContext:
		result = sp
	case *Application:
		result = sp.delegate.context
	case *Supervisor:
		result = sp.delegate.context
	case *GeneralServer:
		result = sp.delegate.context
	default:
		panic("valid spawner should be: nil, *actor.ActorSystem, actor.SpawnerContext, *Application, *Supervisor, *GeneralServer, now:" + reflect.TypeOf(spawner).Name())
	}

	return result
}

func withSpawner(setter spawnerSetter, spawner interface{}) spawnerSetter {
	sp := fetchSpawner(spawner)
	return setter.setSpawner(sp)
}

func withSupervisor(setter supervisorSetter, maybeSupervisor interface{}) supervisorSetter  {
	if maybeSupervisor == nil {
		return setter
	}
	switch sup := maybeSupervisor.(type) {
	case *Supervisor:
		setter.setSupervisor(sup.childStrategy)

	}

	return setter
}
