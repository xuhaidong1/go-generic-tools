package slice

import g "github.com/xuhaidong1/go-generic-tools"

func Sum[T g.Number](src []T) (res T) {
	for _, v := range src {
		res += v
	}
	return res
}

func Max[T g.RealNumber](src []T) (res T) {
	res = src[0]
	for _, v := range src {
		if res < v {
			res = v
		}
	}
	return res
}

func Min[T g.RealNumber](src []T) (res T) {
	res = src[0]
	for _, v := range src {
		if res > v {
			res = v
		}
	}
	return res
}
