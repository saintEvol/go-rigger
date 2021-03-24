package normal_starting

import (
	"fmt"
	"github.com/saintEvol/go-rigger/rigger"
	"github.com/sirupsen/logrus"
	"time"
)

func StartNormal()  {
	logrus.SetLevel(logrus.TraceLevel)
	go func() {
		time.Sleep(3 * time.Second)
		// 模拟注册和登录
		register("test1", "hello kitty")
		register("test2", "coolman")
		login("test1")
		login("test2")
		time.Sleep(1 * time.Second)
		// 模拟广播
		broadcast()
		// 玩家下线
		logout("test1")
		// 测试下线后,广播
		broadcast()
		broadcast()
		broadcast()
	}()
	// 启动游戏应用, 会阻塞当前进程,直到收到打断信号
	if err := rigger.Start(GameAppName, ""); err != nil {
		fmt.Printf("error when starting app, app id: %s, err: %v\r\n", GameAppName, err)
	}
}

func register(userName string, nickname string)  {
	// 模拟玩家注册
	// 为了可以跨机器远程通讯, 所有消息使用了protobuf生成
	spec := Register{UserName: userName, Nickname: nickname}
	if managerPid, ok := rigger.GetPid(playerManagingServerName); ok {
		// 获取应用
		logrus.Tracef("managerPid: %v", managerPid)
		// TODO 是否可以优化发消息
		future := rigger.Root().Root.RequestFuture(managerPid, &spec, 3 * time.Second)
		if resp, err := future.Result(); err == nil {
			ret := resp.(*RegisterResp)
			if ret.Error == "" {
				fmt.Printf("register success, user name: %s, id: %d\r\n", ret.UserName, ret.UserId)
			} else {
				fmt.Printf("register failed, user name: %s, reason: %s\r\n", spec.UserName, ret.Error)
			}

		} else {
			fmt.Printf("register failed, user name: %s, reason: %s\r\n", spec.UserName, err.Error())
		}
	}
}

func login(userName string)  {
	spec := LoginSpec{UserName: userName}
	if managerPid, ok := rigger.GetPid(playerManagingServerName); ok {
		// 获取应用
		// TODO 是否可以优化发消息
		// TODO 应该是从loginServer登录
		future := rigger.Root().Root.RequestFuture(managerPid, &spec, 3 * time.Second)
		if resp, err := future.Result(); err == nil {
			ret := resp.(*LoginResp)
			if ret.Error == "" {
				fmt.Printf("login success, user name: %s, id: %d\r\n", ret.UserName, ret.UserId)
			} else {
				fmt.Printf("login failed, user name: %s, reason: %s\r\n", spec.UserName, ret.Error)
			}

		} else {
			fmt.Printf("login failed, user name: %s, reason: %s\r\n", spec.UserName, err.Error())
		}
	}

}

func logout(username string)  {
	if managerPid, ok := rigger.GetPid(playerManagingServerName); ok {
		f := rigger.Root().Root.RequestFuture(managerPid, &LogoutSpec{UserName: username}, 1 * time.Second)
		if ret, err := f.Result(); err == nil{
			r := ret.(*LogoutResp)
			if r.Error == "" {
				fmt.Printf("success logout:%s\r\n", username)
			} else {
				fmt.Printf("failed to logout: %s, reason: %s\r\n", username, r.Error)
			}
		}
	} else {
		fmt.Printf("faild to get manager pid\r\n")
	}
}

func broadcast()  {
	if broadcastPid, ok := rigger.GetPid(playerBroadcastServerName); ok {
		// 获取应用
		rigger.Root().Root.Send(broadcastPid, &Broadcast{Content: "hello"})
	}
}
