package storage

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"business-dev-bone/pkg/component-base/util/idutil"
	"github.com/go-redsync/redsync/v4"
)

const testKeyPrefix = "test:redis_cluster:"

// TestExpiryWithJitter tests expiryWithJitter (no Redis required).
func TestExpiryWithJitter(t *testing.T) {
	t.Run("zero returns zero", func(t *testing.T) {
		got := expiryWithJitter(0)
		if got != 0 {
			t.Errorf("expiryWithJitter(0) = %v, want 0", got)
		}
	})

	t.Run("negative returns unchanged", func(t *testing.T) {
		d := -1 * time.Second
		got := expiryWithJitter(d)
		if got != d {
			t.Errorf("expiryWithJitter(%v) = %v, want %v", d, got, d)
		}
	})

	t.Run("result in [expiry, expiry*(1+expiryJitterRatio)]", func(t *testing.T) {
		expiry := 5 * time.Second
		minExpiry := expiry
		maxExpiry := time.Duration(float64(expiry) * (1 + expiryJitterRatio))

		seen := make(map[time.Duration]bool)
		for i := 0; i < 100; i++ {
			got := expiryWithJitter(expiry)
			if got < minExpiry || got > maxExpiry {
				t.Errorf("expiryWithJitter(%v) = %v, want in [%v, %v]", expiry, got, minExpiry, maxExpiry)
			}
			seen[got] = true
		}
		if len(seen) < 2 {
			t.Errorf("expected jitter variation, got %d distinct values in 100 runs", len(seen))
		}
	})
}

