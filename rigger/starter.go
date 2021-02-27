package rigger

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/remote"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strings"
	"time"
)

// 默认启动超时时间
const startTimeOut = 10_000_000_000

var (
	isFromConfig bool = false	// 是否是从配置启动
	serversMap = make(map[string]*StartingNode) // 记录已经parse过的server,防止重复
	pidSets = make(map[string]*actor.PID)
	// 所有注册过的producer
	registerInfoMap = make(map[string]*registerInfo)
	runningApplication = make(map[string]*Application)
	startingTasks []*StartingNode
	signalChan chan os.Signal // 接收一些系统信号
)

type location struct {
	host string
	port int
}

// 远程配置
type remoteSpec struct {
	location
}

type StartingNode struct {
	name string
	fullName string // 全名,包含了其父级及祖名称,可以通过其定位的进程
	parent *StartingNode
	spawnSpec *SpawnSpec //  自身的启动规范
	children []*StartingNode
	location *location
	remote *remoteSpec
	supFlag *SupervisorFlag
}

// 启动应用
func Start(applicationId string, configPath string) error  {
	isFromConfig = false
	// 读取应用配置
	readAppConfig(configPath)
	if app, err := startApplication(applicationId); err == nil {
		setRunningApplication(applicationId, app)
	} else {
		return err
	}
	waitInterupt()
	return nil
}

/*
从命令行的启动
函数从命令行相关配置,然后根据这些配置启动应用
命令行选项:
-l 启动文件路径, 路径应该指向一个有效的yum文件, 此选项和-n必须至少有一个,且只有一个会生效, 优化使用-l可选选项
-n 应用名, 此选项和-l必须至少有一个,且只有一个会生效,优化使用-l, 可选选项
-c 应用配置文件路径, 路径可以指向一个有效的yml文件, toml文件, ini文件, json文件等, 可选项
 */
func StartFromCli()  {
	l, n, c := parseCl()
	// l与n必须有一个
	if l == nil && n == nil {
		log.Panicf("need c launch config or application name to launch go-rigger applications")
		return // 不写return, IDE无法判断后续对l的解引用总是安全的
	}

	if l != nil {
		if c == nil {
			StartWithConfig(*l, "")
		} else {
			StartWithConfig(*l, *c)
		}
	} else {
		var p string
		if c == nil {
			p = ""
		} else {
			p = *c
		}
		if err := Start(*n, p); err != nil {
			log.Panicf("error when start application: %s, reason: %s", *n, err.Error())
		}

	}
}

/*
根据配置启动应用, 如果启动失败,会panic,所以未返回错误
函数接受两个配置文件:
launchConfigPath 启动配置文件,目前为yum 文件, 里面描述了应该如何启动一个节点上的应用
appConfigPath 运行时环境配置文件, go-rigger不关心里面的内容, 此配置文件供用户自己使用,用户可以使用viper相关函数获取其中的数据
*/
func StartWithConfig(launchConfigPath string/*应用启动配置文件*/, appConfigPath string/*应用配置文件*/)  {
	isFromConfig = true

	// 先设为最高级
	log.SetLevel(6)
	// 读取配置文件
	readLaunchConfig(launchConfigPath)
	// 生成启动树
	parseConfig()
	printProcessTree(startingTasks)
	// 读取应用配置
	readAppConfig(appConfigPath)
	startApplications()

	waitInterupt()
}

// 获取正在运行中的应用,如果没有,第二个返回值为 false,否则为true
func GetRunningApplication(id string) (*Application, bool) {
	if app, ok := runningApplication[id]; ok {
		return app, true
	}

	return nil, false
}

func setRunningApplication(id string, app *Application)  {
	runningApplication[id] = app
}

// TODO 考虑将所有应用在同一个根上启动,这样各个应用间才好通信
func GetApplicationRoot(name string) *actor.ActorSystem {
	if app, ok := GetRunningApplication(name); ok {
		return app.Parent
	}

	return nil
}

