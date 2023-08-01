package slice

import (
	"github.com/stretchr/testify/assert"
	g "github.com/xuhaidong1/go-generic-tools"
	"testing"
)

func TestMax(t *testing.T) {
	type args[T g.RealNumber] struct {
		src []T
	}
	type testCase[T g.RealNumber] struct {
		name    string
		args    args[T]
		wantRes T
	}
	tests := []testCase[float32]{
		{
			name:    "case 1",
			args:    args[float32]{[]float32{1.23, 13.2, 45.32}},
			wantRes: 45.32,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantRes, Max(tt.args.src), "Max(%v)", tt.args.src)
		})
	}
}

func TestMin(t *testing.T) {
	type args[T g.RealNumber] struct {
		src []T
	}
	type testCase[T g.RealNumber] struct {
		name    string
		args    args[T]
		wantRes T
	}
	tests := []testCase[float32]{
		{
			name:    "case 1",
			args:    args[float32]{[]float32{1.23, 13.2, 45.32}},
			wantRes: 1.23,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantRes, Min(tt.args.src), "Min(%v)", tt.args.src)
		})
	}
}

func TestSum(t *testing.T) {
	type args[T g.Number] struct {
		src []T
	}
	type testCase[T g.Number] struct {
		name    string
		args    args[T]
		wantRes T
	}
	tests := []testCase[float32]{
		{
			name:    "case 1",
			args:    args[float32]{[]float32{1.23, 13.2, 45.32}},
			wantRes: 59.75,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantRes, Sum(tt.args.src), "Sum(%v)", tt.args.src)
		})
	}
}
