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
	return newLocated(parent, 2)
}

func NewEmbeddedLocated(parent error) error {
	return newLocated(parent, 3)
}

func (this locatedErr) Error() string {
	return fmt.Sprintf("%s (at %s:%d)", this.parent.Error(), this.file, this.line)
}

func (this locatedErr) Parent() error {
	return this.parent
}

func newLocated(parent error, ascendFrames int) error {
	if parent == nil {
		return nil
	}
	e := &locatedErr{parent: parent, file: "???", line: 0}
	_, file, line, success := runtime.Caller(ascendFrames)
	if success {
		e.file = file
		e.line = line
	}
	return e
}
