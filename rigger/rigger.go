package rigger

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

/*
根据函数启动一个进程, 并返回进程id
*/
func SpawnFromFun(parent actor.Context, fun actor.ReceiveFunc) *actor.PID {
	props := actor.PropsFromFunc(fun)
	return parent.Spawn(props)
}

/**
启动一个进程
 */
func SpawnFromProducer(parent actor.Context, producer actor.Producer) *actor.PID {
	props := actor.PropsFromProducer(producer)
	return parent.Spawn(props)
}

// 判断PID是否是本地PID
func IsLocalPid(pid *actor.PID) bool {
	return pid.Address == "nonhost"
}

func IsRemotePid(pid *actor.PID) bool {
	return !IsLocalPid(pid)
}

// 启动函数, go-rigger使用此类型的启动进程
type SpawnFun func(parent actor.SpawnerContext, props *actor.Props, args interface{}) (pid *actor.PID, err error)

// 表示不需要回复发送者,意味者,处理者需要自己正确处理回复
const NoReply = noReply(1)

// noReply,一种特殊的消息处理返回值,表示不需要rigger-go回复发送者,相反,需要用户自己进行回复
// 为了符合GeneralServerBehaviour.OnMessage的接口说明, 手动实现了,proto.Message
type noReply int
func (n noReply) Reset() {
}
func (n noReply) String() string {
	return "no reply"
}
func (n noReply) ProtoMessage() {
}

// 回复类型, 用于转发(Forward)中
type RespondType byte
const  (
	RespondNone   RespondType = 1 + iota // 不回复
	RespondOrigin                    // 回复给原始发送进程
	RespondSelf                      // 回复给本进程
)

// 转发结构体,如果处理完消息后返回此值,则会继续将消息转发给指定进程, 且,后续接收进程会根据Responded的值选择是否回复初始发送者
// 为了满足GeneralServerBehaviour.OnMessage的接口规范, 手动实现了proto.Message
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
// 如果进程设置了超时时间,则必须实现本接口,否则,触发超时时,会引发异常
// TODO 考虑在启动进程时进行断言
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
// 所有的go-rigger进程行为模式,都需要实现此接口
type LifeCyclePart interface {
	OnRestarting(ctx actor.Context)
	/*
	启动时的回调,应该在此回调中进行初始化,不管是正常启动或是重启,都会调用此事件
	当回调此接口时,意味着,本进程之前的所有进程都已经回调完成 OnStarted 接口
	如果返回的错误不为空,则认为进程初始化失败,此时:
	1. 监控进程会停止启动其后的所有进程
	2. 会马上停止初始化失败的进程
	*/
	OnStarted(ctx actor.Context, args interface{}) error
	/*
	初始化完成后执行,在调用前会先通知调用者初始化完成,建议将比较费时的初始化操作放在此回调中进行
	以防止初始化太久而导致超时, 对于Supervisro进程来说,调用此方法时,会保证所有子进程都已经启动完成
	*/
	OnPostStarted(ctx actor.Context, args interface{})
	/*
	进程即将停止时,进行回调
	*/
	OnStopping(ctx actor.Context)
	/*
	进程停止后进行回调
	*/
	OnStopped(ctx actor.Context)
}

// pid持有接口
//type PidHolder interface {
//	GetPid() *actor.pid
//	SetPid(pid *actor.pid)
//}

//type LifeCycleProducer func() LifeCyclePart

// 启动子进程的命令
type startChildCmd struct {
	spawnSpec *SpawnSpec
}

type spawnerSetter interface {
	setSpawner(spawner actor.SpawnerContext) spawnerSetter
}

type supervisorSetter interface {
	setSupervisor(strategy actor.SupervisorStrategy) supervisorSetter
}

// 根据spawner的类型,获取spawner
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


func registerNamedProcessMiddleware(next actor.SpawnFunc) actor.SpawnFunc {
	return func(actorSystem *actor.ActorSystem, id string, props *actor.Props, parentContext actor.SpawnerContext) (*actor.PID, error) {
		if pid, err := next(actorSystem, id, props, parentContext); err == nil {
			// 注册名字
			if nil != riggerProcessManagingServerPid {
				//fmt.Printf("treate register, kind: %s, name: %s \r\n", id, parseProcessName(id))
				name := parseProcessName(id)
				if name != "" {
					f := actorSystem.Root.RequestFuture(riggerProcessManagingServerPid, &registerNamedPid{
						name: name,
						pid:  pid,
					}, 3 * time.Second)
					if err := f.Wait(); err == nil {
						logrus.Tracef("success register")
					} else {
						logrus.Tracef("unsuccess register")
					}
				}
			}
			return pid, nil
		} else {
			return nil, err
		}
	}
}

func parseProcessName(name string) string {
	arr := []rune(name)
	length := len(arr)
	for i := length - 1; i >= 0; i -= 1 {
		if arr[i] == '/' {
			// 判断是否是'$'开头
			if arr[i + 1] != '$' {
				return name[i + 1:]
			} else {
				return ""
			}
		}
	}

	return name
}
