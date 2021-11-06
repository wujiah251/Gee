package log

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
)

/*
[info ]颜色为蓝色
[error]颜色为红色
*/

var (
	errorLog = log.New(os.Stdout, "\033[31m[error]\033][0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info ]\033][0m ", log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{errorLog, infoLog}
	mu       sync.Mutex
)

var (
	Error  = errorLog.Println
	Errorf = errorLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
)

const (
	InfoLevel = iota
	ErrorLevel
	Disabled
)

func SerLevel(level int) {
	mu.Lock()
	defer mu.Unlock()
	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}
	if ErrorLevel < level {
		errorLog.SetOutput(ioutil.Discard)
	}
	if InfoLevel < level {
		infoLog.SetOutput(ioutil.Discard)
	}
}
