package rigger
import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/sirupsen/logrus"
)

/*
因为vitural actor无法处理系统消息(actor.SystemMessage)所以将主要逻辑放到普通actor中
*/
// 消息
type _getPid struct {
	name string
}

type _register struct {
	name string
	pid *actor.PID
}

type _reset struct {
	pids []*RegisterGlobalProcessRequest `protobuf:"bytes,1,rep,name=pids,proto3" json:"pids,omitempty"`
}

type _syncAllOthers struct {
	exception string
}

type _join struct {
	name string
	pid *actor.PID
}

const globalManagerName = "globalManagerSer"
type globalManager struct {
	processes map[string]*actor.PID // 所有注册的全局进程, name => *pid
	processes2Names map[string]string // 进程键到进程名的映射 key => name
	nodesProcesses     map[string]*actor.PID // 节点进程: 节点名 => *pid
	//nodes map[string]string // 节点: name => 地址
	address2Node map[string]string // 地址 => name
}

func (g *globalManager) init() {
	g.processes = make(map[string]*actor.PID)
	//g.nodes = make(map[string]string)
	g.processes2Names = make(map[string]string)
	g.nodesProcesses = make(map[string]*actor.PID)
	g.address2Node = make(map[string]string)

}

func (g *globalManager) getPid(name string) (*actor.PID, error) {
	if pid, exists := g.processes[name]; exists {
		return pid, nil
	} else {
		return nil, ErrGlobalNameNotRegistered
	}
}

func (g *globalManager) register(name string, pid *actor.PID, context actor.Context) error {
	// 先检查节点
	if _, exists := g.address2Node[pid.Address]; !exists {
		return ErrNodeNotExists
	}

	if _, exists := g.processes[name]; exists {
		return ErrGlobalNameExists
	} else {
		logrus.Tracef("register global: %s", name)
		g.processes[name] = pid
		g.processes2Names[makeProcessKey(pid)] = name
		// 监听
		context.Watch(pid)
		return nil
	}
}

func (g *globalManager) reset(info []*RegisterGlobalProcessRequest, context actor.Context) error {
	for _, info := range info{
		if err := g.register(info.Name, info.Pid, context); err != nil {
			return err
		}
	}

	return nil
}

func (g *globalManager) syncAllOthers(exception, context actor.Context) (*GlobalProcessList, error) {
	var ret []*GlobalProcess
	for name, pid := range g.processes {
		ret = append(ret, &GlobalProcess{Name: name, Pid: pid})
	}

	return &GlobalProcessList{List: ret}, nil
}

func (g *globalManager) join(node string, nodePid *actor.PID, context actor.Context) (*actor.PID, error) {
	// 先改成如果有ID相同就替换
	//if old, exists := g.nodesProcesses[node]; exists {
	//	if old.Id == nodePid.Id {
	//		g.nodesProcesses[request.Node] = nodePid
	//		delete(g.address2Node, old.Address)
	//		g.address2Node[nodePid.Address] = request.Node
	//		return &JoinResponse{Pid: context.Self()}, nil
	//	} else {
	//		return nil, ErrNodeExists
	//	}
	//}
	if _, exists := g.nodesProcesses[node]; exists {
		return nil, ErrNodeExists
	}

	g.nodesProcesses[node] = nodePid
	g.address2Node[nodePid.Address] = node

	// 监听
	context.Watch(nodePid)
	return context.Self(), nil
}

func (g *globalManager) onProcessDown(who *actor.PID)  {
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
func (g *globalManager) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		g.init()
	case *_getPid:
		if pid, err := g.getPid(msg.name); err == nil {
			ctx.Respond(pid)
		} else {
			ctx.Respond(nil)
		}
	case *_join:
		pid, err := g.join(msg.name, msg.pid, ctx)
		if err == nil {
			ctx.Respond(pid)
		} else {
			ctx.Respond(err)
		}
	case *_register:
		if err := g.register(msg.name, msg.pid, ctx); err == nil {
			ctx.Respond(nil)
		} else {
			ctx.Respond(err)
		}
	case *_reset:
		if err := g.reset(msg.pids, ctx); err == nil {
			ctx.Respond(nil)
		} else {
			ctx.Respond(err)
		}
	case *actor.Terminated:
		g.onProcessDown(msg.Who)
	}
}



