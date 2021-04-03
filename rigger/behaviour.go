package rigger

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)
type MessageHandler func(ctx actor.Context, message interface{}) proto.Message
type Behaviour []MessageHandler

func NewBehaviour() Behaviour {
	return make(Behaviour, 0)
}

func (b *Behaviour) Become(handler MessageHandler) {
	b.clear()
	b.push(handler)
}

func (b *Behaviour) BecomeStacked(receive MessageHandler) {
	b.push(receive)
}

func (b *Behaviour) UnbecomeStacked() {
	b.pop()
}

func (b *Behaviour) clear() {
	if len(*b) == 0 {
		return
	}

	for i := range *b {
		(*b)[i] = nil
	}
	*b = (*b)[:0]
}

func (b *Behaviour) peek() (v MessageHandler, ok bool) {
	l := b.len()
	if l > 0 {
		ok = true
		v = (*b)[l-1]
	}
	return
}

func (b *Behaviour) push(v MessageHandler) {
	*b = append(*b, v)
}

func (b *Behaviour) pop() (v MessageHandler, ok bool) {
	l := b.len()
	if l > 0 {
		l--
		ok = true
		v = (*b)[l]
		(*b)[l] = nil
		*b = (*b)[:l]
	}
	return
}

func (b *Behaviour) len() int {
	return len(*b)
}

func (b *Behaviour) handleMessage(ctx actor.Context, message interface{}) (proto.Message, bool) {
	behavior, ok := b.peek()
	if ok {
		return behavior(ctx, message), true
	} else {
		logrus.Errorf("empty behavior called, pid: %v",  ctx.Self())
		return nil, false
	}
}

type iBehaviour interface {
	handleMessage(ctx actor.Context, message interface{}) (proto.Message, bool)
}