// 根据配置中的名字获取其进程id, 如果存在,则返回进程ID和true, 否则返回 nil,false
// 获取到进程ID,并不意味着此进程依然存活
// 本接口只适用于静态进程
func GetPid(name string) (*actor.PID, bool) {
	if pid, ok := pidSets[name]; ok {
		return pid, ok
	} else {
		if config, ok := getConfigByName(name); ok {
			if config.location == nil {
				pid = actor.NewPID("nonhost", config.fullName)
			} else {
				pid = actor.NewPID(fmt.Sprintf("%s:%d", config.location.host, config.location.port), config.fullName)
			}
			pidSets[name] = pid
			return pid, true
		} else {
			return nil, false
		}
	}
}

// 获取动态进程的PID
func GetDynamicPid(registerName/*注册名*/ string, dynamicName/*动态名*/ string) (*actor.PID, bool) {
	if config, ok := getConfigByName(registerName); ok {
		// 先状态是不是
		if isDynamic(config) {
			fullName := fmt.Sprintf("%s/%s", config.parent.fullName, dynamicName)
			if config.location == nil {
				return actor.NewPID("nonhost", fullName), true
			} else {
				return actor.NewPID(fmt.Sprintf("%s:%d", config.location.host, config.location.port), fullName), true
			}
		} else {
			return nil, false
		}
	} else {
		return nil, false
	}
}


type registerInfo struct {
	id string
	producer interface{}
	startFun SpawnFun
}

/*
注册生成器(producer, 关于producer请参考proto-actor 相关资料)
注册生成器时, 还会注册一个默认启动函数, 默认的启动函数规则:
1. 如果进程是普通(非动态/即非SimpleOneForOne)进程, 启动函数会以id作为启动后的进程的名字
2. 如果进程是动态进程,也即SimpleOneForOne, 启动函数不会对该进程命名
3. 如果进程是SimpleOneForOne,但又需要命名,此时需要调用 RegisterStartFun来注册自定义的启动函数
*/
func Register(id string, producer interface{})  {
	RegisterStartFun(id, producer, nil)
}

/*
给进程注册producer及启动函数
*/
func RegisterStartFun(name string, producer interface{}, startFun SpawnFun)  {
	if _, ok := registerInfoMap[name]; ok {
		panic(fmt.Sprintf("duplicated producer register key:%s", name))
	}
	fmt.Println("register name:", name)
	registerInfoMap[name] = &registerInfo{
		id: name,
		producer: producer,
		startFun: startFun,
	}
}

// 获取注册信息
func getRegisterInfo(name string) (*registerInfo, bool)  {
	info, ok := registerInfoMap[name]
	return info, ok
}

// 等待打断信息
func waitInterupt()  {
	if signalChan == nil {
		signalChan = make(chan os.Signal)
	}
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	s := <- signalChan
	log.Tracef("now quite becase of signal: %s", s)
	for _, app := range runningApplication {
		if err := app.StopFuture().Wait(); err != nil {
			log.Errorf("Error when quite application: %s", app.id)
		}
	}
}

// 加载启动配置
func readLaunchConfig(path string)  {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("error when read starting config, reason:%s", err.Error())
	}
}

func startApplications()  {
	for _, node := range startingTasks{
		startNode(node)
	}
}

func startNode(node *StartingNode) {
	if node.location != nil {
		return
	}

	// 启动应用
	if app, err := startApplicationNode(node); err == nil {
		// 将启动的应用存起来
		setRunningApplication(node.name, app)
		//startRest(app, filterLocalNode(node.children))
	} else {
		log.Panicf("error when start application: %s", node.name)
	}
}

func filterLocalNode(nodes []*StartingNode) (n []*StartingNode)  {
	for _, c := range nodes {
		if c.location == nil {
			n = append(n, c)
		}
	}

	return n
}

//func startRest(app *Application, children []*StartingNode)  {
//	if len(children) <= 0 {
//		return
//	}
//
//	for _, child := range children {
//		// 只启动本地进程
//		if child.location == nil {
//			if info, ok := getRegisterInfo(child.spawnSpec.Id); ok {
//				switch info.producer.(type) {
//				case SupervisorBehaviourProducer:
//					if _, err := StartSupervisorSpec(app, child.spawnSpec); err != nil {
//						log.Panicf("faild to start %s, reason: %s", child.name, err.Error())
//					}
//				case GeneralServerBehaviourProducer:
//					if _, err := StartGeneralServerSpec(app, child.spawnSpec); err != nil {
//						log.Panicf("faild to start %s, reason: %s", child.name, err.Error())
//					}
//				default:
//					log.Panicf("faild to start %s, reason: unexpected producer type: %s", child.name, reflect.TypeOf(info.producer).Name())
//				}
//
//			} else {
//				log.Panicf("not register: %s", child.spawnSpec.Id)
//			}
//		}
//	}
//}

