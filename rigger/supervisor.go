package rigger

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	log "github.com/sirupsen/logrus"
	"reflect"
	"time"
)

// 监控进程行为模式接口
type SupervisorBehaviour interface {
	LifeCyclePart
	/*
	获取监控标志时的回调
	childSpecs: 子进程规范, 描述如何启动子进程, 可以为以下类型:
		1. *SpawnSpec
		2. string, 此时为子规范的ID, 监控进程会根据此id去配置中查找对应的启动规范
	*/
	OnGetSupFlag(ctx actor.Context) (supFlag SupervisorFlag, childSpecs []*SpawnSpec)
	//PidHolder
}

func NewSupervisor() *Supervisor {
	return &Supervisor{}
}

type SupervisorBehaviourProducer func() SupervisorBehaviour

// 策略常量
type StrategyFlag int

const (
	OneForOne StrategyFlag = iota // 只重启失败的进程
	AllForOne
	SimpleOneForOne // 只重启失败的进程,并且,此模式下,只允许动态启动子进程
)

// 监控标志
type SupervisorFlag struct {
	MaxRetries     int           // 最大重启次数
	WithinDuration time.Duration // 多久时间内重启
	Decider        actor.DeciderFunc
	StrategyFlag   StrategyFlag // 重启策略标志
}

// 启动一个子进程,只对Supervisor有效
// spawnSpecOrArgs: 可以接受的类型为*SpawnSpec 或其它任何类型
// 注意: 如果监控进程的模式是 SimpleOneForOne, 则spawnSpecOrArgs会当作动态启动参考原样回传给starter/OnStarted/OnPostStarted
func StartChild(from actor.Context, pid *actor.PID, spawnSpecOrArgs interface{}) error {
	// 判断是否是远程进程
	if IsLocalPid(pid) {
		from.Send(pid, &startChildCmd{specOrArgs: spawnSpecOrArgs})
	} else {
		// 如果是远程进程,先序列化spawnSpecOrArgs
		if specBytes, err := encodeMsg(spawnSpecOrArgs); err == nil {
			from.Send(pid, &RemoteStartChildCmd{SpecOrArgs: specBytes})
		} else {
			return err
		}
	}

	return nil
}

type UnexceptedStartResult string

func (u UnexceptedStartResult) Error() string {
	return "unexpected start result"
}

// 以异步的方式启动子进程,如果启动成功,会将新的进程ID通知给from进程
func StartChildNotified(from actor.Context, pid *actor.PID, spawnSpecOrArgs interface{}) error {
	// 判断是否是远程进程
	if IsLocalPid(pid) {
		from.Request(pid, &startChildCmd{specOrArgs: spawnSpecOrArgs})
	} else {
		// 如果是远程进程,先序列化spawnSpecOrArgs
		if specBytes, err := encodeMsg(spawnSpecOrArgs); err != nil {
			return err
		} else {
			from.Request(pid, &RemoteStartChildCmd{SpecOrArgs: specBytes})
		}
	}

	return nil
}

/*
启动子进程
已知问题: 如果在父进程内部启动子进程,会导致死锁超时
*/
func StartChildSync(from actor.Context/*谁请求*/, pid *actor.PID/*父进程ID*/, spawnSpecOrArgs interface{} /* *SpawnSpec 或启动参数 */, timeout time.Duration) (*actor.PID, error) {
	if pid == nil {
		return nil, ErrPidIsNil{}
	}

	// 是否是远程进程
	if IsLocalPid(pid) {
		future := from.RequestFuture(pid, &startChildCmd{specOrArgs: spawnSpecOrArgs}, timeout)
		return waitStartChildResp(future)
	} else {
		//远程进程,先序列化spawnSpecOrArgs
		if specBytes, err := encodeMsg(spawnSpecOrArgs); err != nil {
			return nil, err
		} else {
			future := from.RequestFuture(pid, &RemoteStartChildCmd{SpecOrArgs: specBytes}, timeout)
			return waitStartChildResp(future)
		}
	}
}

// 启动一个监控进程
func StartSupervisor(parent interface{}, id string) (*Supervisor, error) {
	if _, ok := getRegisterInfo(id); ok {
		server, err := NewSupervisor().WithSupervisor(parent).WithSpawner(parent).StartSpec(&SpawnSpec{
			Id:           id,
			SpawnTimeout: startTimeOut,
		})
		if err != nil {
			return nil, err
		} else {
			return server, nil
		}
	} else {
		return nil, ErrNotRegister(id)
	}
}

