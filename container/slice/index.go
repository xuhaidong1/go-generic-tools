package slice

// Index 返回第一个匹配的元素的下标
func Index[T comparable](src []T, target T) int {
	return IndexFunc[T](src, target, func(src, dst T) bool {
		return src == dst
	})
}

// IndexFunc 用户传入比较逻辑 func(src,dst T)bool
func IndexFunc[T any](src []T, target T, f func(src, dst T) bool) int {
	if src == nil {
		return -1
	}
	for i, v := range src {
		if f(target, v) {
			return i
		}
	}
	return -1
}

// IndexAll 返回所有匹配元素的下标
func IndexAll[T comparable](src []T, target T) []int {
	return IndexAllFunc[T](src, target, func(src, dst T) bool {
		return src == dst
	})
}

// IndexAllFunc 用户传入比较逻辑 func(src,dst T)bool
func IndexAllFunc[T any](src []T, target T, f func(src, dst T) bool) []int {
	res := make([]int, 0, len(src))
	if src == nil {
		return res
	}
	for i, v := range src {
		if f(target, v) {
			res = append(res, i)
		}
	}
	return res
}