func startApplicationNode(node *StartingNode) (*Application, error) {
	if node.parent != nil {
		log.Panicf("application should not have Parent:%s", node.name)
	}
	spawnSpec := node.spawnSpec
	// producer先不判断了,因为生成时,已经判断过了
	// remote
	if node.remote == nil {
		return startApplicationSpec(spawnSpec)
	} else {
		system := actor.NewActorSystem()
		r := remote.NewRemote(system, remote.Configure(node.remote.host, node.remote.port))
		r.Start()
		return startApplicationWithSystem(system, node.spawnSpec)
	}
}

func parseConfig()  {
	n := viper.Get("rigger.node").([]interface{})
	parseNode(n)
}

func printProcessTree(tasks []*StartingNode)  {
	if len(tasks) <= 0 {
		log.Trace("now process to be started!")
		return
	}

	for _, node := range tasks {
		printNode(node, 0)
	}
}

func printNode(node *StartingNode, depth int)  {
	space := generateSpace("", depth)
	var nodeType string
	if node.supFlag != nil {
		nodeType = "sup"
	} else {
		nodeType = "server"
	}
	var locType string
	if node.parent != nil && node.parent.location != nil {
		locType = fmt.Sprintf("|_(host: %s, port: %d)*", node.parent.location.host, node.parent.location.port)
	} else if node.location != nil {
		locType = fmt.Sprintf("<host: %s, port: %d>", node.location.host, node.location.port)
	}
	// 是否是动态进程
	var dynamic string
	if isDynamic(node) {
		dynamic = " (动态进程)*"
	}

	log.Tracef("%s%s:%s	%s%s", space, nodeType, node.name, dynamic, locType)
	// 打印子节点的
	if len(node.children) > 0 {
		for _, c := range node.children {
			printNode(c, depth + 1)
		}
	}
}

// 解析remote信息
func parseRemote(rootMap map[interface{}]interface{}) *remoteSpec {
	if rootMap == nil {
		return nil
	}

	if rawRemoteMap, ok := rootMap["remote"]; ok {
		remoteMap := rawRemoteMap.(map[interface{}]interface{})
		if rawPort, ok := remoteMap["port"]; ok {
			port := rawPort.(int)
			var host = "127.0.0.1"
			if rawHost, ok := remoteMap["host"]; ok {
				host = rawHost.(string)
			}

			return &remoteSpec{location{
				host: host,
				port: port,
			}}
		}
	}

	return nil
}
func parseNode(node []interface{})  {
	for _, rawApp := range node {
		app := rawApp.(map[interface{}]interface{})
		rootNode := parseRootNode(app)
		startingTasks = append(startingTasks, doParseNode(rootNode, app["children"]))
	}
}

func doParseNode(oldParent *StartingNode, rawChildren interface{}) *StartingNode  {
	if rawChildren == nil {
		return oldParent
	}
	children := rawChildren.([]interface{})

	// 处理父节点的子节点
	for _, rawChild := range children {
		child := rawChild.(map[interface{}]interface{})
		childStarting := parseSingleNode(oldParent, child)
		oldParent.children = append(oldParent.children, childStarting)
		if nextChildren := child["children"]; nextChildren != nil {
			doParseNode(childStarting, nextChildren)
		}
	}

	return oldParent
}

func parseRootNode(rootMap map[interface{}]interface{}) *StartingNode {
	ret := parseSingleNode(nil, rootMap)
	// 生成remote信息
	ret.remote = parseRemote(rootMap)
	return ret
}

