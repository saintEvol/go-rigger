package tests

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"testing"
)

type TestApp struct {

}

func (t *TestApp) OnRestarting(ctx actor.Context) {
	panic("implement me")
}

func (t *TestApp) OnStarted(ctx actor.Context) {
	panic("implement me")
}

func (t *TestApp) OnStopping(ctx actor.Context) {
	panic("implement me")
}

func (t *TestApp) OnStopped(ctx actor.Context) {
	panic("implement me")
}

func TestRigger(t *testing.T)  {
}