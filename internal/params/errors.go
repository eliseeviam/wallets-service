package params

import (
	"fmt"
)

type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: `%v`, message: `%s`", e.Code, e.Message)
}
