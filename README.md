# go-rigger
## 关于go-rigger
go-rigger基于actor框架[protoactor-go](https://github.com/AsynkronIT/protoactor-go),[官网地址](https://proto.actor)  
go-rigger对protoactor-go进行了轻度包装,以使框架更符合Erlang开发者的习惯
同时go-rigger配置化了应用的启动,使得应用在开发时可以在单节点上运行,以方便调试,而在发布时又可以根据配置将应用中的各个进程分布到不同的节点上去运行

## 安装
```shell script
go get github.com/saintEvol/go-rigger
```

因为go-rigger推荐服务之前使用protobuffer消息进行通信(为保证跨节点),因此还需要安装protobuf插件, 请注意版本:
```shell script
go get github.com/gogo/protobuf/protoc-gen-gogoslick@v1.2.1
```
protoc使用示例:
```
protoc -I=. --gogoslick_out=plugins=grpc:. -I=./vendor ./rigger/protos.proto
```

## 术语
+ Actor模型 一种并发模型, 与共享内存模型相反, Actor模型不共享任何数据, 所有实体间通过消息来进行合作,交互, 原生支持Actor模型的典型语言是Erlang,
  关于此模型的更多信息, 请自行了解
+ Actor Actor模型中的关键概念, actor是Actor模型中的运行实体, actor拥有自己的状态,资源和行为; actor之间通过消息互相协作,
  因为actor在行为上和系统进程很相似,因此有些语言也将Actor称之为 进程, 本说明中后续也将延续此命名

+ Application 一种进程的类型, 此类型的进程被称之为应用进程, 是go-rigger程序的根节点, 每个go-rigger程序至少包含一个应用进程

+ Supervisor 一种进程的类型, 此类型的进程被称之为监控进程, 此类型的进程, 主要用于对其它进程(子进程)进行监控,并根据一定的策略,在子进程异常终止时,对其进行重启等操作;
  一般不(建议)在此类进程中进行任何与业务相关的逻辑
+ GeneralServer 一种进程的类型, 此类型的进程被称为通用服务器进程, 此类型的进程, 主要用于执行各类业务逻辑, 其往往是某个 Supervisor进程的子进程

## Hello World
```golang
package helloWorld

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/saintEvol/go-rigger/rigger"
	"time"
)


func Start()  {
	_ = rigger.Start(appName, "")
}

/*
hello world 应用
*/
// 应用名,应用标识
const appName = "helloWorldApp"
func init() {
	rigger.Register(appName, rigger.ApplicationBehaviourProducer(func() rigger.ApplicationBehaviour {
		return &helloWorldApp{}
	}))

	// 依赖 rigger-amqp应用, 保证在启动helloWroldApp 前,会先启动 rigger-amqp
	// rigger.DependOn("rigger-amqp")
}
type helloWorldApp struct {

}

func (h *helloWorldApp) OnRestarting(ctx actor.Context) {
}

// 如果返回的错误非空,则表示启动失败,则会停止启动后续进程
func (h *helloWorldApp) OnStarted(ctx actor.Context, args interface{}) error {
	fmt.Print("start hello world")
	return nil
}

func (h *helloWorldApp) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (h *helloWorldApp) OnStopping(ctx actor.Context) {
}

func (h *helloWorldApp) OnStopped(ctx actor.Context) {
}

func (h *helloWorldApp) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	// 一对一
	supFlag.StrategyFlag = rigger.OneForOne
	// 最多尝试10次
	supFlag.MaxRetries = 10
	// 最多尝试1秒
	supFlag.WithinDuration = 1 * time.Second
	// 任何原因下都重启
	supFlag.Decider = func(reason interface{}) actor.Directive {
		return actor.RestartDirective
	}

	// 将helloWorldSup 设置为应用的子进程
	childSpecs = []*rigger.SpawnSpec {
		rigger.DefaultSpawnSpec(helloWorldSupName),
	}

	return
}

/*
helloWorld 的监控进程,负责对业务子进程的监控及重启
此进程不处理任何业务
*/
const helloWorldSupName = "helloWorldSup"

func init() {
	rigger.Register(helloWorldSupName, rigger.SupervisorBehaviourProducer(func() rigger.SupervisorBehaviour {
		return &helloWorldSup{}
	}))
}
type helloWorldSup struct {
	
}

func (h helloWorldSup) OnRestarting(ctx actor.Context) {
}

// 如果返回非空,则表示启动失败,会停止后续进程的启动
func (h helloWorldSup) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (h helloWorldSup) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (h helloWorldSup) OnStopping(ctx actor.Context) {
}

func (h helloWorldSup) OnStopped(ctx actor.Context) {
}

func (h helloWorldSup) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []*rigger.SpawnSpec) {
	childSpecs = []*rigger.SpawnSpec{
		// 配置一个子进程(业务进程)
		rigger.DefaultSpawnSpec(helloWorldServerName),
	}
	return
}

const helloWorldServerName = "helloWorldServer"

/*
helloWorld服务器, 负责响应外部消息,完成业务处理
*/
func init() {
	rigger.Register(helloWorldServerName, rigger.GeneralServerBehaviourProducer(func() rigger.GeneralServerBehaviour {
		return &helloWordServer{}
	}))
}
type helloWordServer struct {
	
}

func (h *helloWordServer) OnRestarting(ctx actor.Context) {
}

func (h *helloWordServer) OnStarted(ctx actor.Context, args interface{}) error {
	return nil
}

func (h *helloWordServer) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (h *helloWordServer) OnStopping(ctx actor.Context) {
}

func (h *helloWordServer) OnStopped(ctx actor.Context) {
}

// 消息处理, 框架会根据需要将返回值回复给请求进程
func (h *helloWordServer) OnMessage(ctx actor.Context, message interface{}) proto.Message {
	return nil
}

```

## 如何使用本框架

go-rigger程序的开发通常开始于一张应用的进程规划图,比如, 想要开发一款在线游戏, 也许,我们脑海里会想到:
+ 我需要一个进程来处理来自客户端的连接, 也许可以命名为: ```GatewayServer```
+ 还需要一个进程来处理玩家的登录请求: ```LoginServer```
+ 对于成功登录的每个玩家, 如果我能用一个单独的进程来维护其运行时状态及行为, 事情应该会简单很多, 所以每个玩家我们需要一个进程: ```PlayerServer```
+ 不同的玩家可能会有一些并发操作, 比如: 创建用户时, 需要生成一个全局的用户ID, 这个ID显然不能由各个用户进程各自生成(除非采用某种UUID算法),
  因此我应该需要一个统一的进程来处理这些操作, 以便将临界操作顺序化, 不妨将这个进程命名为: ```PlayerManagingServer```
+ 当然, 根据之前的说法,每个rigger-go程序应该至少包含一个Application进程, 所以我们再增加一个```GameApp```
+ 为了能够在运行时监测各个进程的状态, 并在各进程异常退出时对其进行重启, 我们还需要一个监控进程, 命名为: ```GameSup```
现在, 我们的游戏的进程树应该如下所示:

```mermaid
graph TD
    GameApp[GameApp] -->  GameSup(GameSup)
    GameSup(GameSup) --> GatewayServer>GatewayServer]
    GameSup(GameSup) --> LoginServer>LoginServer]
    GameSup(GameSup) --> PlayerManagingServer>PlayerManagingServer]
    GameSup(GameSup) --> PlayerServer>PlayerServer]
```

当我们有了关于进程树的规划后,就可以开始着手编码了:

下面我们实现进程树中各个进程, 请注意,go-rigger中各个类型的进程的自定义逻辑不是通过实现如: Appication, Supervisor, GeneralServer等接口来实现的(实际上也不存在这些接口)
相反, go-rigger提供了对应的行为模式接口来供用户自定义逻辑, 分别是: ```ApplicationBehaviour```, ```SupervisorBehaviour``` 与 ```GeneralServerBehaviour```
简单起见,下面将以GameSup为例,来讲解go-rigger程序是如何编写的, 详细的例程请参加_examples

首先我们创建一个结构体, 如下:
```go
import (
	// rigger-go 是基于protoacto-go的,所以这里会自动引入
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/saintEvol/go-rigger/rigger"
)

// 游戏监控进程的进程id, 用于在go-rigger中唯一标识某个进程
const gameSupName = "gameSup"

// gameSup, 游戏的总监控进程
type gameSup struct {

}
```
请注意代码中的```const gameSupName = "gameSup"```, 在go-rigger里我们常常使用这种方式为进程定义一个进程名/进程ID/进程标识,  
go-rigger会使用这个标识来获取进程的一些重要信息(后面还会提及)  

根据前面的进程树,gameSup是一个监控进程, 所以让我们为它实现SupervisorBehaviour接口:
```go
import (
	// rigger-go 是基于protoacto-go的,所以这里会自动引入
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/saintEvol/go-rigger/rigger"
)

// 游戏监控进程的进程id, 用于在go-rigger中唯一标识某个进程
const gameSupName = "gameSup"

// 在init函数中,注册此进程的重要信息
func init() {
	// 注册进程的producer, 每次启动一个此进程时,go-rigger会使用此producer生成一个实例
	// 之后,这个实例会成为进程的回调模块
	var supProducer rigger.SupervisorBehaviourProducer = func() rigger.SupervisorBehaviour {
		return &gameSup{}
	}
	rigger.Register(gameSupName, supProducer)
}

// gameSup, 游戏的总监控进程
type gameSup struct {

// Interface: SupervisorBehaviours
func (g *gameSup) OnRestarting(ctx actor.Context) {
}

// 服务启动时的回调函数, 一般在此回调中进行初始化, 但请不要在此进行比较耗时的初始化操作
// 如果需要进行比较耗时的初始化,请在OnPostStarted中进行
// 此回调结束后,即意味着进程启动完成
func (g *gameSup) OnStarted(ctx actor.Context, args interface{}) {
}

// 进程启动完成后的回调,此回调紧接在OnStarted调用
// 可在此进行一些比较耗时的操作, 此回调结束后,进程才开始处理外部消息
func (g *gameSup) OnPostStarted(ctx actor.Context, args interface{}) {
}

// 进程即将停止时的回调
func (g *gameSup) OnStopping(ctx actor.Context) {
}

// 进程停止后的回调
func (g *gameSup) OnStopped(ctx actor.Context) {
}

// 获取监控信息, 仅监控类和应用类的行为模式具有此回调, 通过此回调,告知go-rigger,此进程具有哪些子进程,并通过何种方式监控它们
func (g *gameSup) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []interface{}) {
	return nil
}
}
````
同时也请注意代码中的init函数部分, 为了保证进程能够被正确启动, go-rigger中需要通过init函数为每个进程注册类似的启动信息

至此,我们已经完成了一监控进程的编写,不过这个监控进程还不会启动和监控任何子进程, 这是因为我们还没有为其完成需要的监控信息, 现在让我们完成它:
```go

import (
	// rigger-go 是基于protoacto-go的,所以这里会自动引入
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/saintEvol/go-rigger/rigger"
)

// 游戏监控进程的进程id, 用于在go-rigger中唯一标识某个进程
const gameSupName = "gameSup"

// 在init函数中,注册此进程的重要信息
func init() {
	// 注册进程的producer, 每次启动一个此进程时,go-rigger会使用此producer生成一个实例
	// 之后,这个实例会成为进程的回调模块
	var supProducer rigger.SupervisorBehaviourProducer = func() rigger.SupervisorBehaviour {
		return &gameSup{}
	}
	rigger.Register(gameSupName, supProducer)
}

// gameSup, 游戏的总监控进程
type gameSup struct {

// Interface: SupervisorBehaviours
func (g *gameSup) OnRestarting(ctx actor.Context) {
}

// 服务启动时的回调函数, 一般在此回调中进行初始化, 但请不要在此进行比较耗时的初始化操作
// 如果需要进行比较耗时的初始化,请在OnPostStarted中进行
// 此回调结束后,即意味着进程启动完成
func (g *gameSup) OnStarted(ctx actor.Context, args interface{}) {
}

// 进程启动完成后的回调,此回调紧接在OnStarted调用
// 可在此进行一些比较耗时的操作, 此回调结束后,进程才开始处理外部消息
func (g *gameSup) OnPostStarted(ctx actor.Context, args interface{}) {
}

// 进程即将停止时的回调
func (g *gameSup) OnStopping(ctx actor.Context) {
}

// 进程停止后的回调
func (g *gameSup) OnStopped(ctx actor.Context) {
}

// 获取监控信息, 仅监控类和应用类的行为模式具有此回调, 通过此回调,告知go-rigger,此进程具有哪些子进程,并通过何种方式监控它们
func (g *gameSup) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []interface{}) {
	// 监控进程会依次同步启动下列进程,
	childSpecs = append(childSpecs, gatewayServerName/*路由服务进程的名字*/)
	childSpecs = append(childSpecs, loginServerName/*登录服务进程的名字*/)
	childSpecs = append(childSpecs, playerManagingServerName/*玩家管理进程的名字*/)
	// 之前的进程树中并没有此进程,但我们发现, playerServer是一个动态进程,它需要在运行时,随着游戏进程,动态添加(当有玩家登录成功后,就需要启动一个玩家进程)
	// 因些,playerServer不适合在这里静态启动,且和其它进程的行为有较大差异, 因此我们新加入了一个玩家监控进程```playerSup```来管理玩家进程
	// 玩家进程将由此进程来根据需要动态启动
	childSpecs = append(childSpecs, playerServerSupName/*玩家监控进程的名字*/)
	
	return
}
}
```
请注意,现在我们在OnGetSupFlag回调中增加了一些代理, 这些代码给GameSup增加了一些子进程信息,现在当GameSup进程被启动时, 其自身启动完成后,就会按照OnGetSupFlag中的顺序依次启动子进程
同时我们发现,go-rigger是通过进程名字来确定进程的
我们还发现,因为之前考虑不周, 我们需要增加一个新的玩家监控进程来管理/监控所有的玩家进程, 下面我们就看看这个进程的代码:
```go

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/saintEvol/go-rigger/rigger"
)