func parseSingleNode(parent *StartingNode, nodeMap map[interface{}]interface{}) (ret *StartingNode) {
	ret = &StartingNode{}
	// 是否是本地进程
	if location := getLocation(parent, nodeMap); location != nil {
		ret.location = location
	}
	name := getNodeName(nodeMap)
	if name == "" {
		log.Panicf("node must have a name, Parent:%s", parent.name)
	}
	// 名字是否重复
	ret.name = name
	if !setConfig(ret) {
		log.Panicf("duplicated name:%s", name)
	}
	// sup flag
	ret.supFlag = getSupFlag(nodeMap)
	if parent == nil {
		// 如果没有根节点,则认为是application,没有sup flag
		ret.fullName = name
	} else {
		// 如果是动态节点,不加自己的名称
		ret.fullName = fmt.Sprintf("%s/%s", parent.fullName, name)
	}

	// spawnSpec
	ret.spawnSpec = getNodeSpawnSpec(parent, nodeMap)
	ret.parent = parent

	return ret
}

func getNodeName(nodeMap map[interface{}]interface{}) string  {
	node := getNode(nodeMap)
	return node["name"].(string)
}

func getSupFlag(nodeMap map[interface{}]interface{}) *SupervisorFlag {
	if rawSup, ok := nodeMap["sup"]; ok {
		name := getNodeName(nodeMap)
		if producer, ok := registerInfoMap[name]; ok {
			// 只有Supervisor才可以有sup flag
			var ifSup bool
			switch producer.producer.(type) {
			case SupervisorBehaviourProducer:
				ifSup = true
			case ApplicationBehaviourProducer:
				ifSup = true
			}
			if ifSup {
				sup := rawSup.(map[interface{}]interface{})
				directive := getDirective(sup)
				ret := &SupervisorFlag{
					MaxRetries: getMaxRetries(sup),
					WithinDuration: getDuration(sup),
					Decider: func(reason interface{}) actor.Directive {
						return directive
					},
					StrategyFlag: getStrategyFlag(sup),
				}

				return ret

			}
		}
	}

	return nil
}

// 获取重启策略, 默认:OneForOne
func getStrategyFlag(node map[interface{}]interface{}) StrategyFlag {
	if stg, ok := node["strategy"]; ok {
		strategy := stg.(string)
		switch strings.ToLower(strategy) {
		case "one_for_one":
			return OneForOne
		case "all_for_one":
			return AllForOne
		case "simple_one_for_one":
			return SimpleOneForOne
		default:
			return OneForOne
		}
	}

	return OneForOne
}

// 获取监控指令,默认是restart
func getDirective(node map[interface{}]interface{}) actor.Directive {
	if directive, ok := node["directive"]; ok {
		d := directive.(string)
		switch strings.ToLower(d) {
		case "resume":
			return actor.ResumeDirective
		case "restart":
			return actor.RestartDirective
		case "stop":
			return actor.StopDirective
		case "escalate":
			return actor.EscalateDirective
		default:
			return actor.RestartDirective
		}
	}

	return actor.RestartDirective
}

func getDuration(node map[interface{}]interface{}) time.Duration {
	if du, ok := node["within_duration"]; ok {
		return time.Duration(du.(int))
	}

	return 10_000_000_000
}

// 获取重启次数,默认: 10
func getMaxRetries(node map[interface{}]interface{}) int {
	if re, ok := node["max_retries"]; ok {
		return re.(int)
	}

	return 10
}

func getNodeSpawnSpec(parent *StartingNode, nodeMap map[interface{}]interface{}) *SpawnSpec {
	spec := NewDefaultSpawnSpec()
	name := getNodeName(nodeMap)
	// 根据名字获注册的producer
	if info, ok := getRegisterInfo(name); ok {
		// 检查producer类型是否正确
		if _, ok := nodeMap["sup"]; ok {
			switch info.producer.(type) {
			case ApplicationBehaviourProducer:
				// Application,不能有父节点
				if parent != nil {
					log.Panic("Application node must be the root node")
				}
			case SupervisorBehaviourProducer:
			default:
				log.Panicf("info of sup must be: SupervisorBehaviourProducer")
			}
		} else {
			if _, ok := info.producer.(GeneralServerBehaviourProducer); !ok {
				log.Panicf("info of normal server must be: GeneralServerBehaviourProducer")
			}
		}
	} else {
		log.Panicf("faild to find registered info of %s", name)
	}
	// 启动参数
	node := getNode(nodeMap)
	if args, ok := node["args"]; ok {
		spec.Args = args
	}
	// SpawnTimeout
	if spawnTimeout, ok := node["spawn_timeout"]; ok {
		spec.SpawnTimeout = spawnTimeout.(time.Duration)
	} else {
		spec.SpawnTimeout = 10_000_000_000
	}
	// receive_timeout
	if rtimeout, ok := node["receive_timeout"]; ok {
		spec.ReceiveTimeout = rtimeout.(time.Duration)
	}
	spec.Id = name
	return spec
}

