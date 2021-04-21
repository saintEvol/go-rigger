package rigger

import "time"

// 新创建一个 SpawnSpec 结构, SpawnSpec.SpawnTimeout 默认值为预定义值: startTimeOut (10_000_000_000 ns)
func NewSpawnSpec() *SpawnSpec {
	return &SpawnSpec{
		Args:           nil,
		SpawnTimeout:   startTimeOut,
		ReceiveTimeout: 0,
		Kind:           "",
	}
}

/*
	生成一个用于 SimpleOneForOne 模式的 SpawnSpec
*/
func SimpleSpawnSpec(name string, args interface{}) *SpawnSpec {
	return NewSpawnSpec().WithName(name).WithArgs(args)
}

// 根据进程类型(kind) 生成一个默认的SpawnSpec
func SpawnSpecWithKind(kind string /*进程种类*/) *SpawnSpec {
	return makeDefaultSpawnSpec(kind)
}

// 生成一个命名的启动规范
func SpawnSpecWithName(kind string/*进程种类*/, name string/*启动后的进程名*/) *SpawnSpec {
	spec := makeDefaultSpawnSpec(kind)
	spec.Name = name

	return spec
}


/* 启动规范,描述了如何启动一个进程
如果父进程的监控策略是 SimpleOneForOne 且是通过接口: StartChild, StartChildNotified, StartChildSync 等接口启动子进程时,
此时参数中的 *SpawnSpec 只有 SpawnSpec.Name, SpawnSpec.Args 这两个字段有意义, 其余字段的数据会被简单丢弃,并使用父进程规定的对应值
*/
type SpawnSpec struct {
	/*
		进程名,如果此值不为空,则使用此名字注册进程, 如果为空,则:
		1. 如果是动态启动的进程, rigger 不会自动注册进程,此时,如果用户想要注册,则应该rigger.RegisterStartFun注册启动函数,
		   并在启动函数中使用SpawnNamed来规定注册名
		2. 如果是静态进程, 则会使用 Kind 作为注册名
		3. Application只有rigger能启动, rigger会以Application的 Kind 作为其它注册名
	*/
	Name string
	Kind string // Kind, 必选参数, 框架会根据此ID查询启动 Producer和StartFun
	//Producer     interface{} // props producer,生成一个对应行为模式的实例,也即行为模式工厂
	//Starter      SpawnFun
	Args         interface{}   // 启动参数, 原样传入 Starter, 与 LifeCyclePart.OnStarted
	SpawnTimeout time.Duration // 超时时间,如果为0表示不等待,也即异步启动
	// 平静期超时时间,如果在指定时间内没收到任何消息,则会触发TimeroutReceiver回调,此值不为0时,需要实现TimeoutReceiver
	ReceiveTimeout time.Duration
	//IsGlobal bool // 该进程是否是全局进程, 对于全局进程,在同一个集群中,只允许存在唯一一个同名进程

	isFromConfig bool // 是否是从配置启动
}

func (ss *SpawnSpec) WithKind(kind string) *SpawnSpec {
	ss.Kind = kind
	return ss
}

func (ss *SpawnSpec) WithName(name string) *SpawnSpec {
	ss.Name = name
	return ss
}

func (ss *SpawnSpec) WithArgs(args interface{}) *SpawnSpec {
	ss.Args = args
	return ss
}

func (ss *SpawnSpec) WithSpawnTimeout(duration time.Duration) *SpawnSpec {
	ss.SpawnTimeout = duration
	return ss
}

func (ss *SpawnSpec) WithReceiveTimeout(duration time.Duration) *SpawnSpec {
	ss.ReceiveTimeout = duration
	return ss
}

//func (ss *SpawnSpec) WithGlobal(isGlobal bool) *SpawnSpec {
//	ss.IsGlobal = isGlobal
//	return ss
//}
func (ss *SpawnSpec) BelongThisNode() bool {
	var name string
	if ss.Name != "" {
		name = ss.Name
	} else {
		name = ss.Kind
	}

	if name == "" {
		return false
	}

	return belongThisNode(name)
}

func filterSpawnSpecOfThisNode(specs []*SpawnSpec) []*SpawnSpec {
	var ret []*SpawnSpec
	for _, spec := range specs {
		if spec.BelongThisNode() {
			ret = append(ret, spec)
		}
	}

	return ret
}

