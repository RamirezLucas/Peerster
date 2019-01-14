package fail

import (
	"fmt"
	"sync"
)

// GlobalPrintLevel controls what will be printed to the console during execution
var GlobalPrintLevel = 1

// Mutex to ensure atomocity of prints
var printMux sync.Mutex

// CustomError - Represents a custom error
type CustomError struct {
	Fun  string // Function's name
	Desc string // Error description
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("Error in %s(): %s", e.Fun, e.Desc)
}

// CustomPanic panics while displaying an error message
func CustomPanic(fun, format string, a ...interface{}) {
	str := fmt.Sprintf(format, a...)
	panic(fun + "() -> " + str + "\n")
}

// LeveledPrint prints text to the console depending on its importance level and
// on the globally set varia²e GlobalPrintLevel
func LeveledPrint(level int, fun, format string, a ...interface{}) {
	// Grab the mutex
	printMux.Lock()
	defer printMux.Unlock()

	if level == 0 {
		// Mandatory prints
		fmt.Printf(format+"\n", a...)
	} else if level <= GlobalPrintLevel {
		fmt.Printf(fun+"() : "+format+"\n", a...)
	}
}
