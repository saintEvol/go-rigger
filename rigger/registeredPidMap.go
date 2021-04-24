package rigger

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"sync"
)

// 已经注册的进程,此MAP只能由applicationTopSup在OnStarted时对自己进行注册时和riggerManagingServer进程对其它进程进行注册时修改
//registeredProcessLock sync.RWMutex
//registeredProcess = make(map[string]*actor.PID)
func newRegisteredPidMap() *registeredPidMap {
	return &registeredPidMap{
		lock: new(sync.RWMutex),
		pids: make(map[string]*actor.PID),
	}
}

type registeredPidMap struct {
	lock *sync.RWMutex
	pids map[string]*actor.PID
}

func (r *registeredPidMap) get(name string) (*actor.PID, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if pid, exists := r.pids[name]; exists {
		return pid, true
	} else {
		return nil, false
	}
}

func (r *registeredPidMap) add(name string, pid *actor.PID, allowExists bool /*是否允许已经存在*/) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if !allowExists {
		if _, exists := r.pids[name]; exists {
			return ErrLocalNameExists
		}
	}

	r.pids[name] = pid
	return nil
}

func (r *registeredPidMap) remove(name string)  {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.pids, name)
}

func (r *registeredPidMap) foreach(f func(name string, pid *actor.PID)) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	for n, p := range r.pids {
		f(n, p)
	}
}


