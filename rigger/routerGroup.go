package rigger

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/router"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"time"
)

// 将一个进程添加到路由里去
func AddRoutee(from actor.Context, routerPid *actor.PID, pid *actor.PID)  {
	from.Send(routerPid, &router.AddRoutee{PID: pid})
}

func RemoveRoutee(from actor.Context, routerPid *actor.PID, pid *actor.PID)  {
	from.Send(routerPid, &router.RemoveRoutee{PID: pid})
}

// 获取所有的Routee
func GetRoutees(from actor.Context, routerPid *actor.PID, timeout time.Duration) (*router.Routees, error) {
	if r, err := from.RequestFuture(routerPid, &router.GetRoutees{}, timeout).Result(); err != nil {
		return r.(*router.Routees), nil
	} else {
		return nil, err
	}
}

// 向所有拥有的进程广播消息,请注意: 如果有跨节点的进程请不要使用此接口,否则会导致运行时错误
// TODO 使可以跨进程
func Broadcast(from actor.Context, routerPid *actor.PID, msg proto.Message)  {
	from.Send(routerPid, &router.BroadcastMessage{Message: msg})
}

// 路由类型
type RouterType int
const (
	RandomType RouterType = 1 + iota // 随机选择一个进程,进行消息转发
	BroadcastType // 广播类型,即将消息转发给所有进程
	RoundRobinType // 轮询
	ConsistentHashType // 固定哈希, 即哈希值一样的消息始终在同一个进程处理
)

// router group行为模式
type RouterGroupBehaviour interface {
	// 获取类型
	OnGetType() RouterType
	// 获取初始的路由进程ID
	OnGetRoutee() []*actor.PID
}

type RouterGroupBehaviourProducer func() RouterGroupBehaviour

// 启动一个router group
func startRouterGroup(parent interface{}, id string) (*routerGroup, error) {
	if _, ok := getRegisterInfo(id); ok {
		group := &routerGroup{}
		_, err := group.WithSupervisor(parent).WithSpawner(parent).StartSpec(&SpawnSpec{
			Kind:         id,
			SpawnTimeout: startTimeOut,
		})
		if err != nil {
			return nil, err
		}
		return group, nil
	} else {
		return nil, ErrNotRegister(id)
	}
}

func startRouterGroupSpec(parent interface{}, spec *SpawnSpec) (*routerGroup, error) {
	group := &routerGroup{}
	return group.WithSupervisor(parent).WithSpawner(parent).StartSpec(spec)
}

type routerGroup struct {
	pid *actor.PID
	strategy actor.SupervisorStrategy
	spawner actor.SpawnerContext
}

func (r *routerGroup) setSupervisor(strategy actor.SupervisorStrategy) supervisorSetter {
	r.strategy = strategy
	return r
}

func (r *routerGroup) setSpawner(spawner actor.SpawnerContext) spawnerSetter {
	r.spawner = spawner
	return r
}

func (r *routerGroup) StartSpec(spec *SpawnSpec) (*routerGroup, error) {
	if info, ok := getRegisterInfo(spec.Kind); ok {
		switch prod := info.producer.(type) {
		case RouterGroupBehaviourProducer:
			be := prod()
			pids := be.OnGetRoutee()
			props := genProps(be.OnGetType(), pids)
			startFun := makeStartFun(spec, info)
			if pid, err := startFun(r.spawner, props, spec.Args); err != nil {
				log.Errorf("error when start actor, reason:%s", err.Error())
				return r, err
			} else {
				r.pid = pid
				return r, nil
			}
		default:
			return r, ErrWrongProducer(spec.Kind)
		}
	} else {
		return nil, ErrNotRegister(spec.Kind)
	}
}

// 添加监控,需要在Start之前执行,并且只能设置一次非空supervisor,如果重复设置,则简单忽略
func (r *routerGroup)WithSupervisor(maybeSupervisor interface{}) *routerGroup  {
	withSupervisor(r, maybeSupervisor)
	return r
}

func (r *routerGroup)WithSpawner(spawner interface{}) *routerGroup {
	withSpawner(r, spawner)
	return r
}

func genProps(t RouterType, pids []*actor.PID) *actor.Props {
	switch t {
	case RandomType:
		return router.NewRandomGroup(pids...)
	case BroadcastType:
		return router.NewBroadcastGroup(pids...)
	case RoundRobinType:
		return router.NewRoundRobinGroup(pids...)
	case ConsistentHashType:
		return router.NewConsistentHashGroup(pids...)
	}

	return nil
}

