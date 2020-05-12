/*
* description: 异步输出本地日志
* author: yebz
* date: 2019-06-10
 */
package log

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
	"time"
)

type Log interface {
	/*新建一个日志对象*/
	NewLog(rootPath string, level ...int) Logger

	NewLogger(rootPath string, level ...int) Logger

	/*设置单个日志文件最大大小*/
	SetLogSize(s int64)

	/*直接输出日志，不带任何格式*/
	Simple(args ...interface{})

	/*debug级别日志输出*/
	Debug(args ...interface{})
	/*debug级别格式化日志输出*/
	Debugf(format string, args ...interface{})

	/*info级别日志输出*/
	Info(args ...interface{})
	/*info级别格式化日志输出*/
	Infof(format string, args ...interface{})

	/*warn级别日志输出*/
	Warn(args ...interface{})
	/*warn级别格式化日志输出*/
	Warnf(format string, args ...interface{})

	/*error级别日志输出*/
	Error(args ...interface{})
	/*error级别格式化日志输出*/
	Errorf(format string, args ...interface{})
}

const (
	black colorAttribute = iota + 30
	red
	green
	yellow
	blue
	magenta
	cyan
	white
)

const (
	LoggerLevelDebug = iota
	LoggerLevelInfo
	LoggerLevelWarn
	LoggerLevelError
)

const (
	defaultCallDepth int = 2
	defaultLogLevel  int = LoggerLevelDebug
)

var (
	LogFileCount int   = 30
	MaxSize      int64 = 100 * 1024 * 1024
)

type colorAttribute int

func color(s string, c colorAttribute) string {
	return fmt.Sprintf("\u001b[%vm%s\u001b[0m", c, s)
}

var (
	log *Logger
)

type logInfo struct {
	level    int
	isSimple bool
	logmsg   string
	line     int
	file     string
}

type Logger struct {
	rootPath     string    // 绝对路径
	file         *os.File  // 日志文件
	level        int       // 级别
	depth        int       // 深度: 2
	nextDay      time.Time // 下一次创建文件的时间
	nowFile      string
	nowFileCount int
	PrefixHeader string
	logchan      chan *logInfo
}

func SetLogSize(s int64) {
	MaxSize = s
}

func SetLogCount(n int) {
	LogFileCount = n
}

func Simple(args ...interface{}) {
	if log == nil {
		NewLogger("")
	}
	if defaultLogLevel > log.level {
		return
	}

	log.logchan <- &logInfo{
		level:    0,
		logmsg:   fmt.Sprint(args...),
		isSimple: true,
		file:     "",
		line:     0,
	}
}

