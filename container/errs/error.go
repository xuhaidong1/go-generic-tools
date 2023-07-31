package errs

import "errors"

func NewErrIndexOutOfRange() error {
	return errors.New("下标超出范围")
}

func NewErrInputNil() error {
	return errors.New("输入为nil")
}