func getNode(nodeMap map[interface{}]interface{}) map[interface{}]interface{} {
	if sup, ok := nodeMap["sup"]; ok {
		return sup.(map[interface{}]interface{})
	}
	if ser, ok := nodeMap["server"]; ok {
		return ser.(map[interface{}]interface{})
	}
	return nodeMap
}

func getLocation(parent *StartingNode, m map[interface{}]interface{}) *location {
	if parent != nil && parent.location != nil {
		return parent.location
	}

	// 目前location只对sup有效
	//if _, ok := m["sup"]; !ok {
	//	return nil
	//}

	if loc, ok := m["location"]; ok {
		l := loc.(map[interface{}]interface{})
		host := l["host"]
		port := l["port"]

		return &location{
			host: host.(string),
			port: port.(int),
		}
	}

	return nil
}

func generateSpace(ret string, depth int) string {
	if ret == "" {
		ret = "|"
	}
	if depth <= 0 {
		return ret
	}
	return generateSpace(ret + "__", depth -1)
}

// 通过注册名获取对应的配置信息
func getConfigByName(name string) (*StartingNode, bool) {
	if config, ok := serversMap[name]; ok {
		return config, true
	} else {
		return nil, false
	}
}

func setConfig(node *StartingNode) bool {
	if _, ok := getConfigByName(node.name); ok {
		return false
	} else {
		serversMap[node.name] = node
		return true
	}
}

func isDynamic(node *StartingNode) bool {
	return node.parent != nil && node.parent.supFlag != nil && node.parent.supFlag.StrategyFlag == SimpleOneForOne
}

func makeStartFun(info *registerInfo) SpawnFun {
	if info.startFun != nil {
		return info.startFun
	}

	if config, ok := getConfigByName(info.id); ok {
		if config.parent == nil {
			log.Warnf("got no parent when make start fun, id: %s", info.id)
			return func(parent actor.SpawnerContext, props *actor.Props, args interface{}) (pid *actor.PID, err error) {
				return parent.SpawnNamed(props, info.id)
			}
		} else {
			if config.parent.supFlag.StrategyFlag == SimpleOneForOne {
				return func(parent actor.SpawnerContext, props *actor.Props, args interface{}) (pid *actor.PID, err error) {
					return parent.Spawn(props), nil
				}
			} else {
				return func(parent actor.SpawnerContext, props *actor.Props, args interface{}) (pid *actor.PID, err error) {
					return parent.SpawnNamed(props, info.id)
				}
			}
		}
	} else {
		log.Warnf("got no config when make start fun, id: %s", info.id)
		return func(parent actor.SpawnerContext, props *actor.Props, args interface{}) (pid *actor.PID, err error) {
			return parent.Spawn(props), nil
		}
	}
}

/*
分析命令行
因为希望只处理go-rigger关心的选项,所以未使用 flag包来处理命令行参数
 */
func parseCl() (launchConfig *string, appName *string, appConfig *string) {
	for idx := 1; idx < len(os.Args); idx += 1 {
		cmd := os.Args[idx]
		switch cmd {
		case "-l":
			launchConfig = &os.Args[idx + 1]
			idx += 1
		case "-c":
			appConfig = &os.Args[idx + 1]
			idx += 1
		case "-n":
			appName = &os.Args[idx + 1]
			idx += 1
		}
	}

	return
}

func readAppConfig(appConfigPath string)  {
	// 读取应用配置文件
	if appConfigPath != "" {
		viper.SetConfigFile(appConfigPath)
		if err := viper.ReadInConfig(); err != nil {
			log.Panicf("error when reading config: %s, reason: %s", appConfigPath, err.Error())
		}
	}
}