func Debug(args ...interface{}) {
	if log == nil {
		NewLogger("")
	}
	if LoggerLevelDebug < log.level {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok == false {
		return
	}
	log.logchan <- &logInfo{
		level:    LoggerLevelDebug,
		logmsg:   fmt.Sprint(args...),
		isSimple: false,
		file:     file,
		line:     line,
	}
}

func Debugf(format string, args ...interface{}) {
	if log == nil {
		NewLogger("")
	}
	if LoggerLevelDebug < log.level {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok == false {
		panic(errors.New("获取行数失败"))
	}

	log.logchan <- &logInfo{
		level:    LoggerLevelDebug,
		logmsg:   fmt.Sprintf(format, args...),
		isSimple: false,
		file:     file,
		line:     line,
	}
}

func Info(args ...interface{}) {
	if log == nil {
		NewLogger("")
	}
	if LoggerLevelInfo < log.level {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok == false {
		return
	}
	log.logchan <- &logInfo{
		level:    LoggerLevelInfo,
		logmsg:   fmt.Sprint(args...),
		isSimple: false,
		file:     file,
		line:     line,
	}
}

func Infof(format string, args ...interface{}) {
	if log == nil {
		NewLogger("")
	}
	if LoggerLevelInfo < log.level {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok == false {
		return
	}
	log.logchan <- &logInfo{
		level:    LoggerLevelInfo,
		logmsg:   fmt.Sprintf(format, args...),
		isSimple: false,
		file:     file,
		line:     line,
	}
}

func Warn(args ...interface{}) {
	if log == nil {
		NewLogger("")
	}
	if LoggerLevelWarn < log.level {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok == false {
		return
	}
	log.logchan <- &logInfo{
		level:    LoggerLevelWarn,
		logmsg:   fmt.Sprint(args...),
		isSimple: false,
		file:     file,
		line:     line,
	}
}

func Warnf(format string, args ...interface{}) {
	if log == nil {
		NewLogger("")
	}
	if LoggerLevelWarn < log.level {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok == false {
		return
	}
	log.logchan <- &logInfo{
		level:    LoggerLevelWarn,
		logmsg:   fmt.Sprintf(format, args...),
		isSimple: false,
		file:     file,
		line:     line,
	}
}

func Error(args ...interface{}) {
	if log == nil {
		NewLogger("")
	}
	if LoggerLevelError < log.level {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok == false {
		return
	}
	log.logchan <- &logInfo{
		level:    LoggerLevelError,
		logmsg:   fmt.Sprint(args...),
		isSimple: false,
		file:     file,
		line:     line,
	}
}

func Errorf(format string, args ...interface{}) {
	if log == nil {
		NewLogger("")
	}
	if LoggerLevelError < log.level {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok == false {
		return
	}
	log.logchan <- &logInfo{
		level:    LoggerLevelError,
		logmsg:   fmt.Sprintf(format, args...),
		isSimple: false,
		file:     file,
		line:     line,
	}
}

func Write(file *os.File, content string) (bool, error) {
	_, err := file.WriteString(content)

	if err != nil {
		return false, err
	}
	return true, nil
}

func NewLog(rootPath string, level ...int) Logger {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	l := Logger{}
	l.depth = defaultCallDepth
	l.rootPath = rootPath
	l.level = defaultLogLevel
	l.logchan = make(chan *logInfo, 1000)

	var levelEnum = 0
	if len(level) > 0 {
		levelEnum = level[0]
		if levelEnum != LoggerLevelDebug &&
			levelEnum != LoggerLevelInfo &&
			levelEnum != LoggerLevelWarn &&
			levelEnum != LoggerLevelError {
			panic("等级不存在")
		}
		l.level = levelEnum
	}

	l.logAsyncWrite()
	return l
}

func NewFlowLogger(rootPath string, level ...int) Logger {
	return NewLog(rootPath, level...)
}

func NewLogger(rootPath string, level ...int) Logger {
	*log = NewLog(rootPath, level...)
	return *log
}

func (this *Logger) logAsyncWrite() {
	for index := 0; index < 2; index++ {
		go func() {
			for {
				logmsg := <-this.logchan
				this.WriteLogFormat(logmsg.level, logmsg.logmsg, logmsg.isSimple, logmsg.file, logmsg.line)
			}
		}()
	}
}

func (this *Logger) SetCallDepth(depth int) {
	if depth > 0 {
		this.depth = depth
	}
}

func (this *Logger) GetLogFile() error {
	rootPath := this.rootPath
	flag, err := this.isFileExist(rootPath)

	if len(rootPath) == 0 {
		return nil
	}

	if err != nil {
		panic(err)
	}

	if flag == false {
		os.MkdirAll(rootPath, os.ModeDir)
	}

	this.removeSurplusFile()

	date := time.Unix(time.Now().Unix(), 0).Format("2006-01-02")
	nextD := time.Unix(time.Now().Unix()+(24*3600), 0)
	nextD = time.Date(nextD.Year(), nextD.Month(), nextD.Day(), 0, 0, 0, 0, nextD.Location())
	this.nextDay = nextD

	logPath := fmt.Sprintf("%s/%s%s.log", rootPath, this.PrefixHeader, date)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if f == nil {
		return errors.New("log文件打开失败")
	}

	this.file = f
	this.nowFile = logPath
	return err
}

func (this *Logger) fileTooBigToCut() {
	if s, err := this.fileSize(this.nowFile); err == nil {
		if s > MaxSize {
			os.Rename(this.nowFile, fmt.Sprintf("%s.%v", this.nowFile, time.Now().Format("2006_01_02_15_04_05")))

			f, err := os.OpenFile(this.nowFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
			if err != nil || f == nil {
				return
			}

			this.file = f
			this.removeSurplusFile()
		}
	}
}

func (this *Logger) removeSurplusFile() {
	dir, err := os.Open(this.rootPath)
	if err != nil {
		return
	}
	defer dir.Close()

	fis, err := dir.Readdir(0)
	if err != nil {
		return
	}

	var files []string
	for _, fi := range fis {
		name := fi.Name()
		if strings.Contains(name, ".log") {
			if len(this.PrefixHeader) > 0 {
				if strings.HasPrefix(name, this.PrefixHeader) {
					files = append(files, name)
				}
			} else {
				if strings.HasPrefix(name, "2") {
					files = append(files, name)
				}
			}
		}
	}
	if len(files) > LogFileCount {
		s := sort.StringSlice(files)
		sort.Sort(s)

		surcount := len(files) - LogFileCount
		for index := 0; index < surcount; index++ {
			f := s[index]
			os.Remove(path.Join(this.rootPath, f))
		}
	}
}

// 格式化的写入日志,level是一个枚举,如LoggerLevelError,log是日志字符串
func (this *Logger) WriteLogFormat(level int, logstr string, isSimple bool, file string, line int) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	// 时间
	now := time.Now()
	if now.Unix() > this.nextDay.Unix() { // 超过了原定的下次创建时间, 重新创建一个文件
		if err := this.GetLogFile(); err != nil {
			panic(err)
		}
	} else {
		this.fileTooBigToCut()
	}

	time := time.Unix(now.Unix(), 0).Format("2006-01-02 15:04:05")

	var (
		logstring string
	)

	if isSimple {
		logstring = fmt.Sprintf("%s\n", logstr)
	} else {
		var (
			flag string
		)

		switch level {
		case LoggerLevelDebug:
			flag = color("DEBUG", blue)
		case LoggerLevelInfo:
			flag = color("INFO", green)
		case LoggerLevelWarn:
			flag = color("WARN", yellow)
		case LoggerLevelError:
			flag = color("ERROR", red)
		}

		if v := strings.Split(file, "/"); len(v) > 0 {
			file = v[len(v)-1]
		}
		logstring = fmt.Sprintf("%s[%s][%s:%d]  %s\n", time, flag, file, line, logstr)
	}

	if len(this.rootPath) == 0 {
		fmt.Printf(logstring)
	} else {
		sta, _ := this.isFileExist(this.nowFile)
		if !sta {
			fmt.Println("log fail not found")
			if this.file != nil {
				this.file.Close()
				this.file = nil
			}

			err := this.GetLogFile()
			if err != nil {
				fmt.Println("GetLogFile fail", err)
				return
			}
		}
		b, err := Write(this.file, logstring)
		if err != nil {
			panic(err)
		}

		if !b {
			fmt.Println("write log file fail")
		}
	}
}

func (this *Logger) WriteSimple(log string) {
	if len(this.rootPath) == 0 {
		fmt.Printf("%s\n", log)
	} else {
		if this.file != nil {
			_, err := Write(this.file, fmt.Sprintf("%s\n", log))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (this *Logger) fileSize(file string) (int64, error) {
	f, err := os.Stat(file)
	if err != nil {
		return 0, err
	}
	return f.Size(), nil
}

func (this *Logger) isFileExist(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
