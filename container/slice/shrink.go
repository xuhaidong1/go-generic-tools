package slice

func Shrink[T any](src []T) []T {
	n, ok := Calcap(len(src), cap(src))
	if ok {
		res := make([]T, 0, n)
		res = append(res, src...)
		return res
	}
	return src
}

func Calcap(length int, cap int) (int, bool) {
	if cap < 64 {
		return cap, false
	} else if cap < 256 {
		if length < (cap / 2) {
			return cap / 2, true
		} else {
			return cap, false
		}
	} else {
		var factor float32 = 0.75
		if length < (cap / 2) {
			return int(float32(cap) * factor), true
		} else {
			return cap, false
		}
	}
}
