package log

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
	"time"
)

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
	defaultCallDepth    int   = 2
	defaultLogFileCount int   = 30
	defaultLogLevel     int   = LoggerLevelDebug
	defaultMaxSize      int64 = 50 * 1024 * 1024
	isLogCache          bool  = false
	logCacheCount       int   = 20
)

type colorAttribute int

func color(s string, c colorAttribute) string {
	return fmt.Sprintf("\u001b[%vm%s\u001b[0m", c, s)
}

var (
	log Logger
)

type logCache struct {
	isCache       bool
	cacheCount    int
	nowCacheCount int
	logBuff       *bytes.Buffer
}

type Logger struct {
	rootPath     string    // desc:	absolute path
	file         *os.File  // desc:	log file
	level        int       // option: 	LoggerLevelDebug\LoggerLevelInfo\LoggerLevelError
	depth        int       // default: 2
	nextDay      time.Time // desc: 	下一次创建文件的时间
	nowFile      string
	nowFileCount int
	lc           logCache
}

func Simple(args ...interface{}) {
	if defaultLogLevel > log.level {
		return
	}
	log.writeSimple(fmt.Sprint(args...))
}

func Debug(args ...interface{}) {
	if LoggerLevelDebug < log.level {
		return
	}
	log.writeLogFormat(LoggerLevelDebug, fmt.Sprint(args...))
}

func Debugf(format string, args ...interface{}) {
	if LoggerLevelDebug < log.level {
		return
	}
	log.writeLogFormat(LoggerLevelDebug, fmt.Sprintf(format, args...))
}

func Info(args ...interface{}) {
	if LoggerLevelInfo < log.level {
		return
	}
	log.writeLogFormat(LoggerLevelInfo, fmt.Sprint(args...))
}

func Infof(format string, args ...interface{}) {
	if LoggerLevelInfo < log.level {
		return
	}
	log.writeLogFormat(LoggerLevelInfo, fmt.Sprintf(format, args...))
}

func Warn(args ...interface{}) {
	if LoggerLevelWarn < log.level {
		return
	}
	log.writeLogFormat(LoggerLevelWarn, fmt.Sprint(args...))
}

func Warnf(format string, args ...interface{}) {
	if LoggerLevelWarn < log.level {
		return
	}
	log.writeLogFormat(LoggerLevelWarn, fmt.Sprintf(format, args...))
}

func Error(args ...interface{}) {
	if LoggerLevelError < log.level {
		return
	}
	log.writeLogFormat(LoggerLevelError, fmt.Sprint(args...))
}

func Errorf(format string, args ...interface{}) {
	if LoggerLevelError < log.level {
		return
	}
	log.writeLogFormat(LoggerLevelError, fmt.Sprintf(format, args...))
}

func Write(file *os.File, content string) (bool, error) {
	_, err := file.WriteString(content)

	if err != nil {
		return false, err
	}
	return true, nil
}

func NewLogger(rootPath string, level ...int) Logger {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	l := Logger{}
	l.depth = defaultCallDepth
	l.rootPath = rootPath
	l.level = defaultLogLevel
	l.lc.isCache = isLogCache
	l.lc.cacheCount = logCacheCount
	l.lc.logBuff = bytes.NewBuffer([]byte{})

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

	err := l.getLogFile()
	if err != nil {
		panic(err)
	}
	log = l
	return l
}

//设置被调用深度
func (this *Logger) SetCallDepth(depth int) {
	if depth > 0 {
		this.depth = depth
	}
}

//打开日志文件
func (this *Logger) getLogFile() error {
	rootPath := this.rootPath
	flag, err := this.isFileExist(rootPath)

	if len(rootPath) == 0 {
		return nil
	}

	if err != nil {
		panic(fmt.Errorf("get file exist state fail,%v", err))
	}

	if flag == false {
		os.MkdirAll(rootPath, os.ModeDir)
	}

	this.removeSurplusFile()

	date := time.Unix(time.Now().Unix(), 0).Format("2006-01-02")
	nextD := time.Unix(time.Now().Unix()+(24*3600), 0)
	nextD = time.Date(nextD.Year(), nextD.Month(), nextD.Day(), 0, 0, 0, 0, nextD.Location())
	this.nextDay = nextD

	logPath := fmt.Sprintf("%s/%s.log", rootPath, date)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if f == nil {
		return errors.New("log文件打开失败")
	}

	this.file = f
	this.nowFile = logPath
	return err
}

//文件切割
func (this *Logger) fileTooBigToCut() {
	if s, err := this.fileSize(this.nowFile); err == nil {
		if s > defaultMaxSize {
			now := time.Now().Unix()
			os.Rename(this.nowFile, fmt.Sprintf("%s.%v", this.nowFile, now))

			f, err := os.OpenFile(this.nowFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
			if err != nil || f == nil {
				return
			}
			this.removeSurplusFile()
			this.file = f
		}
	}
}

//删除最旧的一个日志文件
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
			files = append(files, name)
		}
	}
	if len(files) > defaultLogFileCount {
		s := sort.StringSlice(files)
		sort.Sort(s)

		surcount := len(files) - defaultLogFileCount
		for index := 0; index < surcount; index++ {
			f := s[index]
			os.Remove(path.Join(this.rootPath, f))
		}
	}
}

// 格式化的写入日志,level是一个枚举,如LoggerLevelError,log是日志字符串
func (this *Logger) writeLogFormat(level int, log string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("writeLog fail,", err)
		}
	}()

	// 时间
	now := time.Now()
	if now.Unix() > this.nextDay.Unix() { // 超过了原定的下次创建时间, 重新创建一个文件
		if err := this.getLogFile(); err != nil {
			panic(fmt.Errorf("getLogFile fail,%s", err.Error()))
		}
	} else {
		this.fileTooBigToCut()
	}

	time := time.Unix(now.Unix(), 0).Format("2006-01-02 15:04:05")

	var flag string

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

	_, file, line, ok := runtime.Caller(this.depth)
	if ok == false {
		panic(errors.New("获取行数失败"))
	}
	if v := strings.Split(file, "/"); len(v) > 0 {
		file = v[len(v)-1]
	}

	writeLog := fmt.Sprintf("%s[%s][%s:%d]  %s\n", time, flag, file, line, log)

	//日志缓存输出
	if this.lc.isCache {
		this.lc.logBuff.WriteString(writeLog)
		if this.lc.nowCacheCount++; this.lc.nowCacheCount > this.lc.cacheCount {
			writeLog = this.lc.logBuff.String()
			this.lc.logBuff = bytes.NewBuffer([]byte{})
			this.lc.nowCacheCount = 0
		} else {
			return
		}
	}

	if len(this.rootPath) == 0 {
		fmt.Printf(writeLog)
	} else {
		sta, _ := this.isFileExist(this.nowFile)
		if !sta {
			//日志文件被人删除了，重新创建一个
			// fmt.Println("log file not found")
			if this.file != nil {
				this.file.Close()
				this.file = nil
			}

			err := this.getLogFile()
			if err != nil {
				fmt.Println("getLogFile fail", err)
				return
			}
		}
		b, err := Write(this.file, writeLog)
		if err != nil {
			panic(fmt.Errorf("write fail,%v", err))
		}

		if !b {
			fmt.Println("write log file fail")
		}
	}
}

func (this *Logger) writeSimple(log string) {
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
