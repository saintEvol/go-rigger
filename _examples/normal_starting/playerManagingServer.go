package normal_starting

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/saintEvol/go-rigger/rigger"
	"time"
)

const playerManagingServerName = "playerManagingServerName"

func init() {
	var producer rigger.GeneralServerBehaviourProducer = func() rigger.GeneralServerBehaviour {
		return &playerManagingServer{}
	}
	rigger.Register(playerManagingServerName, producer)
}
type playerManagingServer struct {
	nextPlayerId uint64 // 下一个可用的玩家ID
	allPlayers map[string]*player // 所有玩家
	onlinePlayers map[uint64]*player // 在线玩家
}

func (p *playerManagingServer) OnRestarting(ctx actor.Context) {
}

func (p *playerManagingServer) OnStarted(ctx actor.Context, args interface{}) {
	// 初始化玩家ID
	p.nextPlayerId = 1
	p.allPlayers = make(map[string]*player)
	p.onlinePlayers = make(map[uint64]*player)
	fmt.Printf("%s started!, all players: %v, onlines: %v\r\n", playerManagingServerName, p.allPlayers, p.onlinePlayers)
}

func (p *playerManagingServer) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (p *playerManagingServer) OnStopping(ctx actor.Context) {
}

func (p *playerManagingServer) OnStopped(ctx actor.Context) {
}

func (p *playerManagingServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	switch msg := message.(type) {
	case *Register:
		// 测试,略去检测用户名等是否合法的逻辑
		if _, ok := p.getPlayerByName(msg.UserName); ok {
			return &RegisterResp{
				Error:    fmt.Sprintf("user name: %s existed, change another one pls!", msg.UserName),
			}
		} else {
			userId := p.nextPlayerId
			p.nextPlayerId += 1
			player := &player {
				userName: msg.UserName,
				userId: userId,
			}
			p.allPlayers[msg.UserName] = player

			return &RegisterResp{
				UserName: msg.UserName,
				UserId:   userId,
			}
		}
	case *LoginSpec:
		if data, ok := p.getPlayerByName(msg.UserName); ok {
			// 是否在线
			if p.isOnline(data.userId) {
				return &LoginResp{
					Error:    fmt.Sprintf("%s has been already online!", data.userName),
					UserName: data.userName,
					UserId:   data.userId,
				}
			} else {
				// 启动玩家进程
				supPid, _ := rigger.GetPid(playerServerSupName)
				if _, err := rigger.StartChildSync(ctx, supPid, data.userId, 3 * time.Second); err == nil {
					p.onlinePlayers[data.userId] = data
					return &LoginResp{
						Error:    "",
						UserName: data.userName,
						UserId:   data.userId,
					}
				} else {
					return &LoginResp{
						Error: err.Error(),
						UserName: msg.UserName,
					}
				}
			}
		} else {
			return &LoginResp{
				Error:    fmt.Sprintf("%s not existed, pls check!", msg.UserName),
				UserName: msg.UserName,
				UserId:   0, // 此错误不返回user id
			}
		}
	}

	return nil
}
func (p *playerManagingServer) getPlayerByName(userName string) (*player, bool) {
	data, ok := p.allPlayers[userName]
	return data, ok
}
// 用户是否存在
func (p *playerManagingServer) exists(userName string) bool {
	_, ok := p.allPlayers[userName]
	return ok
}

func (p *playerManagingServer) isOnline(userId uint64) bool {
	_, ok := p.onlinePlayers[userId]
	return ok
}
