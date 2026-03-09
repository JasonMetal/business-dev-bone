package storage

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// 错误定义
var (
	ErrFailedToAcquireLock = errors.New("failed to acquire lock")
	ErrLockNotHeld         = errors.New("lock not held")
)

// RedisLock 表示一个Redis分布式锁
type RedisLock struct {
	client      redis.UniversalClient // 支持单点和集群
	key         string                // 锁的键
	value       string                // 锁的值（唯一标识）
	expiration  time.Duration         // 锁的过期时间
	retryCount  int                   // 重试次数
	retryDelay  time.Duration         // 重试间隔
	renewalTime time.Duration         // 自动续期时间间隔
	ctx         context.Context       // 上下文
	cancel      context.CancelFunc    // 取消续期的函数
	mu          sync.Mutex            // 并发控制
	held        bool                  // 是否持有锁
}

// NewRedisLock 创建一个新的Redis锁实例
func NewRedisLock(client redis.UniversalClient, key string, expiration time.Duration) *RedisLock {
	return &RedisLock{
		client:     client,
		key:        key,
		expiration: expiration,
		retryCount: 5,
		retryDelay: 100 * time.Millisecond,
		// 自动续期时间设为锁过期时间的1/3，确保在锁过期前续期
		renewalTime: expiration / 3,
	}
}

// SetRetry 设置重试策略
func (l *RedisLock) SetRetry(retryCount int, retryDelay time.Duration) *RedisLock {
	l.retryCount = retryCount
	l.retryDelay = retryDelay
	return l
}

// Acquire 获取锁
func (l *RedisLock) Acquire(ctx context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 生成唯一的锁值，防止误释放其他客户端的锁
	l.value = fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())
	l.ctx, l.cancel = context.WithCancel(ctx)

	for i := 0; i <= l.retryCount; i++ {
		// 使用SETNX命令尝试获取锁
		set, err := l.client.SetNX(l.ctx, l.key, l.value, l.expiration).Result()
		if err != nil {
			return false, err
		}

		if set {
			l.held = true
			// 启动自动续期（看门狗机制）
			go l.autoRenew()
			return true, nil
		}

		// 重试前等待
		if i < l.retryCount {
			select {
			case <-time.After(l.retryDelay):
			case <-ctx.Done():
				return false, ctx.Err()
			}
		}
	}

	return false, ErrFailedToAcquireLock
}

// Release 释放锁
func (l *RedisLock) Release(ctx context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.held {
		return false, ErrLockNotHeld
	}

	// 停止自动续期
	if l.cancel != nil {
		l.cancel()
	}

	// 使用Lua脚本确保原子性：先检查锁是否是自己的，再释放
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Int64()
	if err != nil {
		return false, err
	}

	if result == 1 {
		l.held = false
		return true, nil
	}

	return false, ErrLockNotHeld
}

// autoRenew 自动续期（看门狗机制）
func (l *RedisLock) autoRenew() {
	ticker := time.NewTicker(l.renewalTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 使用Lua脚本续期：先检查锁是否是自己的，再续期
			script := `
				if redis.call("GET", KEYS[1]) == ARGV[1] then
					return redis.call("PEXPIRE", KEYS[1], ARGV[2])
				else
					return 0
				end
			`

			// 将过期时间转换为毫秒
			expirationMs := int64(l.expiration / time.Millisecond)
			result, err := l.client.Eval(l.ctx, script, []string{l.key}, l.value, expirationMs).Int64()
			if err != nil || result == 0 {
				// 续期失败，可能锁已过期或被释放
				l.mu.Lock()
				l.held = false
				l.mu.Unlock()
				return
			}

		case <-l.ctx.Done():
			return
		}
	}
}

// IsHeld 检查是否持有锁
func (l *RedisLock) IsHeld() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.held
}
