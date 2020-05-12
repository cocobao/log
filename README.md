# Log


### Get
```
go get github.com/cocobao/log
```

### 特点
1. 异步输出日志
2. 自动按天分割文件
3. 自动按大小分割文件

### 用法
```
import "github.com/cocobao/log"

如果只想输出到控制台，直接调用即可
log.Debug("hello")
log.Debugf("%s", "hello")
log.Info("hello")
log.Infof("hello")
log.Warn("hello")
log.Warnf("%s", "hello")
log.Error("hello")
log.Errorf("%s", "hello")

如果想输出到文件，则需要初始化并带上日志路径，日志会自动每天一个文件，并且超过100M自动分割
log.NewLogger("./log")
Debug("hello")
```