package multilog

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	//"go.uber.org/zap"
)

var (
	_golabMu = &sync.Mutex{}
	_golabL  = _globalLogger()

	_connMu = &sync.Mutex{}
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
	Conn      net.Conn
	DB        *DatabaseConfig
	RedisConn *redis.Conn
	Outer     *os.File
	Level     int
}

type logInfo struct {
	PrefixTag string
	Time      string
}

func _globalLogger() *Logger {
	l := &Logger{
		Conn:      nil,
		DB:        nil,
		RedisConn: nil,
		Outer:     os.Stdout,
	}
	return l
}

func G() *Logger {
	_golabMu.Lock()
	g := _golabL
	_golabMu.Unlock()
	return g
}

func (l *Logger) output(calldepth int, prefixTag int, str string) error {
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
		return err
	}
	return nil
}

func (l *Logger) Info(v interface{}) error { return l.output(2, 1, fmt.Sprintf("%v\n", v)) }
