package rigger

import (
	"errors"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/cluster"
	"github.com/sirupsen/logrus"
	"time"
)

const GlobalManagingGatewayKindName = "GlobalManagingGateway"

type globalManagerGatewyGrain struct {
	managerPid *actor.PID
}

const globalRequestTimeout = 10 * time.Second

func (g *globalManagerGatewyGrain) Init(id string) {
	props := actor.PropsFromProducer(func() actor.Actor {
		return &globalManager{}
	}).WithSupervisor(actor.NewOneForOneStrategy(10, 1 * time.Second, func(reason interface{}) actor.Directive {
		return actor.RestartDirective
	}))
	pid, _ := root.Root.SpawnNamed(props, globalManagerName)
	g.managerPid = pid
}

func (g *globalManagerGatewyGrain) Terminate() {
	if g.managerPid != nil {
		root.Root.Stop(g.managerPid)
	}
	logrus.Warn("global process managing server now terminate")
}

func (g *globalManagerGatewyGrain) ReceiveDefault(ctx actor.Context) {
	logrus.Warn("global process managing server now receive default")
}

func (g *globalManagerGatewyGrain) GetPid(request *GetPidRequest, context cluster.GrainContext) (*GetPidResponse, error) {
	f := context.RequestFuture(g.managerPid, &_getPid{name: request.Name}, globalRequestTimeout)
	if ret, err := f.Result(); err == nil {
		if ret == nil {
			return nil, ErrGlobalNameNotRegistered
		}
		return &GetPidResponse{Pid: ret.(*actor.PID)}, nil
	} else {
		return nil, err
	}
}

func (g *globalManagerGatewyGrain) Register(request *RegisterGlobalProcessRequest, context cluster.GrainContext) (*Noop, error) {
	f := context.RequestFuture(g.managerPid, &_register{
		name: request.Name,
		pid:  request.Pid,
	}, globalRequestTimeout)
	if ret, err := f.Result(); err == nil {
		if ret == nil {
			return nil, nil
		} else {
			return nil, ret.(error)
		}
	} else {
		return nil, err
	}
}

//TODO 暂未实现
func (g *globalManagerGatewyGrain) Reset(request *ResetRequest, context cluster.GrainContext) (*Noop, error) {
	for _, info := range request.Pids {
		if _, err := g.Register(info, context); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

//TODO 暂未实现
func (g *globalManagerGatewyGrain) SyncAllOthers(request *SyncAllOthersRequest, context cluster.GrainContext) (*GlobalProcessList, error) {
	return nil, nil
}

func (g *globalManagerGatewyGrain) Join(request *JoinRequest, context cluster.GrainContext) (*JoinResponse, error) {
	f := context.RequestFuture(g.managerPid, &_join{
		name: request.Node,
		pid:  request.Pid,
	}, globalRequestTimeout)
	if resp, err := f.Result(); err == nil {
		switch t := resp.(type) {
		case *actor.PID:
			return &JoinResponse{Pid: t}, nil
		case error:
			return nil, t
		default:
			return nil, errors.New("unknow error")
		}
	} else {
		return nil, err
	}
}