// Parent:*Application, *Supervisor, *Generalserver, actor.Context, *actor.ActorSystem
func StartSupervisorSpec(parent interface{}, spec *SpawnSpec) (*Supervisor, error) {
	return NewSupervisor().WithSupervisor(parent).WithSpawner(parent).StartSpec(spec)
}

// TODO 当前重启会有问题,重启后可能不会走回调
type Supervisor struct {
	pid      *actor.PID
	id       string
	spawner  actor.SpawnerContext // TODO 是否无用
	delegate *supDelegate

	childStrategy actor.SupervisorStrategy
	strategy      actor.SupervisorStrategy

	// SpawnSpec part
	initArgs       interface{}
	receiveTimeout time.Duration
	isFromConfig bool
	//id string // 内部标识,只有由config启动时才会设置,可以根据ID获取配置值
}

func (sup *Supervisor) StartSpec(spec *SpawnSpec) (*Supervisor, error) {
	if info, ok := getRegisterInfo(spec.Id); ok {
		switch prod := info.producer.(type) {
		case SupervisorBehaviourProducer:
			sup.id = spec.Id
			sup.isFromConfig = spec.isFromConfig
			//props, initFuture := sup.prepareSpawn(prod, specOrArgs.SpawnTimeout)
			props, initFuture := sup.prepareSpawn(prod, spec.SpawnTimeout)
			if spec.ReceiveTimeout <= 0 {
				sup.receiveTimeout = -1
			} else {
				sup.receiveTimeout = spec.ReceiveTimeout
			}
			sup.initArgs = spec.Args
			// 检查下是否有startFun
			startFun := makeStartFun(info)
			if pid, err := startFun(sup.spawner, props, spec.Args); err != nil {
				sup.initArgs = nil
				log.Errorf("error when start supervisor, reason:%s", err.Error())
				return sup, err
			} else {
				// 等待
				if initFuture != nil {
					if ret, err := initFuture.Result(); err != nil {
						log.Errorf("error when wait start supervisor reason:%s", err)
						return sup, err
					} else {
						r := ret.(*SpawnResponse)
						if r.Error == "" {
							sup.pid = pid
							return sup, nil
						} else {
							log.Errorf("error when wait start supervisor reason:%s", err)
							return sup, errors.New(r.Error)
						}
					}
				}
			}
		default:
			log.Errorf("wrong producer")
			return sup, ErrWrongProducer(reflect.TypeOf(prod).Name())
		}
	}

	return sup, nil
}

// 设置其监控者,只能设置一次,重复设置则简单忽略
func (sup *Supervisor) WithSupervisor(maybeSupervisor interface{}) *Supervisor {
	if sup.strategy != nil {
		return sup
	}
	withSupervisor(sup, maybeSupervisor)
	return sup
}

func (sup *Supervisor) WithSpawner(spawner interface{}) *Supervisor {
	withSpawner(sup, spawner)
	return sup
}

// Interface: Stoppable
func (sup *Supervisor) Stop() {
	sup.spawner.ActorSystem().Root.Stop(sup.pid)
}

func (sup *Supervisor) StopFuture() *actor.Future {
	return sup.spawner.ActorSystem().Root.StopFuture(sup.pid)
}

func (sup *Supervisor) Poison() {
	sup.spawner.ActorSystem().Root.Poison(sup.pid)
}

func (sup *Supervisor) PoisonFuture() *actor.Future {
	return sup.spawner.ActorSystem().Root.PoisonFuture(sup.pid)
}

// Interface: spawnerSetter
func (sup *Supervisor) setSpawner(spawner actor.SpawnerContext) spawnerSetter {
	sup.spawner = spawner
	return sup
}

// Interface: supervisorSetter
func (sup *Supervisor) setSupervisor(strategy actor.SupervisorStrategy) supervisorSetter {
	sup.strategy = strategy
	return sup
}

func (sup *Supervisor) generateProps(producer SupervisorBehaviourProducer, future *actor.Future) *actor.Props {
	props := actor.PropsFromProducer(func() actor.Actor {
		return &supDelegate{
			initFuture: future,
			callback:   producer(),
			owner:      sup,
		}
	})

	// 是否需要监控
	if sup.strategy != nil {
		props.WithSupervisor(sup.strategy)
	}

	return props
}

// private methods
func (sup *Supervisor) prepareSpawn(producer SupervisorBehaviourProducer, timeout time.Duration) (*actor.Props, *actor.Future) {
	if sup.spawner == nil {
		sup.WithSpawner(nil)
	}
	var future *actor.Future
	if timeout < 0 {
		future = nil
	} else {
		future = actor.NewFuture(sup.spawner.ActorSystem(), timeout)
	}
	props := sup.generateProps(producer, future)

	return props, future
}

