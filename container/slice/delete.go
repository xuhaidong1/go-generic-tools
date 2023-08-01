package slice

import (
	"github.com/xuhaidong1/go-generic-tools/container/errs"
)

// Delete 删除切片指定位置的元素
// 如果下标不是合法的下标，返回 ErrIndexOutOfRange
// 返回true删除成功 false删除失败
func Delete[T any](s []T, idx int) ([]T, bool, error) {
	if s == nil {
		return nil, false, errs.NewErrInputNil()
	}
	length := len(s)
	if idx >= length || idx < 0 {
		return nil, false, errs.NewErrIndexOutOfRange()
	}
	j := 0
	for i, v := range s {
		if i != idx {
			s[j] = v
			j++
		}
	}
	return Shrink(s[:length-1]), true, nil
}

//func ExampleDelete() {
//	res, _ := slice.Delete[int]([]int{1, 2, 3, 4}, 2)
//	fmt.Println(res)
//	_, errs := slice.Delete[int]([]int{1, 2, 3, 4}, -1)
//	fmt.Println(errs)
//	// Output:
//	// [1 2 4]
//	// ekit: 下标超出范围，长度 4, 下标 -1
//}
