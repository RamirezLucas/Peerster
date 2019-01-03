package fail

import (
	"Peerster/logger"
	"fmt"
)

func HandleError(e error) {
	if e != nil {
		logger.Printlnf("ERROR: %s", e)
	}
}
func HandleAbort(msg string, e error) {
	errorString := ""
	if e != nil {
		errorString = fmt.Sprintf(":\n\t->ERROR: %s", e)
	}
	logger.Printlnf("ABORT: %s%s", msg, errorString)
}
