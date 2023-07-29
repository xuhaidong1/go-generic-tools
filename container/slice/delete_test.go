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
		wantErr  error
	}
	tests := []testCase[int]{
		{
			name:     "IndexOutofRange",
			args:     args[int]{[]int{1, 2, 3, 4, 5}, 5},
			want:     nil,
			wantbool: false,
			wantErr:  errs.NewErrIndexOutOfRange(),
		},
		{
			name:     "Remove first",
			args:     args[int]{[]int{1, 2, 3, 4, 5}, 0},
			want:     []int{2, 3, 4, 5},
			wantbool: true,
			wantErr:  nil,
		},
		{
			name:     "Remove mid",
			args:     args[int]{[]int{1, 2, 3, 4, 5}, 2},
			want:     []int{1, 2, 4, 5},
			wantbool: true,
			wantErr:  nil,
		},
		{
			name:     "Remove first",
			args:     args[int]{[]int{1, 2, 3, 4, 5}, 4},
			want:     []int{1, 2, 3, 4},
			wantbool: true,
			wantErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotbool, err := Delete(tt.args.s, tt.args.idx)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantbool, gotbool)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