// Interface: supDelegateHolder
func (sup *Supervisor) GetId() string {
	return sup.id
}

func (sup *Supervisor) IsFromConfig() bool {
	return sup.isFromConfig
}

func (sup *Supervisor) SetDelegate(delegate *supDelegate) {
	sup.delegate = delegate
}

func (sup *Supervisor) GetReceiveTimeout() time.Duration {
	return sup.receiveTimeout
}

func (sup *Supervisor) SetChildStrategy(strategy actor.SupervisorStrategy) {
	sup.childStrategy = strategy
}

func (sup *Supervisor) GetInitArgs() interface{} {
	return sup.initArgs
}

// actor代理
type supDelegate struct {
	owner          supDelegateHolder
	initFuture     *actor.Future
	callback       SupervisorBehaviour
	supervisorFlag *SupervisorFlag
	childSpecs     []*SpawnSpec
	context        actor.Context
}

// Interface: actor.Actor
func (sup *supDelegate) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		var err error = nil
		defer sup.notifyInitComplete(context, err)
		sup.context = context
		sup.owner.SetDelegate(sup)
		//sup.callback.SetPid(context.RespondSelf())
		// 先设置超时, 以便用户可以选择在OnStarted里覆盖
		receiveTimeout := sup.owner.GetReceiveTimeout()
		if receiveTimeout >= 0 {
			context.SetReceiveTimeout(receiveTimeout)
		}
		initArgs := sup.owner.GetInitArgs()
		err = sup.callback.OnStarted(context, initArgs)
		if err == nil {
			// 启动子进程
			flag, specs := sup.getSupFlag(context)
			sup.treateSupFlag(&flag, specs)
			// 初始化完成,通知后再继续进行后续的初始化
			sup.notifyInitComplete(context, nil)
			sup.callback.OnPostStarted(context, initArgs)
		} else {
			sup.notifyInitComplete(context, err)
			// 停止进程
			context.Stop(context.Self())
			//context.Stop(context.Self())
		}
	case *actor.Restarting:
		sup.callback.OnRestarting(context)
	case *actor.Stopping:
		sup.callback.OnStopping(context)
	case *actor.Stopped:
		sup.callback.OnStopped(context)
		sup.context = nil
		sup.owner.SetDelegate(nil)
		sup.owner = nil
	case *actor.ReceiveTimeout:
		re := sup.callback.(TimeoutReceiver)
		re.OnTimeout(context)
	case *startChildCmd:
		// 启动子进程的命令
		sup.startChild(context, msg.specOrArgs)
	case *RemoteStartChildCmd: // 远程启动子进程
		// 解码
		if data, err := decodeMsg(msg.SpecOrArgs); err == nil {
			sup.startChild(context, data)
		}
	case *actor.Terminated:

	default:
		log.Errorf("unexpected msg:%s", msg)
	}
}

func (sup *supDelegate) notifyInitComplete(context actor.Context, err error) {
	var errStr string
	if err == nil {
		errStr = ""
	} else {
		errStr = err.Error()
	}
	if sup.initFuture != nil {
		context.Send(sup.initFuture.PID(), &SpawnResponse{
			Sender: context.Sender(),
			Parent: context.Parent(),
			Pid: context.Self(),
			Error: errStr,
		})
		sup.initFuture = nil
	}
}

// 根据监控模式及子进程规范,启动子进程, 在些函数中启动子进程时,均为同步启动,默认超时为10S
func (sup *supDelegate) treateSupFlag(supFlag *SupervisorFlag, childSpecs []*SpawnSpec) {
	switch supFlag.StrategyFlag {
	case OneForOne:
		sup.owner.SetChildStrategy(actor.NewOneForOneStrategy(supFlag.MaxRetries, supFlag.WithinDuration, supFlag.Decider))
	case AllForOne:
		sup.owner.SetChildStrategy(actor.NewAllForOneStrategy(supFlag.MaxRetries, supFlag.WithinDuration, supFlag.Decider))
	case SimpleOneForOne:
		// 此模式下,childSpecs有且只能有一个成员
		if l := len(childSpecs); l != 1 {
			log.Panic(fmt.Sprintf("%s:%s", "SimpleOneForOne Can Have And Must Have One Child Spec, Now Len:", fmt.Sprint(l)))
		}
		sup.owner.SetChildStrategy(actor.NewOneForOneStrategy(supFlag.MaxRetries, supFlag.WithinDuration, supFlag.Decider))
		// 此模式下所有子进程均为动态创建
	}

	// 存储下
	sup.supervisorFlag = supFlag
	sup.childSpecs = childSpecs

	// 生成节点配置信息
	sup.treateConfig(supFlag)

	// 只启动非动态进程
	if supFlag.StrategyFlag != SimpleOneForOne {
		// SimpleOneForOne模式下所有进程均为动态创建
		// 依次启动所有子进程
		sup.spawnSpecs(sup.childSpecs)
	}
}

