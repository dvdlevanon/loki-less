package utils

import (
	"os"

	"github.com/go-errors/errors"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("utils")

func ConfigureLogger() error {
	logFormat := `[%{time:2006-01-02 15:04:05.000}] %{color}%{level:-7s}%{color:reset} %{message} [%{module} - %{shortfile}]`
	formatter, err := logging.NewStringFormatter(logFormat)
	if err != nil {
		return err
	}

	logWriter, err := os.OpenFile("loki-less.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	logging.SetBackend(logging.NewLogBackend(logWriter, "", 0))
	logging.SetFormatter(formatter)

	logger.Debugf("Logger initialized with format %v", logFormat)
	return nil
}

func LogError(err error) {
	if err == nil {
		return
	}

	var e *errors.Error
	if errors.As(err, &e) {
		logger.Errorf("Error: %v", e.ErrorStack())
	} else {
		logger.Errorf("Error: %v", err)
	}
}
