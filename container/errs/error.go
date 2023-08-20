package errs

import "errors"

func NewErrIndexOutOfRange() error {
	return errors.New("下标超出范围")
}

func NewErrInputNil() error {
	return errors.New("输入为nil")
}

func NewErrFullQueue() error {
	return errors.New("队列满")
}

func NewErrEmptyQueue() error {
	return errors.New("队列空")
}
