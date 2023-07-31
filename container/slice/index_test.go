package slice

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIndex(t *testing.T) {
	type args[T comparable] struct {
		src    []T
		target T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want int
	}
	tests := []testCase[int]{
		{
			name: "find",
			args: args[int]{
				src:    []int{1, 2, 3, 4, 5, 6, 7, 8, 2, 4, 4},
				target: 4,
			},
			want: 3,
		},
		{
			name: "not find",
			args: args[int]{
				src:    []int{1, 2, 3, 4, 5, 6, 7, 8, 2, 4, 4},
				target: 40,
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Index(tt.args.src, tt.args.target))
		})
	}
}

func TestIndexFunc(t *testing.T) {
	type args[T any] struct {
		src    []T
		target T
		f      func(src, dst T) bool
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want int
	}
	tests := []testCase[user]{
		{
			name: "find the first user whose age=18",
			args: args[user]{
				src: []user{user{1, 6, "beijing"}, user{2, 13, "shanghai"},
					user{3, 18, "guangzhou"}, user{4, 30, "shenzhen"}, user{5, 23, "hangzhou"}},
				target: user{100, 18, ""},
				f: func(src user, target user) bool {
					return target.Age == src.Age
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IndexFunc(tt.args.src, tt.args.target, tt.args.f))
		})
	}
}

type user struct {
	Id      int
	Age     int
	Address string
}

func TestIndexAll(t *testing.T) {
	type args[T comparable] struct {
		src    []T
		target T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want []int
	}
	tests := []testCase[int]{
		{
			name: "find",
			args: args[int]{
				src:    []int{1, 2, 3, 4, 5, 6, 7, 8, 2, 4, 4},
				target: 4,
			},
			want: []int{3, 9, 10},
		},
		{
			name: "not find",
			args: args[int]{
				src:    []int{1, 2, 3, 4, 5, 6, 7, 8, 2, 4, 4},
				target: 40,
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IndexAll(tt.args.src, tt.args.target))
		})
	}
}
