package moni

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/saintEvol/go-rigger/rigger"
)

const mainApplicationName = "main_application"
const mainApplicationSupName = "main_application_sup"
const userManagerSupName = "user_manager_sup"
const userManagerServerName = "user_manager"
const userSupName = "user_sup"
const userServerName = "user_server"
const loginServerName = "user_login_server"
const missionServerSupName = "mission_server_sup"
const missionServerName = "missions_server"
const restServerSupName = "rest_server_sup"
const restServerName = "rest_server"
func init() {
	rigger.Register(mainApplicationName, rigger.ApplicationBehaviourProducer(func() rigger.ApplicationBehaviour {
		return &mainApplication{}
	}))
	rigger.Register(mainApplicationSupName, rigger.SupervisorBehaviourProducer(func() rigger.SupervisorBehaviour {
		return &mainApplicationSup{}
	}))
	rigger.Register(userManagerSupName, rigger.SupervisorBehaviourProducer(func() rigger.SupervisorBehaviour {
		return &userManagerSup{}
	}))
	rigger.Register(userManagerServerName, rigger.GeneralServerBehaviourProducer(func() rigger.GeneralServerBehaviour {
		return &userManagerServer{}
	}))
	rigger.Register(userSupName, rigger.SupervisorBehaviourProducer(func() rigger.SupervisorBehaviour {
		return &userSup{}
	}))
	rigger.Register(userServerName, rigger.GeneralServerBehaviourProducer(func() rigger.GeneralServerBehaviour {
		return &userServer{}
	}))
	rigger.Register(loginServerName, rigger.GeneralServerBehaviourProducer(func() rigger.GeneralServerBehaviour {
		return &userLoginServer{}
	}))
	rigger.Register(missionServerSupName, rigger.SupervisorBehaviourProducer(func() rigger.SupervisorBehaviour {
		return &missionServerSup{}
	}))
	rigger.Register(missionServerName, rigger.GeneralServerBehaviourProducer(func() rigger.GeneralServerBehaviour {
		return &missionServer{}
	}))
	rigger.Register(restServerSupName, rigger.SupervisorBehaviourProducer(func() rigger.SupervisorBehaviour {
		return &restServerSup{}
	}))
	rigger.Register(restServerName, rigger.GeneralServerBehaviourProducer(func() rigger.GeneralServerBehaviour {
		return &restServer{}
	}))
}

type mainApplication struct {

}

func (m mainApplication) OnRestarting(ctx actor.Context) {
	return
}

func (m mainApplication) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (m mainApplication) OnPostStarted(ctx actor.Context, args interface{}) {
	return
}

func (m mainApplication) OnStopping(ctx actor.Context) {
	return
}

func (m mainApplication) OnStopped(ctx actor.Context) {
	return
}

func (m mainApplication) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	return
}

type mainApplicationSup struct {

}

func (m mainApplicationSup) OnRestarting(ctx actor.Context) {
	return
}

func (m mainApplicationSup) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (m mainApplicationSup) OnPostStarted(ctx actor.Context, args interface{}) {
	return
}

func (m mainApplicationSup) OnStopping(ctx actor.Context) {
	return
}

func (m mainApplicationSup) OnStopped(ctx actor.Context) {
	return
}

func (m mainApplicationSup) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	return
}

type userManagerSup struct {

}

func (u userManagerSup) OnRestarting(ctx actor.Context) {
	return
}

func (u userManagerSup) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (u userManagerSup) OnPostStarted(ctx actor.Context, args interface{}) {
	return
}

func (u userManagerSup) OnStopping(ctx actor.Context) {
	return
}

func (u userManagerSup) OnStopped(ctx actor.Context) {
	return
}

func (u userManagerSup) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	return
}

type userManagerServer struct {
	
}

func (u userManagerServer) OnRestarting(ctx actor.Context) {
	return
}

func (u userManagerServer) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (u userManagerServer) OnPostStarted(ctx actor.Context, args interface{}) {
	return
}

func (u userManagerServer) OnStopping(ctx actor.Context) {
	return
}

func (u userManagerServer) OnStopped(ctx actor.Context) {
	return
}

func (u userManagerServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	return nil
}

type userSup struct {

}

func (u userSup) OnRestarting(ctx actor.Context) {
	return
}

func (u userSup) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (u userSup) OnPostStarted(ctx actor.Context, args interface{}) {
	return
}

func (u userSup) OnStopping(ctx actor.Context) {
	return
}

func (u userSup) OnStopped(ctx actor.Context) {
	return
}

func (u userSup) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	return
}

type userServer struct {

}

func (u userServer) OnRestarting(ctx actor.Context) {
	return
}

func (u userServer) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (u userServer) OnPostStarted(ctx actor.Context, args interface{}) {
	return
}

func (u userServer) OnStopping(ctx actor.Context) {
	return
}

func (u userServer) OnStopped(ctx actor.Context) {
	return
}

func (u userServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	return nil
}

type userLoginServer struct {

}

func (u userLoginServer) OnRestarting(ctx actor.Context) {
	return
}

func (u userLoginServer) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (u userLoginServer) OnPostStarted(ctx actor.Context, args interface{}) {
	return
}

func (u userLoginServer) OnStopping(ctx actor.Context) {
	return
}

func (u userLoginServer) OnStopped(ctx actor.Context) {
	return
}

func (u userLoginServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	return nil
}

type missionServerSup struct {

}

func (m missionServerSup) OnRestarting(ctx actor.Context) {
	return
}

func (m missionServerSup) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (m missionServerSup) OnPostStarted(ctx actor.Context, args interface{}) {
	return
}

func (m missionServerSup) OnStopping(ctx actor.Context) {
	return
}

func (m missionServerSup) OnStopped(ctx actor.Context) {
	return
}

func (m missionServerSup) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	return
}

type missionServer struct {

}

func (m missionServer) OnRestarting(ctx actor.Context) {
	return
}

func (m missionServer) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (m missionServer) OnPostStarted(ctx actor.Context, args interface{}) {
	return
}

func (m missionServer) OnStopping(ctx actor.Context) {
	return
}

func (m missionServer) OnStopped(ctx actor.Context) {
	return
}

func (m missionServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	return nil
}

type restServerSup struct {

}

func (r restServerSup) OnRestarting(ctx actor.Context) {
	return
}

func (r restServerSup) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (r restServerSup) OnPostStarted(ctx actor.Context, args interface{}) {
	return
}

func (r restServerSup) OnStopping(ctx actor.Context) {
	return
}

func (r restServerSup) OnStopped(ctx actor.Context) {
	return
}

func (r restServerSup) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	return
}

type restServer struct {

}

func (r restServer) OnRestarting(ctx actor.Context) {
	return
}

func (r restServer) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (r restServer) OnPostStarted(ctx actor.Context, args interface{}) {
	return
}

func (r restServer) OnStopping(ctx actor.Context) {
	return
}

func (r restServer) OnStopped(ctx actor.Context) {
	return
}

func (r restServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	return nil
}
