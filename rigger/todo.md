#### 默认startFun的优化,现在使用的是未命名spawn done
#### RemoteStartChildCmd的处理 done

#### 不从配置启动时,也要生成 StartingNode信息,将从配置启动和不从配置启动的过程统一 done
#### 只有不是从配置启动时,才需要生成配置信息, done
#### 跨节点创建进程及获得回应 done

### 为了随意跨进程运行,应该只通过进程获取数据,而不是通过公共变量获取数据
### 启动子进程时,带的参数,不能是函数等
#### OnMessage 返回 NO_REPLY时,将由用户手动回复
