package rigger

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/cluster"
	"github.com/sirupsen/logrus"
)

const GlobalProcessManagingServerKindName = "GlobalProcessManagingServer"

type globalProcessManagingServerGrain struct {
	processes map[string]*actor.PID // 所有注册的全局进程, name => *pid
	processes2Names map[string]string // 进程键到进程名的映射 key => name
	nodesProcesses     map[string]*actor.PID // 节点进程: 节点名 => *pid
	//nodes map[string]string // 节点: name => 地址
	address2Node map[string]string // 地址 => name
}

func (g *globalProcessManagingServerGrain) Init(id string) {
	g.processes = make(map[string]*actor.PID)
	//g.nodes = make(map[string]string)
	g.processes2Names = make(map[string]string)
	g.nodesProcesses = make(map[string]*actor.PID)
	g.address2Node = make(map[string]string)
}

func (g *globalProcessManagingServerGrain) Terminate() {
	logrus.Warn("global process managing server now terminate")
}

func (g *globalProcessManagingServerGrain) ReceiveDefault(ctx actor.Context) {
	logrus.Warn("global process managing server now receive default")
	switch msg := ctx.Message().(type) {
	case *actor.Terminated:
		g.onProcessDown(msg.Who)
	}
}

func (g *globalProcessManagingServerGrain) GetPid(request *GetPidRequest, context cluster.GrainContext) (*GetPidResponse, error) {
	if pid, exists := g.processes[request.Name]; exists {
		return &GetPidResponse{Pid: pid}, nil
	} else {
		return nil, ErrGlobalNameNotRegistered
	}
}

func (g *globalProcessManagingServerGrain) Register(request *RegisterGlobalProcessRequest, context cluster.GrainContext) (*Noop, error) {
	// 先检查节点
	if _, exists := g.address2Node[request.Pid.Address]; !exists {
		return nil, ErrNodeNotExists
	}

	if _, exists := g.processes[request.Name]; exists {
		return nil, ErrGlobalNameExists
	} else {
		g.processes[request.Name] = request.Pid
		g.processes2Names[makeProcessKey(request.Pid)] = request.Name
		// 监听
		context.Watch(request.Pid)
		return nil, nil
	}
}

func (g *globalProcessManagingServerGrain) Reset(request *ResetRequest, context cluster.GrainContext) (*Noop, error) {
	for _, info := range request.Pids {
		if _, err := g.Register(info, context); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (g *globalProcessManagingServerGrain) SyncAllOthers(request *SyncAllOthersRequest, context cluster.GrainContext) (*GlobalProcessList, error) {
	var ret []*GlobalProcess
	for name, pid := range g.processes {
		ret = append(ret, &GlobalProcess{Name: name, Pid: pid})
	}

	return &GlobalProcessList{List: ret}, nil
}

func (g *globalProcessManagingServerGrain) Join(request *JoinRequest, context cluster.GrainContext) (*JoinResponse, error) {
	if old, exists := g.nodesProcesses[request.Node]; exists {
		if old.Address == context.Sender().Address && old.Id == context.Sender().Id {
			return &JoinResponse{Pid: context.Self()}, nil
		}
		return nil, ErrNodeExists
	}

	g.nodesProcesses[request.Node] = context.Sender()
	g.address2Node[context.Sender().Address] = request.Node
	// 监听
	context.Watch(context.Sender())
	return &JoinResponse{Pid: context.Self()}, nil
}

func (g *globalProcessManagingServerGrain) onProcessDown(who *actor.PID)  {
	key := makeProcessKey(who)
	if name, exists := g.processes2Names[key]; exists {
		delete(g.processes2Names, key)
		delete(g.processes, name)
	} else {
		// 是否是节点管理进程
		if name, exists = g.address2Node[who.Address]; exists {
			nodePid := g.nodesProcesses[name]
			if nodePid.Address == who.Address && nodePid.Id == who.Id {
				delete(g.nodesProcesses, name)
				delete(g.address2Node, who.Address)
			}
		}
	}
}

func makeProcessKey(pid *actor.PID) string {
	return fmt.Sprintf("%s@%s", pid.Address, pid.Id)
}
