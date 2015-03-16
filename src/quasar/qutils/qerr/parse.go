package qerr

type parseErr struct {
	chainedErr
	input string
}

func NewParse(msg string, parent error, input string) error {
	return &parseErr{chainedErr: chainedErr{msg: msg, parent: parent}, input: input}
}

func (this *parseErr) Input() string {
	return this.input
}
