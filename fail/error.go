package fail

import "fmt"

// CustomError - Represents a custom error
type CustomError struct {
	Fun  string // Function's name
	Desc string // Error description
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("Error in %s(): %s", e.Fun, e.Desc)
}
