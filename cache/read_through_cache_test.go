package cache

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestReadThroughCache_Get(t *testing.T) {
	type User struct {
		id   int
		name string
	}
	var db gorm.DB
	UserCache := NewReadThroughCache(NewLocalCache(nil), time.Hour, func(ctx context.Context, key string) (any, error) {
		if strings.HasPrefix(key, "/user/") {
			id := strings.Trim(key, "/user/")
			atoi, _ := strconv.Atoi(id)
			u := &User{id: atoi}
			return u.name, db.WithContext(ctx).First(&u).Error
		}
		return nil, errors.New("key不对")
	})
	fmt.Println(UserCache)
}
