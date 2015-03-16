package qerr

import "fmt"

type chainedErr struct {
	msg    string
	parent error
}

func Chain(msg string, parent error) error {
	if parent == nil {
		return nil
	}
	return &chainedErr{msg: msg, parent: parent}
}

func (this chainedErr) Error() string {
	return fmt.Sprintf("%s\n  caused by: %s", this.msg, this.parent)
}

func (this chainedErr) Parent() error {
	return this.parent
}
