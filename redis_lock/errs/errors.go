package errs

import "errors"

var (
	ErrLockNotHold         = errors.New("没有持有锁")
	ErrFailedToPreemptLock = errors.New("没有抢到锁")
)
