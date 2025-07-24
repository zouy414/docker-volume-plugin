package log

import (
	"fmt"
	"log"
	"os"
)

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

const (
	debugPrefix string = " [DEBUG] "
	infoPrefix  string = " [INFO] "
	warnPrefix  string = " [WARNING] "
	errorPrefix string = " [ERROR] "
	fatalPrefix string = " [FATAL] "
)

// New logger
func New(service string) *Logger {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	return &Logger{
		logLevel: DebugLevel,
		service:  service,
		logger:   logger,
	}
}

// Logger ...
type Logger struct {
	logLevel LogLevel
	service  string
	logger   *log.Logger
}

// WithService fork a new logger
func (l *Logger) WithService(service string) *Logger {
	return &Logger{
		logLevel: l.logLevel,
		service:  service,
		logger:   l.logger,
	}
}

// WithLogLevel fork a new logger
func (l *Logger) WithLogLevel(logLevel LogLevel) *Logger {
	return &Logger{
		logLevel: logLevel,
		service:  l.service,
		logger:   l.logger,
	}
}

// Debug message
func (l *Logger) Debug(v ...interface{}) {
	if l.logLevel > DebugLevel {
		return
	}

	_ = l.logger.Output(2, fmt.Sprint(append([]interface{}{l.service, debugPrefix}, v...)...))
}

// Debugf message
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.logLevel > DebugLevel {
		return
	}

	_ = l.logger.Output(2, fmt.Sprintf("%s%s"+format, append([]interface{}{l.service, debugPrefix}, v...)...))
}

// Info message
func (l *Logger) Info(v ...interface{}) {
	if l.logLevel > InfoLevel {
		return
	}

	_ = l.logger.Output(2, fmt.Sprint(append([]interface{}{l.service, infoPrefix}, v...)...))
}

// Infof message
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.logLevel > InfoLevel {
		return
	}

	_ = l.logger.Output(2, fmt.Sprintf("%s%s"+format, append([]interface{}{l.service, infoPrefix}, v...)...))
}

// Warn message
func (l *Logger) Warning(v ...interface{}) {
	if l.logLevel > WarnLevel {
		return
	}

	_ = l.logger.Output(2, fmt.Sprint(append([]interface{}{l.service, warnPrefix}, v...)...))
}

// Warnf message
func (l *Logger) Warningf(format string, v ...interface{}) {
	if l.logLevel > WarnLevel {
		return
	}

	_ = l.logger.Output(2, fmt.Sprintf("%s%s"+format, append([]interface{}{l.service, warnPrefix}, v...)...))
}

// Error message
func (l *Logger) Error(v ...interface{}) {
	if l.logLevel > ErrorLevel {
		return
	}

	_ = l.logger.Output(2, fmt.Sprint(append([]interface{}{l.service, errorPrefix}, v...)...))
}

// Errorf message
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.logLevel > ErrorLevel {
		return
	}

	_ = l.logger.Output(2, fmt.Sprintf("%s%s"+format, append([]interface{}{l.service, errorPrefix}, v...)...))
}

// Fatal message
func (l *Logger) Fatal(v ...interface{}) {
	_ = l.logger.Output(2, fmt.Sprint(append([]interface{}{l.service, fatalPrefix}, v...)...))
	os.Exit(1)
}

// Fatalf message
func (l *Logger) Fatalf(format string, v ...interface{}) {
	_ = l.logger.Output(2, fmt.Sprintf("%s%s"+format, append([]interface{}{l.service, fatalPrefix}, v...)...))
	os.Exit(1)
}
