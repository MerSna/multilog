package multilog

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	//"go.uber.org/zap"
)

var (
	_golabMu = &sync.Mutex{}
	_golabL  = _globalLogger()

	//_connMu = &sync.Mutex{}
)

const (
	DebugLevel = iota
	InfoLevel
	ErrorLevel
	CriticalLevel
)

var prefixTagMap = map[int]string{
	0: "[Debug]",
	1: "[Info]",
	2: "[Error]",
	3: "[Critical]",
}

type DatabaseConfig struct {
	DB        *sql.DB
	TableName string
}

type Logger struct {
	Conn       net.Conn
	DBConf     *DatabaseConfig
	RedisCli   *redis.Client
	Outer      *os.File
	Level      int
	StrictMode bool
}

type logInfo struct {
	PrefixTag string
	Time      string
}

func _globalLogger() *Logger {
	l := &Logger{
		Conn:       nil,
		DBConf:     nil,
		RedisCli:   nil,
		Outer:      os.Stdout,
		StrictMode: true,
	}
	return l
}

func G() *Logger {
	_golabMu.Lock()
	g := _golabL
	_golabMu.Unlock()
	return g
}

func (l *Logger) redisWrite(k, v string) {
	if l.RedisCli == nil {
		return
	}
	l.RedisCli.RPush(nil, k, v)
}

func (l *Logger) sqlWrite(datestr, prefix, callpath, context string) {
	if l.DBConf == nil || l.DBConf.DB == nil {
		return
	}
	rows, err := l.DBConf.DB.Query(fmt.Sprintf(
		"INSERT INTO %s (datestr, prefix, callpath, context) VALUES ('%s', '%s', '%s', '%s')",
		l.DBConf.TableName,
		datestr, prefix, callpath, context,
	))
	if err != nil {
		return
	}
	_ = rows.Close()

}

func (l *Logger) output(calldepth int, prefixTag int, str string) {
	if prefixTag < l.Level {
		return
	}
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}
	//
	t := time.Now()
	tstring := fmt.Sprintf("%d-%d-%d %d:%d:%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())

	// set short
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	file = short
	info := fmt.Sprintf("%s %s %s:%d %s", prefixTagMap[prefixTag], tstring, file, line, str)
	_, err := l.Outer.WriteString(info)
	if err != nil {
		return
	}

	var wg = &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		l.redisWrite(tstring, info)
		wg.Done()
	}()
	go func() {
		l.sqlWrite(tstring, prefixTagMap[prefixTag], fmt.Sprintf("%s:%d", file, line), str)
		wg.Done()
	}()
	if l.StrictMode {
		wg.Wait()
	}
	return
}

func (l *Logger) Debug(v interface{})    { l.output(2, 0, fmt.Sprintf("%v\n", v)) }
func (l *Logger) Info(v interface{})     { l.output(2, 1, fmt.Sprintf("%v\n", v)) }
func (l *Logger) Error(v interface{})    { l.output(2, 2, fmt.Sprintf("%v\n", v)) }
func (l *Logger) Critical(v interface{}) { l.output(2, 3, fmt.Sprintf("%v\n", v)) }
