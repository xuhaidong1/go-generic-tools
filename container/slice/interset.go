package slice

func InterSet[T comparable](src1, src2 []T) []T {
	if src1 == nil || src2 == nil {
		return []T{}
	}
	n := min(len(src1), len(src2))
	res := make([]T, 0, n)
	s1 := toMap[T](src1)
	for _, v := range src2 {
		if _, ok := s1[v]; ok {
			res = append(res, v)
		}
	}
	return DeduplicateBymap[T](res)
}

func InterSetFunc[T any](src1 []T, src2 []T, f equalFunc[T]) []T {
	if src1 == nil || src2 == nil {
		return []T{}
	}
	n := min(len(src1), len(src2))
	res := make([]T, 0, n)
	tmp := make([]T, 0, len(src2))
	tmp = append(tmp, src2...)
	for _, v := range src1 {
		if len(tmp) > 0 {
			if idx := IndexFunc[T](tmp, v, f); idx != -1 {
				tmp, _, _ = Delete[T](tmp, idx)
				res = append(res, v)
			}
		}
	}
	//dst遍历完成了，都找到了
	return DeduplicateFunc[T](res, f)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