// TestConfigValidate tests Config.Validate (no Redis required).
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{"valid host+port", &Config{Host: "127.0.0.1", Port: 6379}, false},
		{"valid addrs", &Config{Addrs: []string{"127.0.0.1:6379"}}, false},
		{"empty config", &Config{}, true},
		{"invalid port negative", &Config{Host: "x", Port: -1}, true},
		{"invalid port overflow", &Config{Host: "x", Port: 70000}, true},
		{"invalid database", &Config{Host: "x", Port: 6379, Database: -1}, true},
		{"invalid database >15", &Config{Host: "x", Port: 6379, Database: 16}, true},
		{"invalid timeout", &Config{Host: "x", Port: 6379, Timeout: -1}, true},
		{"invalid maxidle", &Config{Host: "x", Port: 6379, MaxIdle: -1}, true},
		{"invalid maxactive", &Config{Host: "x", Port: 6379, MaxActive: -1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// setupRedisForTest connects to Redis for integration tests.
// Skips if REDIS_SKIP=1 or Redis is unreachable.
func setupRedisForTest(t *testing.T) {
	if os.Getenv("REDIS_SKIP") == "1" {
		t.Skip("REDIS_SKIP=1, skipping Redis integration tests")
	}
	// Already connected from previous test
	if Connected() {
		return
	}
	config := &Config{
		Host:     "127.0.0.1",
		Port:     6379,
		Database: 0,
	}
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		config.Addrs = []string{addr}
		config.Host = ""
		config.Port = 0
	}
	if pwd := os.Getenv("REDIS_PASSWORD"); pwd != "" {
		config.Password = pwd
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go ConnectToRedis(ctx, config)
	for i := 0; i < 5; i++ {
		if Connected() {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Skip("Redis not available, skipping integration tests (start Redis or set REDIS_SKIP=1)")
}

func TestRedisCluster_GetSetDelete(t *testing.T) {
	setupRedisForTest(t)
	key := testKeyPrefix + idutil.NewNoCylinderLineID()
	val := "hello-world"
	ttl := 10 * time.Second

	if err := Redis.SetKey(key, val, ttl); err != nil {
		t.Fatalf("SetKey: %v", err)
	}
	got, err := Redis.GetKey(key)
	if err != nil {
		t.Fatalf("GetKey: %v", err)
	}
	if got != val {
		t.Errorf("GetKey = %q, want %q", got, val)
	}
	if !Redis.DeleteKey(key) {
		t.Error("DeleteKey failed")
	}
	_, err = Redis.GetKey(key)
	if err != ErrKeyNotFound {
		t.Errorf("GetKey after delete = %v, want ErrKeyNotFound", err)
	}
}

func TestRedisCluster_GetRawKey_SetRawKey(t *testing.T) {
	setupRedisForTest(t)
	key := testKeyPrefix + "raw:" + idutil.NewNoCylinderLineID()
	val := "raw-value"

	if err := Redis.SetRawKey(key, val, 10*time.Second); err != nil {
		t.Fatalf("SetRawKey: %v", err)
	}
	got, err := Redis.GetRawKey(key)
	if err != nil {
		t.Fatalf("GetRawKey: %v", err)
	}
	if got != val {
		t.Errorf("GetRawKey = %q, want %q", got, val)
	}
	Redis.DeleteRawKey(key)
}

func TestRedisCluster_Exists(t *testing.T) {
	setupRedisForTest(t)
	key := testKeyPrefix + "exists:" + idutil.NewNoCylinderLineID()

	ok, err := Redis.Exists(key)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if ok {
		t.Error("Exists on missing key should be false")
	}
	_ = Redis.SetKey(key, "x", 10*time.Second)
	ok, err = Redis.Exists(key)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !ok {
		t.Error("Exists on existing key should be true")
	}
	Redis.DeleteKey(key)
}

func TestRedisCluster_Hash(t *testing.T) {
	setupRedisForTest(t)
	key := testKeyPrefix + "hash:" + idutil.NewNoCylinderLineID()

	if err := Redis.HSet(key, "a", "1", "b", "2"); err != nil {
		t.Fatalf("HSet: %v", err)
	}
	got, err := Redis.HGet(key, "a")
	if err != nil {
		t.Fatalf("HGet: %v", err)
	}
	if got != "1" {
		t.Errorf("HGet = %q, want 1", got)
	}
	m, err := Redis.HGetAll(key)
	if err != nil {
		t.Fatalf("HGetAll: %v", err)
	}
	if m["a"] != "1" || m["b"] != "2" {
		t.Errorf("HGetAll = %v", m)
	}
	Redis.HDel(key, "a", "b")
}

func TestRedisCluster_List(t *testing.T) {
	setupRedisForTest(t)
	key := testKeyPrefix + "list:" + idutil.NewNoCylinderLineID()

	_ = Redis.LPush(key, "a", "b", "c")
	n, err := Redis.LLen(key)
	if err != nil {
		t.Fatalf("LLen: %v", err)
	}
	if n != 3 {
		t.Errorf("LLen = %d, want 3", n)
	}
	got, err := Redis.RPop(key)
	if err != nil {
		t.Fatalf("RPop: %v", err)
	}
	if got != "a" {
		t.Errorf("RPop = %q, want a (LIFO)", got)
	}
	Redis.DeleteKey(key)
}

func TestRedisCluster_TryLock(t *testing.T) {
	setupRedisForTest(t)
	lockKey := testKeyPrefix + "lock:" + idutil.NewNoCylinderLineID()
	threadID := idutil.NewNoCylinderLineID()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	code, err := Redis.TryLock(ctx, lockKey, threadID)
	if err != nil {
		t.Fatalf("TryLock: %v", err)
	}
	if code != LockSuccess {
		t.Errorf("TryLock = %d, want LockSuccess", code)
	}
	if err := Redis.Unlock(ctx, lockKey, threadID); err != nil {
		t.Errorf("Unlock: %v", err)
	}
}

func TestRedisCluster_MutexLock(t *testing.T) {
	setupRedisForTest(t)
	lockKey := testKeyPrefix + "mutex:" + idutil.NewNoCylinderLineID()
	threadID := idutil.NewNoCylinderLineID()
	ctx := context.Background()

	mutex := Redis.MutexLock(ctx, lockKey, threadID, 5*time.Second)
	if err := mutex.Lock(); err != nil {
		t.Fatalf("MutexLock.Lock: %v", err)
	}
	ok, err := mutex.Unlock()
	if !ok || err != nil {
		t.Errorf("MutexLock.Unlock: ok=%v err=%v", ok, err)
	}
}

// getMutexExpiry uses reflection to read redsync.Mutex.expiry (unexported).
func getMutexExpiry(m *redsync.Mutex) time.Duration {
	v := reflect.ValueOf(m).Elem()
	f := v.FieldByName("expiry")
	return time.Duration(f.Int())
}

func TestRedisCluster_MutexLock_ExpiryHasJitter(t *testing.T) {
	setupRedisForTest(t)
	ctx := context.Background()
	expiry := 5 * time.Second
	minExpiry := expiry
	maxExpiry := time.Duration(float64(expiry) * (1 + expiryJitterRatio))

	seen := make(map[time.Duration]bool)
	for i := 0; i < 50; i++ {
		lockKey := testKeyPrefix + "mutex_jitter:" + idutil.NewNoCylinderLineID()
		threadID := idutil.NewNoCylinderLineID()
		mutex := Redis.MutexLock(ctx, lockKey, threadID, expiry)
		got := getMutexExpiry(mutex)
		if got < minExpiry || got > maxExpiry {
			t.Errorf("MutexLock expiry = %v, want in [%v, %v]", got, minExpiry, maxExpiry)
		}
		seen[got] = true
	}
	if len(seen) < 2 {
		t.Errorf("expected jitter variation, got %d distinct expiry values in 50 runs", len(seen))
	}
}
