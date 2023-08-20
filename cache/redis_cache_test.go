package cache

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/xuhaidong1/go-generic-tools/cache/mocks"
	"testing"
	"time"
)

// gomock命令：mockgen -package=mocks -destination=mocks/redis_cmdable.go github.com/redis/go-redis/v9 Cmdable
func TestRedisCache_Set(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := []struct {
		name       string
		mock       func() redis.Cmdable
		key        string
		val        any
		expiration time.Duration
		wantErr    error
	}{
		{
			name: "ok",
			mock: func() redis.Cmdable {
				res := mocks.NewMockCmdable(ctrl)
				cmd := redis.NewStatusCmd(nil)
				cmd.SetVal("OK")
				res.EXPECT().Set(gomock.Any(), "key1", "val1", time.Minute).Return(cmd)
				return res
			},
			key:        "key1",
			val:        "val1",
			expiration: time.Minute,
		},
		{
			name: "timeout",
			mock: func() redis.Cmdable {
				res := mocks.NewMockCmdable(ctrl)
				cmd := redis.NewStatusCmd(nil)
				cmd.SetErr(context.DeadlineExceeded)
				//mock了redis那边的返回：我们规定了用上述参数调用redis.cmdable接口的set方法，返回是我们设定的结果（cmd）
				res.EXPECT().Set(gomock.Any(), "key1", "val1", time.Minute).Return(cmd)
				return res
			},
			key:        "key1",
			val:        "val1",
			expiration: time.Minute,
			wantErr:    context.DeadlineExceeded,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmdable := tc.mock()
			client := NewRedisCache(cmdable)
			err := client.Set(context.Background(), tc.key, tc.val, tc.expiration)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
