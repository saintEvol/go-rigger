package rigger

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"testing"
)

type behaviourStruct struct {
	Behaviour
}
func TestBehaviour(t *testing.T)  {
	b := &behaviourStruct{}
	be, ok := interface{}(b).(iBehaviour)
	if !ok {
	 t.Error("not implement")
	}
	count := 0
	b.Become(func(ctx actor.Context, message interface{}) proto.Message {
		add := message.(int)
		count += add
		return nil
	})
	be.handleMessage(nil, 10)
	if count != 10 {
		t.Error("not exec")
	}
}