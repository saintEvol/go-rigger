package rigger

import (
	"errors"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"reflect"
	"time"
)

/**
通用服务器
 */

// 通用服务器行为模式
type GeneralServerBehaviour interface {
	// 启动时的回调,应该在此回调中进行初始化,不管是正常启动或是重启,都会调用此事件
	LifeCyclePart
	// 结果需要返回给请求进程,为了保证能够跨节点,需要是proto.Message
	OnMessage(ctx actor.Context, message interface{}) proto.Message
}

// 行为模式生成器
type GeneralServerBehaviourProducer func() GeneralServerBehaviour

func startGeneralServer(parent interface{}, id string) (*GeneralServer, error) {
	if _, ok := getRegisterInfo(id); ok {
		server := newGeneralServer()
		_, err := server.WithSupervisor(parent).WithSpawner(parent).StartSpec(&SpawnSpec{
			Kind:         id,
			SpawnTimeout: startTimeOut,
		})
		if err != nil {
			return nil, err
		}
		return server, nil
	} else {
		return nil, ErrNotRegister(id)
	}
}

func startGeneralServerSpec(parent interface{}, spec *SpawnSpec) (*GeneralServer, error) {
	server := newGeneralServer()
	return server.WithSupervisor(parent).WithSpawner(parent).StartSpec(spec)
}

// 生成一个新的GeneralServer
func newGeneralServer() *GeneralServer  {
	server := &GeneralServer{}
	return server
}

// 通用服务器
type GeneralServer struct {
	pid            *actor.PID // 进程ID
	spawner        actor.SpawnerContext
	strategy       actor.SupervisorStrategy
	initArgs       interface{} // 初始化参数
	delegate       *genServerDelegate
	receiveTimeout time.Duration
}

// 添加监控,需要在Start之前执行,并且只能设置一次非空supervisor,如果重复设置,则简单忽略
func (server *GeneralServer)WithSupervisor(maybeSupervisor interface{}) *GeneralServer  {
	withSupervisor(server, maybeSupervisor)
	return server
}

func (server *GeneralServer)WithRawSupervisor(strategy actor.SupervisorStrategy) *GeneralServer  {
	server.strategy = strategy
	return server
}

func (server *GeneralServer)WithSpawner(spawner interface{}) *GeneralServer {
	withSpawner(server, spawner)
	return server
}

// 使用启动规范启动一个Actor
func (server *GeneralServer)StartSpec(spec *SpawnSpec) (*GeneralServer, error){
	if info, ok := getRegisterInfo(spec.Kind); ok {
		switch prod := info.producer.(type) {
		case GeneralServerBehaviourProducer:
			props, initFuture := server.prepareSpawn(prod, spec)
			// 检查startFun
			startFun := makeStartFun(spec, info)
			// 在启动完成前设置启动参数
			server.initArgs = spec.Args
			if pid, err := startFun(server.spawner, props, spec.Args); err != nil {
				log.Errorf("error when start general server, reason:%s", err.Error())
				server.initArgs = nil
				return server, err
			} else {
				// 设置receiveTimeout
				if spec.ReceiveTimeout >= 0 {
					server.receiveTimeout = spec.ReceiveTimeout
				} else {
					server.receiveTimeout = -1
				}
				// 等待
				if initFuture != nil {
					if ret, err := initFuture.Result(); err != nil {
						log.Errorf("error when wait start actor reason:%s", err)
						return server, err
					} else {
						r := ret.(*SpawnResponse)
						if r.Error == "" {
							server.pid = pid
							return server, nil
						} else {
							return server, errors.New(r.Error)
						}
					}
				}
				server.pid = pid
			}
		default:
			return server, ErrWrongProducer(reflect.TypeOf(prod).Name())
		}
		return server, nil

	} else {
		return nil, ErrNotRegister(spec.Kind)
	}
}

// Interface: Stoppable
func (server *GeneralServer) Stop() {
	server.spawner.ActorSystem().Root.Stop(server.pid)
}

func (server *GeneralServer) StopFuture() *actor.Future {
	return server.spawner.ActorSystem().Root.StopFuture(server.pid)
}

func (server *GeneralServer) Poison() {
	server.spawner.ActorSystem().Root.Poison(server.pid)
}

func (server *GeneralServer) PoisonFuture() *actor.Future {
	return server.spawner.ActorSystem().Root.PoisonFuture(server.pid)
}

// Interface: Sender
// 给genser发送一条消息
func (server *GeneralServer)Send(sender actor.SenderContext, msg interface{})  {
	sender.Send(server.pid, msg)
}

func (server *GeneralServer)Request(sender actor.SenderContext, msg interface{})  {
	sender.Request(server.pid, msg)
}

// 默认5秒超时
func (server *GeneralServer)RequestFutureDefault(sender actor.SenderContext, msg interface{}) *actor.Future {
	return server.RequestFuture(sender, msg, 5 * time.Second)
}

func (server *GeneralServer)RequestFuture(sender actor.SenderContext, msg interface{}, timeout time.Duration) *actor.Future {
	return sender.RequestFuture(server.pid, msg, timeout)
}

