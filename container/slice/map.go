package slice

func toMap[T comparable](src []T) map[T]struct{} {
	res := make(map[T]struct{}, len(src))
	for _, v := range src {
		res[v] = struct{}{}
	}
	return res
}

// Deduplicate 去除重复元素 map不保真结果元素的相对顺序与源切片一致
func DeduplicateBymap[T comparable](src []T) []T {
	mp := toMap(src)
	res := make([]T, 0, len(mp))
	for k, _ := range mp {
		res = append(res, k)
	}
	return res
}

func Deduplicate[T comparable](src []T) []T {
	return DeduplicateFunc[T](src, func(src, dst T) bool {
		return src == dst
	})
}

func DeduplicateFunc[T any](src []T, f equalFunc[T]) []T {
	res := make([]T, 0, len(src))
	for _, v := range src {
		if !ContainsFunc(res, v, f) {
			res = append(res, v)
		}
	}
	return res
}
