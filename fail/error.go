package fail

import "fmt"

// GlobalPrintLevel controls what will be printed to the console during execution
var GlobalPrintLevel = 0

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
// on the globally set variable GlobalPrintLevel
func LeveledPrint(level int, fun, format string, a ...interface{}) {
	if level == 0 {
		fmt.Printf(format+"\n", a...)
	} else if level <= GlobalPrintLevel {
		fmt.Printf(fun+"() : "+format+"\n", a...)
	}
}