func (server *GeneralServer) generateProps(producer GeneralServerBehaviourProducer, future *actor.Future) *actor.Props  {
	props := actor.PropsFromProducer(func() actor.Actor {
		return &genServerDelegate{
			initFuture: future,
			callback: producer(),
			owner: server,
		}
	})

	// 是否需要监控
	if server.strategy != nil {
		props.WithSupervisor(server.strategy)
	}

	return props
}

func (server *GeneralServer) prepareSpawn(producer GeneralServerBehaviourProducer, spec *SpawnSpec) (*actor.Props, *actor.Future) {
	if server.spawner == nil {
		server.WithSpawner(actor.NewActorSystem().Root)
	}
	var initFuture *actor.Future
	timeout := spec.SpawnTimeout
	if timeout <= 0 {
		initFuture = nil
	} else {
		initFuture = actor.NewFuture(server.spawner.ActorSystem(), timeout)
	}
	props := server.generateProps(producer, initFuture).WithSpawnMiddleware(registerNamedProcessMiddleware)


	return props, initFuture
}

// Interface: SpawnerSetter
func (server *GeneralServer) setSpawner(spawner actor.SpawnerContext) spawnerSetter {
	server.spawner = spawner
	return server
}

// Interface: supervisorSetter
func (server *GeneralServer) setSupervisor(strategy actor.SupervisorStrategy) supervisorSetter {
	server.strategy = strategy
	return server
}

// GeneralServer代理
type genServerDelegate struct {
	onMessageFunc func(ctx actor.Context, msg interface{}) proto.Message
	behaviour iBehaviour
	callback GeneralServerBehaviour
	initFuture *actor.Future // 初始化future,如果设置了,初始化完成后通过future进行通知
	owner *GeneralServer
	context actor.Context
}

func (server *genServerDelegate) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		var err error = nil
		// TODO 如果是这里通知的应该是有异常了
		defer server.notifyInitComplete(context, err)

		server.decideOnMessageFun()

		// 设置GeneralServer中的代理指针
		server.owner.delegate = server
		//server.callback.SetPid(context.RespondSelf())
		// 设置超时
		if server.owner.receiveTimeout > 0 {
			context.SetReceiveTimeout(server.owner.receiveTimeout)
		}
		err = server.callback.OnStarted(context, server.owner.initArgs)
		if err == nil {
			// 初始化完成了,通知后,继续进行后面的初始化
			server.notifyInitComplete(context, nil)
			server.callback.OnPostStarted(context, server.owner.initArgs)
			server.context = context
		} else {
			server.notifyInitComplete(context, err)
			context.Stop(context.Self())
		}
	case *actor.Restarting:
		server.callback.OnRestarting(context)
	case *actor.Stopping:
		server.callback.OnStopping(context)
	case *actor.Stopped:
		server.callback.OnStopped(context)
		server.owner.delegate = nil
		server.owner = nil
		server.context = nil
	case *actor.ReceiveTimeout:
		re := server.callback.(TimeoutReceiver)
		re.OnTimeout(context)
	default:
		ret := server.onMessageFunc(context, msg)
		if ret != NoReply {
			switch r := ret.(type) {
			case *Forward:
				server.forward(context, r)
			default:
				if context.Sender() != nil {
					context.Respond(ret)
				}
			}
		}
	}
}

func (server *genServerDelegate) decideOnMessageFun()  {
	if b, ok := (interface{})(server.callback).(iBehaviour); ok {
		server.behaviour = b
		server.onMessageFunc = server.onMessageWithBehaviour
	} else {
		server.behaviour = nil
		server.onMessageFunc = server.onMessagePlain
	}
}

func (server *genServerDelegate) onMessagePlain(context actor.Context, msg interface{}) proto.Message {
	return server.callback.OnMessage(context, msg)
}

func (server *genServerDelegate) onMessageWithBehaviour(context actor.Context, msg interface{}) proto.Message {
	if ret, ok := server.behaviour.handleMessage(context, msg); ok {
		return ret
	} else {
		return server.onMessagePlain(context, msg)
	}
}

func (server *genServerDelegate) notifyInitComplete(context actor.Context, err error)  {
	if server.initFuture != nil {
		var errStr string
		if err != nil {
			errStr = err.Error()
		} else {
			errStr = ""
		}
		context.Send(server.initFuture.PID(), &SpawnResponse{
			Sender: context.Sender(),
			Parent: context.Parent(),
			Pid:    context.Self(),
			Error:  errStr,
		})
		server.initFuture = nil
	}
}

func (server *genServerDelegate) forward(context actor.Context, forwardInfo *Forward)  {
	if forwardInfo.To == nil {
		return
	}
	switch forwardInfo.RespondType {
	case RespondNone:
		context.Send(forwardInfo.To, forwardInfo.Message)
	case RespondOrigin: // 回复给原始进程
		context.RequestWithCustomSender(forwardInfo.To, forwardInfo.Message, context.Sender())
	case RespondSelf:
		context.Request(forwardInfo.To, forwardInfo.Message)
	}
}
