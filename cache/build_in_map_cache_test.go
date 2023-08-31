package cache

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBuildinMapCache_Get(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name    string
		key     string
		wantVal any
		wantErr error
	}{
		{
			name:    "exist",
			key:     "key1",
			wantVal: "value1",
		},
		{
			name:    "not exist",
			key:     "invalid",
			wantErr: ErrCacheKeyNotExist,
		},
	}

	c := NewBuildinMapCache()
	err := c.Set(context.Background(), "key1", "value1", 2*time.Second)
	require.NoError(t, err)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := c.Get(context.Background(), tc.key)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVal, val)
		})
	}
	time.Sleep(time.Second * 3)
	_, err = c.Get(context.Background(), "key1")
	assert.Equal(t, ErrCacheKeyNotExist, err)
}

func TestBuildinMapCache_checkCycle(t *testing.T) {
	c := NewBuildinMapCache(WithCycleInterval(time.Second))
	err := c.Set(context.Background(), "key1", "value1", time.Millisecond*100)
	require.NoError(t, err)
	// 以防万一
	time.Sleep(time.Second * 3)
	_, err = c.Get(context.Background(), "key1")
	assert.Equal(t, ErrCacheKeyNotExist, err)
}
