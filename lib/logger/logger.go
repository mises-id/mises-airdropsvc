package logger

import (
	"io"

	"github.com/sirupsen/logrus"
)

// Logrus : implement Logger
type Logrus struct {
	*logrus.Logger
}

// Logger ...
var Logger *logrus.Logger

// GetEchoLogger for e.Logger
func GetEchoLogger() Logrus {
	return Logrus{Logger}
}

var log = logrus.New()

// SetHeader is a stub to satisfy interface
// It's controlled by Logger
func (l Logrus) SetHeader(_ string) {}

// SetPrefix It's controlled by Logger
func (l Logrus) SetPrefix(s string) {}

// Prefix It's controlled by Logger
func (l Logrus) Prefix() string {
	return ""
}

// Output logger output func
func (l Logrus) Output() io.Writer {
	return l.Out
}

// SetOutput change output, default os.Stdout
func (l Logrus) SetOutput(w io.Writer) {
	Logger.SetOutput(w)
}

// Print string log
func Print(i ...interface{}) {
	log.Print(i[0].(string))
}
func Println(i ...interface{}) {
	log.Println(i[0].(string))
}

// Debug string log
func (l Logrus) Debug(i ...interface{}) {
	Logger.Debug(i[0].(string))
}

// Info string log
func (l Logrus) Info(i ...interface{}) {
	Logger.Info(i[0].(string))
}

// Warn string log
func (l Logrus) Warn(i ...interface{}) {
	Logger.Warn(i[0].(string))
}

// Error string log
func (l Logrus) Error(i ...interface{}) {
	Logger.Error(i[0].(string))
}

// Fatal string log
func (l Logrus) Fatal(i ...interface{}) {
	Logger.Fatal(i[0].(string))
}

// Panic string log
func (l Logrus) Panic(i ...interface{}) {
	Logger.Panic(i[0].(string))
}
