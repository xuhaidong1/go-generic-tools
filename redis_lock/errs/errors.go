package errs

import "errors"

func NewErrLockNotHold() error {
	return errors.New("没有持有锁")
}

func NewErrFailedToPreemptLock() error {
	return errors.New("没有抢到锁")
}
