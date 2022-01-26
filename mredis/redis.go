package mredis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/moremorefun/mtool/mlog"

	"github.com/go-redis/redis/v8"
)

// baseKey 基础key
var baseKey = ""

// Create 创建数据库
func Create(address string, password string, dbIndex int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, // no password set
		DB:       dbIndex,  // use default DB
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		mlog.Log.Fatalf("redis ping error: %s", err.Error())
		return nil
	}
	return client
}

// SetBaseKey 设置基础key
func SetBaseKey(v string) {
	baseKey = v
}

// Get 获取
func Get(ctx context.Context, client *redis.Client, key string) (string, error) {
	key = fmt.Sprintf("%s_%s", baseKey, key)
	ret, err := client.WithContext(ctx).Get(context.Background(), key).Result()
	if err != nil {
		// "redis: nil" 不存在
		if !strings.Contains(err.Error(), "redis: nil") {
			return "", err
		}
		return "", nil
	}
	return ret, nil
}

// Set 设置
func Set(ctx context.Context, client *redis.Client, key, value string, du time.Duration) error {
	key = fmt.Sprintf("%s_%s", baseKey, key)
	err := client.WithContext(ctx).Set(context.Background(), key, value, du).Err()
	if err != nil {
		return err
	}
	return nil
}

// Rm 删除
func Rm(ctx context.Context, client *redis.Client, key string) error {
	key = fmt.Sprintf("%s_%s", baseKey, key)
	err := client.WithContext(ctx).Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}
