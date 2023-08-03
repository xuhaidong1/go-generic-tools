package slice

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInterSet(t *testing.T) {
	type args[T comparable] struct {
		src1 []T
		src2 []T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[int]{
		{
			name: "交集是真子集",
			args: args[int]{
				src1: []int{1, 2, 3, 4, 5},
				src2: []int{6, 7, 5, 4, 3},
			},
			want: []int{3, 4, 5},
		},
		{
			name: "交集是某个源集合",
			args: args[int]{
				src1: []int{1, 2, 3, 4, 5},
				src2: []int{2, 3, 5},
			},
			want: []int{2, 3, 5},
		},
		{
			name: "没有交集",
			args: args[int]{
				src1: []int{1, 2, 3, 4, 5},
				src2: []int{8, 9},
			},
			want: []int{},
		},
		{
			name: "没有交集",
			args: args[int]{
				src1: []int{},
				src2: []int{8, 9},
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InterSet(tt.args.src1, tt.args.src2)
			gotbool := ContainsAll(tt.want, got)
			fmt.Printf("want:%v,got:%v\n", tt.want, got)
			assert.Equal(t, true, gotbool)
		})
	}
}

func TestInterSetFunc(t *testing.T) {
	type args[T any] struct {
		src1 []T
		src2 []T
		f    equalFunc[T]
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[int]{
		{
			name: "交集是真子集",
			args: args[int]{
				src1: []int{1, 2, 3, 4, 5},
				src2: []int{6, 7, 5, 4, 3},
				f:    eqint,
			},
			want: []int{3, 4, 5},
		},
		{
			name: "交集是某个源集合",
			args: args[int]{
				src1: []int{1, 2, 3, 4, 5},
				src2: []int{2, 3, 5},
				f:    eqint,
			},
			want: []int{2, 3, 5},
		},
		{
			name: "没有交集",
			args: args[int]{
				src1: []int{1, 2, 3, 4, 5},
				src2: []int{8, 9},
				f:    eqint,
			},
			want: []int{},
		},
		{
			name: "没有交集",
			args: args[int]{
				src1: []int{},
				src2: []int{8, 9},
				f:    eqint,
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InterSetFunc(tt.args.src1, tt.args.src2, tt.args.f)
			gotbool := ContainsAll(tt.want, got)
			fmt.Printf("want:%v,got:%v\n", tt.want, got)
			assert.Equal(t, true, gotbool)
		})
	}
}

func eqint(src, dst int) bool {
	return src == dst
}
