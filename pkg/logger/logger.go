package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

var infoLogger *log.Logger
var errorLogger *log.Logger
var logFile *os.File

func init() {
	infoLogger = log.New(io.Discard, "", 0)
	errorLogger = log.New(io.Discard, "", 0)
}

func Init(verbose bool, path string) error {
	_ = Close()

	if !verbose {
		infoLogger.SetOutput(io.Discard)
		errorLogger.SetOutput(io.Discard)
		return nil
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error while opening/creating a file: %w", err)
	}
	logFile = file

	infoLogger = log.New(file, "[INFO]: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(file, "[ERROR]: ", log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

func Info(msg string) {
	infoLogger.Output(2, msg)
}

func Error(msg string) {
	errorLogger.Output(2, msg)
}

func Fatal(err error) {
	Error(err.Error())
	os.Exit(1)
}

func Close() error {
	if logFile == nil {
		return nil
	}
	err := logFile.Close()
	logFile = nil
	return err
}
