package slice

import (
	"github.com/stretchr/testify/assert"
	"github.com/xuhaidong1/go-generic-tools/container/errs"
	"testing"
)

func TestDelete(t *testing.T) {
	type args[T any] struct {
		s   []T
		idx int
	}
	type testCase[T any] struct {
		name     string
		args     args[T]
		want     []T
		wantbool bool
		wantcap  int
		wantErr  error
	}
	tests := []testCase[int]{
		{
			name:     "Input nil",
			args:     args[int]{nil, 0},
			want:     nil,
			wantbool: false,
			wantcap:  0,
			wantErr:  errs.NewErrInputNil(),
		},
		{
			name:     "IndexOutofRange",
			args:     args[int]{[]int{1, 2, 3, 4, 5}, 5},
			want:     nil,
			wantbool: false,
			wantcap:  0,
			wantErr:  errs.NewErrIndexOutOfRange(),
		},
		{
			name:     "Remove first",
			args:     args[int]{[]int{1, 2, 3, 4, 5}, 0},
			want:     []int{2, 3, 4, 5},
			wantbool: true,
			wantcap:  5,
			wantErr:  nil,
		},
		{
			name:     "Remove mid",
			args:     args[int]{[]int{1, 2, 3, 4, 5}, 2},
			want:     []int{1, 2, 4, 5},
			wantbool: true,
			wantcap:  5,
			wantErr:  nil,
		},
		{
			name:     "Remove last",
			args:     args[int]{[]int{1, 2, 3, 4, 5}, 4},
			want:     []int{1, 2, 3, 4},
			wantbool: true,
			wantcap:  5,
			wantErr:  nil,
		},
		{
			name: "Remove mid and Shrink(cap<64)",
			args: args[int]{
				s: func() []int {
					res := make([]int, 0, 60)
					return append(res, []int{1, 2, 3, 4, 5}...)
				}(),
				idx: 4,
			},
			want:     []int{1, 2, 3, 4},
			wantbool: true,
			wantcap:  60,
			wantErr:  nil,
		},
		{
			name: "Remove mid and Shrink(cap>256)",
			args: args[int]{
				s: func() []int {
					res := make([]int, 0, 400)
					return append(res, []int{1, 2, 3, 4, 5}...)
				}(),
				idx: 4,
			},
			want:     []int{1, 2, 3, 4},
			wantbool: true,
			wantcap:  300,
			wantErr:  nil,
		},
		{
			name: "Remove mid and Shrink(cap>64&&cap<256)",
			args: args[int]{
				s: func() []int {
					res := make([]int, 0, 200)
					return append(res, []int{1, 2, 3, 4, 5}...)
				}(),
				idx: 4,
			},
			want:     []int{1, 2, 3, 4},
			wantbool: true,
			wantcap:  100,
			wantErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotbool, err := Delete(tt.args.s, tt.args.idx)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantcap, cap(got))
			assert.Equal(t, tt.wantbool, gotbool)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
