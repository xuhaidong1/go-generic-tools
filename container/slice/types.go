package slice

// 用于比较两个元素是否相等，用户传入方法实现
type equalFunc[T any] func(src, dst T) bool