func (sup *supDelegate) treateConfig(supFlag *SupervisorFlag) {
	if sup.owner.IsFromConfig() {
		return
	}

	// 生成节点配置信息
	if config, err := sup.completeConfig(supFlag); err == nil {
		// 先生成子进程的配置信息
		sup.initChildrenConfigs(config, sup.childSpecs)
	} else {
		log.Panicf("error when complete config, reason: %s", err.Error())
	}
}

func (sup *supDelegate) completeConfig(supFlag *SupervisorFlag) (*StartingNode, error) {
	id := sup.owner.GetId()
	if node, ok := getConfigByName(id); ok {
		node.supFlag = supFlag
		return node, nil
	} else {
		return nil, ErrNoParentNode(sup.owner.GetId())
	}
}

func (sup *supDelegate) initChildConfig(parent *StartingNode, spec *SpawnSpec) {
	// 不重复生成
	if _, ok := getConfigByName(spec.Id); ok {
		return
	}
	config := &StartingNode{
		name:      spec.Id,
		fullName:  fmt.Sprintf("%s/%s", parent.fullName, spec.Id),
		parent:    parent,
		spawnSpec: spec,
		children:  nil, // 启动子进程时再生成
		location:  nil, // 非配置启动,都为local进程
		remote:    nil, // 只有Application才有
		supFlag:   nil, // 启动子进程时再生成
	}
	// TODO  特定情形下是否有可能造成竞争
	parent.children = append(parent.children, config)
	setConfig(config)

}

func (sup *supDelegate) initChildrenConfigs(parent *StartingNode, specs []*SpawnSpec) {
	for _, spec := range specs {
		sup.initChildConfig(parent, spec)
	}
}

func (sup *supDelegate) startChild(context actor.Context, specOrArgs interface{}) {
	if sup.supervisorFlag.StrategyFlag == SimpleOneForOne {
		// 此模式下,直接取原来的Spcs, 并使用新传入的参数
		pid, err := sup.spawnBySpec(&SpawnSpec{
			Id:           sup.childSpecs[0].Id,
			Args:         specOrArgs, // SimpleOneForOne模式下,是参数
			SpawnTimeout: sup.childSpecs[0].SpawnTimeout,
		})
		sup.responseStartChild(context, pid, err)
	} else {
		spec :=  unifySpawnSpec(specOrArgs)
		// 初始化子进程的配置
		node, _ := getConfigByName(sup.owner.GetId())
		sup.initChildrenConfigs(node, []*SpawnSpec{spec})
		pid, err := sup.spawnBySpec(spec)
		sup.responseStartChild(context, pid, err)
	}
}

// 启动子进程规范中的进程
func (sup *supDelegate) spawnSpecs(childSpecs []*SpawnSpec) {
	// 依次启动所有子进程
	for _, childSpec := range childSpecs {
		if _, err := sup.spawnBySpec(childSpec); err != nil {
			// 如果出现错误,直接跳出
			// TODO 也许需要抛出错误,但同时不应该使用sup本身崩溃
			log.Fatalf("error when start specs, error: %s", err)
			return
		}
	}
}

