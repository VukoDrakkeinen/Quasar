package qerr

import "fmt"

type parseErr struct {
	chainedErr
	input string
}

func NewParse(msg string, parent error, input string) error {
	return &parseErr{chainedErr: chainedErr{msg: msg, parent: parent}, input: input}
}

func (this *parseErr) Input() string { //TODO: this can't be used, because parseErr is unexported
	return this.input
}

func (this parseErr) Error() string {
	return fmt.Sprintf("%s (input: %s)\n  caused by: %s", this.msg, this.input, this.parent)
}
