package qlog

import (
	"fmt"
	"github.com/VukoDrakkeinen/Quasar/datadir"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type msgSeverity int

func (this msgSeverity) String() string {
	switch this {
	case Info:
		return "Info"
	case Warning:
		return "WARNING"
	case Error:
		return "ERROR"
	}
	return "-UNKNOWN-"
}

type logMessage struct {
	s msgSeverity
	t time.Time
	m string
}

const (
	Info msgSeverity = iota
	Warning
	Error
)
const (
	logTimeFormat = "2006-01-02 15:04:05"
)

type LogWriter interface {
	Write(logMessage) error
}

type QLogger struct {
	writers []LogWriter
	lock    sync.Mutex
}

type FileLog struct {
	file     *os.File
	newLined bool
}

type StdLog struct{}
type NullLog struct{}

func (this *FileLog) Write(msg logMessage) (err error) {
	if !this.newLined {
		_, err = this.file.WriteString("\n")
		if err != nil {
			return err
		}
		this.newLined = true
	}
	bstr := []byte(msg.t.Format(logTimeFormat))
	bstr = append(bstr, []byte(" "+msg.s.String()+": [")...)
	bstr = append(bstr, []byte(msg.m+"]\n")...)
	_, err = this.file.Write(bstr)
	return err
}

func (this *StdLog) Write(msg logMessage) (err error) {
	switch msg.s {
	case Info:
		_, err = fmt.Println(msg.m)
	case Warning:
		_, err = fmt.Println(msg.s.String()+":", msg.m)
	case Error:
		_, err = fmt.Fprintln(os.Stderr, msg.s.String()+":", msg.m)
	}
	return err
}

func (this *NullLog) Write(msg logMessage) error { return nil }

var defaultLogger QLogger
var logFilenames map[string]struct{}
var cLock sync.Mutex

func init() {
	logFilenames = make(map[string]struct{})
	defaultLogger = *New(&StdLog{})
	defaultLogger.AddWriter(NewFileLog("debug.log"))
}

func NewFileLog(filename string) LogWriter {
	cLock.Lock()
	defer cLock.Unlock()
	if _, exists := logFilenames[filename]; exists {
		Log(Warning, "Attempted to create another LogWriter for file", filename) //...we won't hit an infinite recurrence, will we?
		return &NullLog{}
	}

	path := filepath.Join(datadir.Logs(), filename)
	rotateLogs(path)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		Log(Error, `Unable to open log file "`, filename, `".`)
		return &NullLog{}
	}
	ret := &FileLog{file: file}
	logFilenames[filename] = struct{}{}
	return ret
}

func New(writers ...LogWriter) *QLogger {
	ret := &QLogger{
		writers: make([]LogWriter, 0, len(writers)),
	}
	for _, writer := range writers {
		ret.AddWriter(writer)
	}
	return ret
}

func (this *QLogger) AddWriter(writer LogWriter) {
	if _, isNull := writer.(*NullLog); !isNull {
		this.writers = append(this.writers, writer)
	}
} //TODO?: RemoveWriter?

func rotateLogs(path string) {
	if info, err := os.Stat(path); err == nil && info.Size() > 512*1024 {
		os.Remove(path + ".9")
		for i := int64(8); i > 0; i-- {
			os.Rename(path+strconv.FormatInt(i-1, 10), path+strconv.FormatInt(i, 10))
		}
		os.Rename(path, path+".1")
	}
}

func (this *QLogger) Log(s msgSeverity, what ...interface{}) (err error) {
	m := fmt.Sprintln(what...)
	msg := logMessage{s, time.Now().UTC(), m[:len(m)-1]}
	for _, writer := range this.writers {
		werr := writer.Write(msg)
		if err == nil {
			err = werr
		}
	}
	return err
}

func (this *QLogger) Logf(s msgSeverity, format string, what ...interface{}) (err error) {
	msg := logMessage{s, time.Now().UTC(), fmt.Sprintf(format, what...)}
	for _, writer := range this.writers {
		werr := writer.Write(msg)
		if err == nil {
			err = werr
		}
	}
	return err
}

func Log(s msgSeverity, what ...interface{}) error {
	return defaultLogger.Log(s, what...)
}

func Logf(s msgSeverity, format string, what ...interface{}) error {
	return defaultLogger.Logf(s, format, what...)
}

/*

func (this *BufferedLog) Write(msg logMessage) {
	if len(this.msgs) != cap(this.msgs) {
		this.msgs = append(this.msgs, msg)
	} else {
		copy(this.msgs, this.msgs[1:])
		this.msgs[len(this.msgs)-1] = msg
	}
}

func (this *BufferedLog) SetBufferSize(i int) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if cap(this.msgs) > i {
		newMsgs := make([]logMessage, i)
		copy(newMsgs, this.msgs[int(math.Dim(float64(len(this.msgs), float64(i)))):])
		this.msgs = newMsgs
	} else if cap(this.msgs) < i {
		newMsgs := make([]logMessage, 0, i)
		newMsgs = append(newMsgs, this.msgs...)
		this.msgs = newMsgs
	}
}

func (this *BufferedLog) BufferSize() int {
	this.lock.Lock()
	defer this.lock.Unlock()
	return cap(this.msgs)
}
//*/
