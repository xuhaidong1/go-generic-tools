package slice

func Contains[T comparable](src []T, dst T) bool {
	return ContainsFunc[T](src, dst, func(src, dst T) bool {
		return src == dst
	})
}

func ContainsFunc[T any](src []T, dst T, f equalFunc[T]) bool {
	for _, v := range src {
		if f(v, dst) {
			return true
		}
	}
	return false
}

func ContainsAll[T comparable](src, dst []T) bool {
	return ContainsAllFunc[T](src, dst, func(src, dst T) bool {
		return src == dst
	})
}

func ContainsAllFunc[T any](src, dst []T, f equalFunc[T]) bool {
	//由于是any，不能用map，map的key不能比较
	//遍历src，在dst中找到就在dst中删除，如果dst空了，则true，若src遍历结束dst没空，则false
	//涉及到对切片修改，建临时变量拷贝dst
	tmp := make([]T, 0, len(src))
	tmp = append(tmp, src...)
	for _, v := range dst {
		if len(tmp) > 0 {
			if idx := IndexFunc[T](tmp, v, f); idx != -1 {
				tmp, _, _ = Delete[T](tmp, idx)
			} else {
				//src内没找到dst的某个元素，不全部包含
				return false
			}
		} else {
			//在循环内src没了，但dst没遍历完，不全部包含
			return false
		}
	}
	//dst遍历完成了，都找到了
	return true
}

//func ContainsAllFunc[T any](src, dst []T, f equalFunc[T]) bool {
//	//bug：没有区分重复的元素，not contain 3用例过不了
//	for _, v := range dst {
//		if !ContainsFunc[T](src, v, f) {
//			return false
//		}
//	}
//	return true
//}
