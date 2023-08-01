package slice

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContains(t *testing.T) {
	type args[T comparable] struct {
		src []T
		dst T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want bool
	}
	tests := []testCase[int]{
		{
			name: "contain",
			args: args[int]{src: srcbig, dst: 6},
			want: true,
		},
		{
			name: "not contain",
			args: args[int]{src: srcbig, dst: 100},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Contains(tt.args.src, tt.args.dst), "Contains(%v, %v)", tt.args.src, tt.args.dst)
		})
	}
}

func TestContainsFunc(t *testing.T) {
	type args[T any] struct {
		src []T
		dst T
		f   equalFunc[T]
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want bool
	}
	tests := []testCase[user]{
		{
			name: "contain",
			args: args[user]{
				src: usersbig,
				dst: user{3, 24, ""},
				f:   eqf,
			},
			want: true,
		},
		{
			name: "not contain",
			args: args[user]{
				src: usersbig,
				dst: user{43, 24, ""},
				f:   eqf,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ContainsFunc(tt.args.src, tt.args.dst, tt.args.f), "ContainsFunc(%v, %v, %v)", tt.args.src, tt.args.dst, tt.args.f)
		})
	}
}

func TestContainsAll(t *testing.T) {
	type args[T comparable] struct {
		src []T
		dst []T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want bool
	}
	tests := []testCase[int]{
		{
			name: "contain",
			args: args[int]{src: srcbig, dst: srcsmall},
			want: true,
		},
		{
			name: "not contain",
			args: args[int]{src: srcbig, dst: srcsmallno},
			want: false,
		},
		{
			name: "not contain 2",
			args: args[int]{src: srcsmall, dst: srcbig},
			want: false,
		},
		{
			name: "not contain 3 重复的dst元素",
			args: args[int]{src: srcbig, dst: []int{2, 2, 2, 2, 2, 4, 4, 6, 7}},
			want: false,
		},
		{
			name: "not contain 4",
			args: args[int]{src: []int{1, 2, 3, 4}, dst: srcbig},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ContainsAll(tt.args.src, tt.args.dst), "ContainsAll(%v, %v)", tt.args.src, tt.args.dst)
		})
	}
}

func TestContainsAllFunc(t *testing.T) {
	type args[T any] struct {
		src []T
		dst []T
		f   equalFunc[user]
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want bool
	}
	tests := []testCase[user]{
		{
			name: "contain",
			args: args[user]{
				src: usersbig,
				dst: usersmall,
				f:   eqf,
			},
			want: true,
		},
		{
			name: "not contain",
			args: args[user]{
				src: usersbig,
				dst: usersmallnoall,
				f:   eqf,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ContainsAllFunc(tt.args.src, tt.args.dst, tt.args.f), "ContainsAllFunc(%v, %v, %v)", tt.args.src, tt.args.dst, tt.args.f)
		})
	}
}

var srcbig []int = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
var srcsmall []int = []int{2, 4, 6, 7}
var srcsmallno []int = []int{2, 4, 6, 7, 20}
var usersbig []user = []user{
	{1, 2, ""}, {2, 4, ""}, {3, 24, ""}, {2, 1, ""}, {5, 6, ""}, {7, 4, ""},
}
var usersmall []user = []user{{3, 24, ""}, {2, 1, ""}}
var usersmallnoall []user = []user{{3, 24, ""}, {2, 1, ""}, {63, 24, ""}}

func eqf(src, dst user) bool {
	return src.Id == dst.Id
}