// 根据spec启动子进程,所有启动都采用同步方式,默认超时时间:10S
func (sup *supDelegate) spawnBySpec(spec *SpawnSpec) (*actor.PID, error) {
	if info, ok := getRegisterInfo(spec.Id); ok {
		// 由监控树启动的进程,如果未设置超时,则强行设置为10S
		if spec.SpawnTimeout == 0 {
			spec.SpawnTimeout = startTimeOut
		}

		switch producer := info.producer.(type) {
		case GeneralServerBehaviourProducer:
			if server, err := startGeneralServerSpec(sup.owner, spec); err != nil {
				return nil, err
			} else {
				// 如果不是从配置启动,且不是sample设置进程ID
				return server.pid, nil
			}
		case SupervisorBehaviourProducer:
			if childSup, err := StartSupervisorSpec(sup.owner, spec); err != nil {
				return nil, err
			} else {
				return childSup.pid, nil
			}
		case RouterGroupBehaviourProducer:
			if group, err := startRouterGroupSpec(sup.owner, spec); err != nil {
				return nil, err
			} else {
				return group.pid, nil
			}
		case ApplicationBehaviourProducer:
			// TODO 应用特殊处理,应用的父进程只能为rigger内部的超级监控进程
			if sup.owner.GetId() ==  allApplicationTopSupName{
				if app, err := startApplicationSpec(sup.context, spec); err == nil {
					return app.pid, err
				} else {
					return nil, err
				}
			} else {
				log.Error("currently application can only be launched by rigger super supervisor")
				return nil, ErrUnSurportedProducer("currently application can only be launched by rigger super supervisor")
			}
		default:
			typeName := reflect.TypeOf(producer).Name()
			log.Errorf("currently still not support producer:%s", typeName)
			return nil, ErrUnSurportedProducer(typeName)
		}

	} else {
		return nil, ErrNotRegister(spec.Id)
	}
}

func (sup *supDelegate) getSupFlag(context actor.Context) (supFlag SupervisorFlag, childSpecs []*SpawnSpec) {
	if sup.owner.IsFromConfig() {
		// 从配置启动,所以要从配置中拿
		// 根据ID拿取自身的信息
		if config, ok := getConfigByName(sup.owner.GetId()); ok {
			supFlag = *config.supFlag
			// 所有的子进程的SpawnSpec
			localChildren := filterLocalNode(config.children)
			for _, spec := range localChildren {
				childSpecs = append(childSpecs, spec.spawnSpec)
			}
		} else {
			log.Panicf("failed to find config of %s", sup.owner.GetId())
		}
		return
	} else {
		return sup.callback.OnGetSupFlag(context)
	}
}

func (sup *supDelegate) responseStartChild(context actor.Context, pid *actor.PID, err error) {
	sender := context.Sender()
	if sender != nil {
		// 有发送者,进行回应
		log.Tracef("send back spawn resp: %v", sender)
		var errStr string
		if err != nil {
			errStr = err.Error()
		}
		context.Respond(&SpawnResponse{Sender: sender, Parent: context.Self(), Pid: pid, Error: errStr})
	}
}

func unifySpawnSpecs(maybeSpecs []*SpawnSpec) []*SpawnSpec {
	var ret = make([]*SpawnSpec, len(maybeSpecs))
	for idx, spec := range maybeSpecs {
		ret[idx] = unifySpawnSpec(spec)
	}

	return ret
}

// 将其它类型归整化为*SpawnSpec
func unifySpawnSpec(specOrOtherThing interface{}) *SpawnSpec {
	switch spec := specOrOtherThing.(type) {
	case *SpawnSpec:
		return spec
	case string:
		return makeDefaultSpawnSpec(spec)
	default:
		log.Errorf("unexcepted spawn specOrArgs:%s", reflect.TypeOf(spec).Name())
	}

	return nil
}

func makeDefaultSpawnSpec(id string) *SpawnSpec {
	return &SpawnSpec{
		Id:           id,
		SpawnTimeout: startTimeOut,
	}
}

// 可以持有一个supDelegate
type supDelegateHolder interface {
	GetId() string
	IsFromConfig() bool
	SetDelegate(delegate *supDelegate)
	GetReceiveTimeout() time.Duration
	SetChildStrategy(strategy actor.SupervisorStrategy)
	GetInitArgs() interface{}
}

//type startChildCmd struct {
//	specOrArgs interface{}
//}
// 将启动规范或参数包装在切片里,序列化,以便后续跨节点发送
func encodeMsg(argsOrSpawnSpec interface{}) ([]byte, error) {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	arr := []interface{}{argsOrSpawnSpec}
	if err := encoder.Encode(arr); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func decodeMsg(b []byte) (interface{}, error) {
	decoder := gob.NewDecoder(bytes.NewReader(b))
	arr := make([]interface{}, 1, 1)
	if err := decoder.Decode(&arr); err != nil {
		return nil, err
	} else if len(arr) != 1 {
		return nil, ErrSerializedSlicLenWrong{}
	} else {
		return arr[0], nil
	}
}

// 等等子进程启动结果
func waitStartChildResp(future *actor.Future) (*actor.PID, error) {
	if ret, err := future.Result(); err != nil {
		return nil, err
	} else {
		switch resp := ret.(type) {
		case *SpawnResponse:
			var err error
			if resp.Error != "" {
				err = ErrSpawn(resp.Error)
			}
			return resp.Pid, err
		default:
			return nil, UnexceptedStartResult("")
		}
	}
}

