package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type _logger struct {
	isEnabled bool
	isSetup   bool
	sync.RWMutex
}

var logger = _logger{
	isEnabled: true,
	isSetup:   false,
}

func isEnabledAndSetup() bool {
	logger.RLock()
	if logger.isSetup {
		logger.RUnlock()
		return logger.isEnabled
	}
	logger.RUnlock()

	logger.Lock()
	defer logger.Unlock()

	if logger.isSetup {
		return logger.isEnabled
	}

	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(os.Stdout)
	logger.isSetup = true
	return logger.isEnabled
}

func Set(enabled bool) {
	logger.Lock()
	defer logger.Unlock()
	logger.isEnabled = enabled
}

func Printlnf(s string, v ...interface{}) {
	if isEnabledAndSetup() {
		text := fmt.Sprintf(s, v...)
		log.Printf("%s\n", text)
	}
}
