package rigger

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	log "github.com/sirupsen/logrus"
	"reflect"
	"time"
)

/**
应用的行为模式模块
 */
type ApplicationBehaviour interface {
	SupervisorBehaviour
	//LifeCyclePart
	//PidHolder
	//OnMessage(msg interface{}, ctx actor.Context)
}

type ApplicationBehaviourProducer func() ApplicationBehaviour

// 根据应用ID启动应用
func startApplication(id string) (*Application, error) {
	return startApplicationSpec(makeDefaultSpawnSpec(id))
}

// 根据启动规范启动应用
func startApplicationSpec(spec *SpawnSpec)	(*Application, error)  {
	return (&Application{}).StartSpec(spec)
}

// 根据启动规范和system启动应用
func startApplicationWithSystem(system *actor.ActorSystem, spec *SpawnSpec) (*Application, error) {
	return (&Application{Parent: system}).StartSpec(spec)
}

// 应用
// ----- Application ------
type Application struct {
	id             string
	pid            *actor.PID
	delegate       *supDelegate
	Parent         *actor.ActorSystem
	initArgs       interface{}
	receiveTimeout time.Duration
	childStrategy  actor.SupervisorStrategy
}

func (app *Application)Start(producer ApplicationBehaviourProducer) *Application  {
	app.Parent = actor.NewActorSystem()
	cb := producer()
	props := actor.PropsFromProducer(func() actor.Actor {
		return &supDelegate{
			owner: app,
			callback: cb,
		}
	})
	app.pid = app.Parent.Root.Spawn(props)
	//cb.SetPid(app.pid)

	return app
}

func (app *Application)StartSpec(spec *SpawnSpec) (*Application, error) {
	if info, ok := getRegisterInfo(spec.Id); ok {
		switch prod := info.producer.(type) {
		case ApplicationBehaviourProducer:
			app.initConfig(spec)
			app.id = spec.Id
			if app.Parent == nil {
				app.Parent = actor.NewActorSystem()
			}
			// 准备启动, 会准备好props, 初始化future等
			props, initFuture := app.prepareSpawn(prod, spec.SpawnTimeout)
			if spec.ReceiveTimeout <= 0 {
				app.receiveTimeout = -1
			} else {
				app.receiveTimeout = spec.ReceiveTimeout
			}
			app.initArgs = spec.Args
			// 检查startFun
			startFun := makeStartFun(info)
			if pid, err := startFun(app.Parent.Root, props, spec.Args); err != nil {
				log.Errorf("error when start actor, reason:%s", err.Error())
				return app, err
			} else {
				app.pid = pid
				// application不设置receive timeout了
				if initFuture != nil {
					if err = initFuture.Wait(); err != nil {
						log.Errorf("error when wait start actor reason:%s", err)
						return app, err
					}
				}
			}
		default:
			log.Errorf("wrong producer")
			return app, ErrWrongProducer(reflect.TypeOf(prod).Name())
		}
	} else {
		return nil, ErrNotRegister(spec.Id)
	}
	return app, nil
}

func (app *Application ) prepareSpawn(producer ApplicationBehaviourProducer, timeout time.Duration) (*actor.Props, *actor.Future)  {
	var future *actor.Future
	if timeout < 0 {
		future = nil
	} else {
		future = actor.NewFuture(app.Parent, timeout)
	}

	props := app.generateProps(producer, future)

	return props, future
}
func (app *Application) generateProps(producer ApplicationBehaviourProducer, future *actor.Future) *actor.Props  {
	props := actor.PropsFromProducer(func() actor.Actor {
		return &supDelegate{
			initFuture: future,
			callback: producer(),
			owner: app,
		}
	})
	// application是最上层的,没有监控
	return props
}

func (app *Application) Stop()  {
	app.Parent.Root.Stop(app.pid)
}

func (app *Application)StopFuture() *actor.Future {
	return app.Parent.Root.StopFuture(app.pid)
}

func (app *Application)Poison() {
	app.Parent.Root.Poison(app.pid)
}

func (app *Application)PoisonFuture() *actor.Future  {
	return app.Parent.Root.PoisonFuture(app.pid)
}

// Interface: supDelegateHolder

func (app *Application) GetId() string {
	return app.id
}

func (app *Application) SetDelegate(delegate *supDelegate) {
	app.delegate = delegate
}

func (app *Application) GetReceiveTimeout() time.Duration {
	return app.receiveTimeout
}

func (app *Application) SetChildStrategy(strategy actor.SupervisorStrategy) {
	app.childStrategy = strategy
}

func (app *Application) GetInitArgs() interface{} {
	return app.initArgs
}

// private
func (app *Application) initConfig(spec *SpawnSpec)  {
	if isFromConfig {
		return
	}

	if _, ok := getConfigByName(spec.Id); ok {
		return
	}
	// 初始化配置
	config := &StartingNode{
		name:      spec.Id,
		fullName:  spec.Id,
		parent:    nil,
		spawnSpec: spec,
		children:  nil, // 启动子进程时再生成
		location:  nil,  // 非配置启动,都为local进程
		remote:    nil, // 只有Application才有
		supFlag:   nil, // 启动子进程时再生成
	}
	setConfig(config)
}
// actor代理
//type appDelegate struct {
//	owner *Application
//	context actor.Context
//	callback ApplicationBehaviour
//	initFuture *actor.Future
//}

//func (app *appDelegate)Receive(context actor.Context)  {
//	switch msg := context.Message().(type) {
//	case *actor.Started:
//		defer app.notifyInitComplete(context)
//
//		app.context = context
//		app.owner.delegate = app
//		app.callback.OnStarted(context, app.owner.initArgs)
//		app.notifyInitComplete(context)
//		app.callback.OnPostStarted(context, app.owner.initArgs)
//	case *actor.Restarting:
//		app.callback.OnRestarting(context)
//	case *actor.Stopping:
//		app.callback.OnStopping(context)
//	case *actor.Stopped:
//		app.callback.OnStopped(context)
//	default:
//		fmt.Println("unexpected msg:", msg)
//	}
//
//}
//
//func (app *appDelegate)notifyInitComplete(context actor.Context)  {
//	if app.initFuture != nil {
//		context.Send(app.initFuture.pid(), context.RespondSelf())
//		app.initFuture = nil
//	}
//}

