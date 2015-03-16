package qerr

import (
	"fmt"
	"runtime"
)

type locatedErr struct {
	parent error
	file   string
	line   int
}

func NewLocated(parent error) error {
	if parent == nil {
		return nil
	}
	e := &locatedErr{parent: parent, file: "???", line: 0}
	_, file, line, success := runtime.Caller(1)
	if success {
		e.file = file
		e.line = line
	}
	return e
}

func (this locatedErr) Error() string {
	return fmt.Sprintf("%s (at %s:%d)", this.parent.Error(), this.file, this.line)
}

func (this locatedErr) Parent() error {
	return this.parent
}