// 玩家监控进程名字
const playerServerSupName = "playerServerSup"

func init() {
	var producer rigger.SupervisorBehaviourProducer = func() rigger.SupervisorBehaviour {
		return &playerServerSup{}
	}
	rigger.Register(playerServerSupName, producer)
}

type playerServerSup struct {

}

func (p *playerServerSup) OnRestarting(ctx actor.Context) {
}

func (p *playerServerSup) OnStarted(ctx actor.Context, args interface{}) {
}

func (p *playerServerSup) OnPostStarted(ctx actor.Context, args interface{}) {
}

func (p *playerServerSup) OnStopping(ctx actor.Context) {
}

func (p *playerServerSup) OnStopped(ctx actor.Context) {
}

// 
func (p *playerServerSup) OnGetSupFlag(ctx actor.Context) (supFlag rigger.SupervisorFlag, childSpecs []interface{}) {
	supFlag.StrategyFlag = rigger.SimpleOneForOne // 将子进程(玩家进程)变为动态进程
	childSpecs = append(childSpecs, playerServerName)

	return
}

```
通过代码,我们发现 ```playerSup``` 和```gameSup```并没有太大差别, 最主要的差别在OnGetSupFlag回调中, 在```playerSup.OnGetSup```中,
我们通过代码```supFlag.StrategyFlag = rigger.SimleOneForOne``` 将重启策略设置为了"SimpleOneForOne", 此模式和"OneForOne"的唯一区别是:
监控进程在自身启动后不会自动启动子进程,而是需要,则用户手动调用 StartChild来启动子进程, 这种模式非常适合我们的玩家进程

详细的用例,请参见 _examples

另外, 从上面的代码可以看出, go-rigger中的进程往往需要通过一个"进程名"来注册一些诸如producer, start fun之类的重要信息, 实际上, 在最初的版本中,
go-rigger并不需要如此,而是在OnGetSupFlag函数中直接返回这些必要信息, 但为了使go-rigger程序更方便的布署为分布式应用, 后面,改用了当前的方式

go-rigger基于actor框架 protoactor,  因此,天然的, go-rigger也拥有了actor框架的一些独有的优势, 如:

+  天然分布式, actor框架通常具有位置透明性的特点,这导致actor中的进程在何处运行,往往并不重要
+  天然并发, actor框架中,各个actor之前独立并发运行,相互之间只通过消息进行交互
+  无锁


## go-rigger应用启动
目前,go-rigger提供了三种启动方式:
+ 直接指定应用名称进行启动:
    ```rigger.Start(applicationName, applicationPath)```  
   上面的代码将直接启动对应的应用及其下所有服务,并自动加载applicationPath路径上的配置文件  
  如,要启动示例中的应用(_examples/normal_starting),可以在main函数中调用:```rigger.Start(gameAppName, "")```
+ 通过配置启动:```rigger.StartWithConfig("launchConfigPath", "applicationConfigPath")```
+ 通过命令行启动: ```rigger.StartFromCli()```  
 这是最常用也是最推荐的启动方式,使用此方式启动应用时, 支持以下命令行选项:
  + -l launchConfigPath 指定启动配置文件的路径
  + -c applicationConfigPath 指定应用的配置文件路径
  + -n 直接指定要起动的应用的名称

   如果使用这种方式启动应用,-l与-n选项,必须指定一个

   如: ```go run ./main.go -l xxx.yml -c app.toml``` ```go run ./main.go -n myApp```


