package slice

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeduplicate(t *testing.T) {
	type args[T comparable] struct {
		src []T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[int]{
		{
			name: "去除重复元素1",
			args: args[int]{src: []int{2, 2, 2, 2, 2, 4, 4, 6, 7}},
			want: []int{2, 4, 6, 7},
		},
		{
			name: "去除重复元素2",
			args: args[int]{src: []int{2, 4, 6, 7}},
			want: []int{2, 4, 6, 7},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Deduplicate(tt.args.src), "Deduplicate(%v)", tt.args.src)
		})
	}
}
