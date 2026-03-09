package storage

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"

	"business-dev-bone/pkg/component-base/errors"
	"business-dev-bone/pkg/component-base/log"
	"business-dev-bone/pkg/component-base/util/idutil"
)

// Config defines options for redis cluster.
type Config struct {
	Host                  string
	Port                  int
	Addrs                 []string
	MasterName            string
	Username              string
	Password              string
	Database              int
	MaxIdle               int
	MaxActive             int
	Timeout               int
	EnableCluster         bool
	UseSSL                bool
	SSLInsecureSkipVerify bool
}

// Validate validates the redis config and returns an error if invalid.
func (c *Config) Validate() error {
	if len(c.Addrs) == 0 && c.Host == "" && c.Port == 0 {
		return errors.New("redis config: at least one of Addrs, Host, or Port must be provided")
	}
	if c.Port < 0 || c.Port > 65535 {
		return errors.New("redis config: port must be between 0 and 65535")
	}
	if c.Database < 0 || c.Database > 15 {
		return errors.New("redis config: database must be between 0 and 15")
	}
	if c.Timeout < 0 {
		return errors.New("redis config: timeout cannot be negative")
	}
	if c.MaxIdle < 0 {
		return errors.New("redis config: max idle connections cannot be negative")
	}
	if c.MaxActive < 0 {
		return errors.New("redis config: max active connections cannot be negative")
	}

	return nil
}

// ErrRedisIsDown is returned when we can't communicate with redis.
var ErrRedisIsDown = errors.New("storage: Redis is either down or ws not configured")

var (
	singlePool      atomic.Value
	singleCachePool atomic.Value
	redisUp         atomic.Value
)

var disableRedis atomic.Value

// DisableRedis very handy when testsing it allows to dynamically enable/disable talking with redisW.
func DisableRedis(ok bool) {
	if ok {
		redisUp.Store(false)
		disableRedis.Store(true)

		return
	}
	redisUp.Store(true)
	disableRedis.Store(false)
}

func shouldConnect() bool {
	if v := disableRedis.Load(); v != nil {
		return !v.(bool)
	}

	return true
}

// Connected returns true if we are connected to redis.
func Connected() bool {
	if v := redisUp.Load(); v != nil {
		return v.(bool)
	}

	return false
}

func singleton(cache bool) redis.UniversalClient {
	if cache {
		v := singleCachePool.Load()
		if v != nil {
			return v.(redis.UniversalClient)
		}

		return nil
	}
	if v := singlePool.Load(); v != nil {
		return v.(redis.UniversalClient)
	}

	return nil
}

// nolint: unparam
func connectSingleton(cache bool, config *Config) bool {
	if singleton(cache) == nil {
		log.Debug("Connecting to redis cluster")
		if cache {
			singleCachePool.Store(NewRedisClusterPool(cache, config))

			return true
		}
		singlePool.Store(NewRedisClusterPool(cache, config))

		return true
	}

	return true
}

// RedisCluster is a storage manager that uses the redis database.
type RedisCluster struct {
	KeyPrefix string
	HashKeys  bool
	IsCache   bool
	Client    *redis.UniversalClient
	ctx       context.Context
}

var Redis *RedisCluster = &RedisCluster{ctx: context.Background()}

func clusterConnectionIsOpen(cluster RedisCluster) bool {
	c := singleton(cluster.IsCache)
	testKey := "redis-test-" + idutil.NewNoCylinderLineID()
	if err := c.Set(c.Context(), testKey, "test", time.Second).Err(); err != nil {
		log.Warnf("Error trying to set test key: %s", err.Error())

		return false
	}
	if _, err := c.Get(c.Context(), testKey).Result(); err != nil {
		log.Warnf("Error trying to get test key: %s", err.Error())

		return false
	}

	return true
}

// ConnectToRedis starts a go routine that periodically tries to connect to redis.
func ConnectToRedis(ctx context.Context, config *Config) {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	c := []RedisCluster{
		{}, {IsCache: true},
	}
	var ok bool
	for _, v := range c {
		if !connectSingleton(v.IsCache, config) {
			break
		}

		if !clusterConnectionIsOpen(v) {
			redisUp.Store(false)

			break
		}
		ok = true
	}
	redisUp.Store(ok)
again:
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if !shouldConnect() {
				continue
			}
			for _, v := range c {
				if !connectSingleton(v.IsCache, config) {
					redisUp.Store(false)

					goto again
				}

				if !clusterConnectionIsOpen(v) {
					redisUp.Store(false)

					goto again
				}
			}
			redisUp.Store(true)
		}
	}
}

// NewRedisClusterPool create a redis cluster pool.
func NewRedisClusterPool(isCache bool, config *Config) redis.UniversalClient {
	// Validate config first
	if err := config.Validate(); err != nil {
		log.Errorf("Invalid redis config: %s", err.Error())
		return nil
	}

	// redisSingletonMu is locked and we know the singleton is nil
	log.Debug("Creating new Redis connection pool")

	// poolSize applies per cluster node and not for the whole cluster.
	poolSize := 500
	if config.MaxActive > 0 {
		poolSize = config.MaxActive
	}

	timeout := 5 * time.Second

	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	var tlsConfig *tls.Config

	if config.UseSSL {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: config.SSLInsecureSkipVerify,
		}
	}

	var client redis.UniversalClient
	opts := &RedisOpts{
		Addrs:        getRedisAddrs(config),
		MasterName:   config.MasterName,
		Password:     config.Password,
		DB:           config.Database,
		DialTimeout:  timeout,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  240 * timeout,
		PoolSize:     poolSize,
		TLSConfig:    tlsConfig,
	}

	if opts.MasterName != "" {
		log.Info("--> [REDIS] Creating sentinel-backed failover client")
		client = redis.NewFailoverClient(opts.failover())
	} else if config.EnableCluster {
		log.Info("--> [REDIS] Creating cluster client")
		client = redis.NewClusterClient(opts.cluster())
	} else {
		log.Info("--> [REDIS] Creating single-node client")
		client = redis.NewClient(opts.simple())
	}
	Redis.Client = &client
	return client
}

func getRedisAddrs(config *Config) (addrs []string) {
	if len(config.Addrs) != 0 {
		// Return a copy to prevent external modifications
		addrs = make([]string, len(config.Addrs))
		copy(addrs, config.Addrs)

		return addrs
	}

	if len(addrs) == 0 && config.Port != 0 {
		host := config.Host
		if host == "" {
			host = "127.0.0.1"
		}
		addr := host + ":" + strconv.Itoa(config.Port)
		addrs = append(addrs, addr)
	}

	return addrs
}

// RedisOpts is the overridden type of redis.UniversalOptions. simple() and cluster() functions are not public in redis
// library.
// Therefore, they are redefined in here to use in creation of new redis cluster logic.
// We don't want to use redis.NewUniversalClient() logic.
type RedisOpts redis.UniversalOptions

func (o *RedisOpts) cluster() *redis.ClusterOptions {
	if len(o.Addrs) == 0 {
		o.Addrs = []string{"127.0.0.1:6379"}
	}

	return &redis.ClusterOptions{
		Addrs:     o.Addrs,
		OnConnect: o.OnConnect,

		Password: o.Password,

		MaxRedirects:   o.MaxRedirects,
		ReadOnly:       o.ReadOnly,
		RouteByLatency: o.RouteByLatency,
		RouteRandomly:  o.RouteRandomly,

		MaxRetries:      o.MaxRetries,
		MinRetryBackoff: o.MinRetryBackoff,
		MaxRetryBackoff: o.MaxRetryBackoff,

		DialTimeout:        o.DialTimeout,
		ReadTimeout:        o.ReadTimeout,
		WriteTimeout:       o.WriteTimeout,
		PoolSize:           o.PoolSize,
		MinIdleConns:       o.MinIdleConns,
		MaxConnAge:         o.MaxConnAge,
		PoolTimeout:        o.PoolTimeout,
		IdleTimeout:        o.IdleTimeout,
		IdleCheckFrequency: o.IdleCheckFrequency,

		TLSConfig: o.TLSConfig,
	}
}

func (o *RedisOpts) simple() *redis.Options {
	addr := "127.0.0.1:6379"
	if len(o.Addrs) > 0 {
		addr = o.Addrs[0]
	}

	return &redis.Options{
		Addr:      addr,
		OnConnect: o.OnConnect,

		DB:       o.DB,
		Password: o.Password,

		MaxRetries:      o.MaxRetries,
		MinRetryBackoff: o.MinRetryBackoff,
		MaxRetryBackoff: o.MaxRetryBackoff,

		DialTimeout:  o.DialTimeout,
		ReadTimeout:  o.ReadTimeout,
		WriteTimeout: o.WriteTimeout,

		PoolSize:           o.PoolSize,
		MinIdleConns:       o.MinIdleConns,
		MaxConnAge:         o.MaxConnAge,
		PoolTimeout:        o.PoolTimeout,
		IdleTimeout:        o.IdleTimeout,
		IdleCheckFrequency: o.IdleCheckFrequency,

		TLSConfig: o.TLSConfig,
	}
}

func (o *RedisOpts) failover() *redis.FailoverOptions {
	if len(o.Addrs) == 0 {
		o.Addrs = []string{"127.0.0.1:26379"}
	}

	return &redis.FailoverOptions{
		SentinelAddrs: o.Addrs,
		MasterName:    o.MasterName,
		OnConnect:     o.OnConnect,

		DB:       o.DB,
		Password: o.Password,

		MaxRetries:      o.MaxRetries,
		MinRetryBackoff: o.MinRetryBackoff,
		MaxRetryBackoff: o.MaxRetryBackoff,

		DialTimeout:  o.DialTimeout,
		ReadTimeout:  o.ReadTimeout,
		WriteTimeout: o.WriteTimeout,

		PoolSize:           o.PoolSize,
		MinIdleConns:       o.MinIdleConns,
		MaxConnAge:         o.MaxConnAge,
		PoolTimeout:        o.PoolTimeout,
		IdleTimeout:        o.IdleTimeout,
		IdleCheckFrequency: o.IdleCheckFrequency,

		TLSConfig: o.TLSConfig,
	}
}

// Connect will establish a connection this is always true because we are dynamically using redis.
func (r *RedisCluster) Connect() bool {
	return true
}

func (r *RedisCluster) singleton() redis.UniversalClient {
	return singleton(r.IsCache)
}

func (r *RedisCluster) hashKey(in string) string {
	if !r.HashKeys {
		// Not hashing? Return the raw key
		return in
	}

	return HashStr(in)
}

func (r *RedisCluster) fixKey(keyName string) string {
	return r.KeyPrefix + r.hashKey(keyName)
}

func (r *RedisCluster) cleanKey(keyName string) string {
	return strings.Replace(keyName, r.KeyPrefix, "", 1)
}

func (r *RedisCluster) up() error {
	if !Connected() {
		return ErrRedisIsDown
	}

	return nil
}

// GetKey will retrieve a key from the database.
func (r *RedisCluster) GetKey(keyName string) (string, error) {
	if err := r.up(); err != nil {
		return "", err
	}

	cluster := r.singleton()

	value, err := cluster.Get(cluster.Context(), r.fixKey(keyName)).Result()
	if err != nil {
		log.Debugf("Error trying to get value: %s", err.Error())

		return "", ErrKeyNotFound
	}

	return value, nil
}

// GetMultiKey gets multiple keys from the database.
func (r *RedisCluster) GetMultiKey(keys []string) ([]string, error) {
	if err := r.up(); err != nil {
		return nil, err
	}
	cluster := r.singleton()
	keyNames := make([]string, len(keys))
	copy(keyNames, keys)
	for index, val := range keyNames {
		keyNames[index] = r.fixKey(val)
	}

	result := make([]string, 0)

	switch v := cluster.(type) {
	case *redis.ClusterClient:
		{
			getCmds := make([]*redis.StringCmd, 0)
			pipe := v.Pipeline()
			for _, key := range keyNames {
				getCmds = append(getCmds, pipe.Get(cluster.Context(), key))
			}
			_, err := pipe.Exec(cluster.Context())
			if err != nil && !errors.Is(err, redis.Nil) {
				log.Debugf("Error trying to get value: %s", err.Error())

				return nil, ErrKeyNotFound
			}
			for _, cmd := range getCmds {
				result = append(result, cmd.Val())
			}
		}
	case *redis.Client:
		{
			values, err := cluster.MGet(cluster.Context(), keyNames...).Result()
			if err != nil {
				log.Debugf("Error trying to get value: %s", err.Error())

				return nil, ErrKeyNotFound
			}
			for _, val := range values {
				strVal := fmt.Sprint(val)
				if strVal == "<nil>" {
					strVal = ""
				}
				result = append(result, strVal)
			}
		}
	}

	for _, val := range result {
		if val != "" {
			return result, nil
		}
	}

	return nil, ErrKeyNotFound
}

// GetKeyTTL return ttl of the given key.
func (r *RedisCluster) GetKeyTTL(keyName string) (ttl int64, err error) {
	if err = r.up(); err != nil {
		return 0, err
	}
	duration, err := r.singleton().TTL(r.ctx, r.fixKey(keyName)).Result()

	return int64(duration.Seconds()), err
}

// GetRawKey return the value of the given key.
func (r *RedisCluster) GetRawKey(keyName string) (string, error) {
	if err := r.up(); err != nil {
		return "", err
	}
	value, err := r.singleton().Get(r.ctx, keyName).Result()
	if err != nil {
		log.Debugf("Error trying to get value: %s", err.Error())

		return "", ErrKeyNotFound
	}

	return value, nil
}

// GetExp return the expiry of the given key.
func (r *RedisCluster) GetExp(keyName string) (int64, error) {
	log.Debugf("Getting exp for key: %s", r.fixKey(keyName))
	if err := r.up(); err != nil {
		return 0, err
	}

	value, err := r.singleton().TTL(r.ctx, r.fixKey(keyName)).Result()
	if err != nil {
		log.Errorf("Error trying to get TTL: ", err.Error())

		return 0, ErrKeyNotFound
	}

	return int64(value.Seconds()), nil
}

// SetExp set expiry of the given key.
func (r *RedisCluster) SetExp(keyName string, timeout time.Duration) error {
	if err := r.up(); err != nil {
		return err
	}
	err := r.singleton().Expire(r.ctx, r.fixKey(keyName), timeout).Err()
	if err != nil {
		log.Errorf("Could not EXPIRE key: %s", err.Error())
	}

	return err
}

// SetKey will create (or update) a key value in the store.
func (r *RedisCluster) SetKey(keyName, session string, timeout time.Duration) error {
	log.Debugf("[STORE] SET Raw key is: %s", keyName)
	log.Debugf("[STORE] Setting key: %s", r.fixKey(keyName))

	if err := r.up(); err != nil {
		return err
	}
	err := r.singleton().Set(r.ctx, r.fixKey(keyName), session, timeout).Err()
	if err != nil {
		log.Errorf("Error trying to set value: %s", err.Error())

		return err
	}

	return nil
}

// SetRawKey set the value of the given key.
func (r *RedisCluster) SetRawKey(keyName, session string, timeout time.Duration) error {
	if err := r.up(); err != nil {
		return err
	}
	err := r.singleton().Set(r.ctx, keyName, session, timeout).Err()
	if err != nil {
		log.Errorf("Error trying to set value: %s", err.Error())

		return err
	}

	return nil
}

// SetNX set the value of the given key.
func (r *RedisCluster) SetNX(key string, value interface{}, expiration time.Duration) error {
	if err := r.up(); err != nil {
		return err
	}
	isSetKey, err := r.singleton().SetNX(r.ctx, key, value, expiration).Result()
	if err != nil {
		log.Errorf("Error trying to setnx value: %s", err.Error())

		return err
	}

	if !isSetKey {
		return ErrKeyNotFound
	}

	return nil
}

// Decrement will decrement a key in redis.
func (r *RedisCluster) Decrement(keyName string) (int64, error) {
	keyName = r.fixKey(keyName)
	log.Debugf("Decrementing key: %s", keyName)
	if err := r.up(); err != nil {
		return 0, err
	}
	val, err := r.singleton().Decr(r.ctx, keyName).Result()
	if err != nil {
		log.Errorf("Error trying to decrement value: %s", err.Error())
	}
	return val, nil
}

// Incrememnt will incrememnt a key in redis.
func (r *RedisCluster) Incrememnt(keyName string) (int64, error) {
	keyName = r.fixKey(keyName)
	log.Debugf("Incrementing key: %s", keyName)
	if err := r.up(); err != nil {
		return 0, err
	}
	val, err := r.singleton().Incr(r.ctx, keyName).Result()
	if err != nil {
		log.Errorf("Error trying to increment value: %s", err.Error())
	}
	return val, nil
}
func (r *RedisCluster) IncrememntBy(keyName string, value int64) (int64, error) {
	keyName = r.fixKey(keyName)
	log.Debugf("Incrementing key: %s", keyName)
	if err := r.up(); err != nil {
		return 0, err
	}
	val, err := r.singleton().IncrBy(r.ctx, keyName, value).Result()
	if err != nil {
		log.Errorf("Error trying to increment value: %s", err.Error())
	}
	return val, nil
}

// IncrememntWithExpire will increment a key in redis.
func (r *RedisCluster) IncrememntWithExpire(keyName string, expire int64) int64 {
	log.Debugf("Incrementing raw key: %s", keyName)
	if err := r.up(); err != nil {
		return 0
	}
	// This function uses a raw key, so we shouldn't call fixKey
	fixedKey := keyName
	val, err := r.singleton().Incr(r.ctx, fixedKey).Result()

	if err != nil {
		log.Errorf("Error trying to increment value: %s", err.Error())
	} else {
		log.Debugf("Incremented key: %s, val is: %d", fixedKey, val)
	}

	if val == 1 && expire > 0 {
		log.Debug("--> Setting Expire")
		r.singleton().Expire(r.ctx, fixedKey, time.Duration(expire)*time.Second)
	}

	return val
}

// GetKeys will return all keys according to the filter (filter is a prefix - e.g. tyk.keys.*).
func (r *RedisCluster) GetKeys(filter string) []string {
	if err := r.up(); err != nil {
		return nil
	}
	client := r.singleton()

	filterHash := ""
	if filter != "" {
		filterHash = r.hashKey(filter)
	}
	searchStr := r.KeyPrefix + filterHash + "*"
	log.Debugf("[STORE] Getting list by: %s", searchStr)

	fnFetchKeys := func(client *redis.Client) ([]string, error) {
		values := make([]string, 0)

		iter := client.Scan(r.ctx, 0, searchStr, 0).Iterator()
		for iter.Next(r.ctx) {
			values = append(values, iter.Val())
		}

		if err := iter.Err(); err != nil {
			return nil, err
		}

		return values, nil
	}

	var err error
	var values []string
	sessions := make([]string, 0)

	switch v := client.(type) {
	case *redis.ClusterClient:
		ch := make(chan []string)

		go func() {
			err = v.ForEachMaster(r.ctx, func(ctx context.Context, client *redis.Client) error {
				values, err = fnFetchKeys(client)
				if err != nil {
					return err
				}

				ch <- values

				return nil
			})
			close(ch)
		}()

		for res := range ch {
			sessions = append(sessions, res...)
		}
	case *redis.Client:
		sessions, err = fnFetchKeys(v)
	}

	if err != nil {
		log.Errorf("Error while fetching keys: %s", err)

		return nil
	}

	for i, v := range sessions {
		sessions[i] = r.cleanKey(v)
	}

	return sessions
}

// GetKeysAndValuesWithFilter will return all keys and their values with a filter.
func (r *RedisCluster) GetKeysAndValuesWithFilter(filter string) map[string]string {
	if err := r.up(); err != nil {
		return nil
	}
	keys := r.GetKeys(filter)
	if keys == nil {
		log.Error("Error trying to get filtered client keys")

		return nil
	}

	if len(keys) == 0 {
		return nil
	}

	for i, v := range keys {
		keys[i] = r.KeyPrefix + v
	}

	client := r.singleton()
	values := make([]string, 0)

	switch v := client.(type) {
	case *redis.ClusterClient:
		{
			getCmds := make([]*redis.StringCmd, 0)
			pipe := v.Pipeline()
			for _, key := range keys {
				getCmds = append(getCmds, pipe.Get(r.ctx, key))
			}
			_, err := pipe.Exec(r.ctx)
			if err != nil && !errors.Is(err, redis.Nil) {
				log.Errorf("Error trying to get client keys: %s", err.Error())

				return nil
			}

			for _, cmd := range getCmds {
				values = append(values, cmd.Val())
			}
		}
	case *redis.Client:
		{
			result, err := v.MGet(r.ctx, keys...).Result()
			if err != nil {
				log.Errorf("Error trying to get client keys: %s", err.Error())

				return nil
			}

			for _, val := range result {
				strVal := fmt.Sprint(val)
				if strVal == "<nil>" {
					strVal = ""
				}
				values = append(values, strVal)
			}
		}
	}

	m := make(map[string]string)
	for i, v := range keys {
		m[r.cleanKey(v)] = values[i]
	}

	return m
}

// GetKeysAndValues will return all keys and their values - not to be used lightly.
func (r *RedisCluster) GetKeysAndValues() map[string]string {
	return r.GetKeysAndValuesWithFilter("")
}

// DeleteKey will remove a key from the database.
func (r *RedisCluster) DeleteKey(keyName string) bool {
	if err := r.up(); err != nil {
		// log.Debug(err)
		return false
	}
	log.Debugf("DEL Key was: %s", keyName)
	log.Debugf("DEL Key became: %s", r.fixKey(keyName))
	n, err := r.singleton().Del(r.ctx, r.fixKey(keyName)).Result()
	if err != nil {
		log.Errorf("Error trying to delete key: %s", err.Error())
	}

	return n > 0
}

// DeleteAllKeys will remove all keys from the database.
func (r *RedisCluster) DeleteAllKeys() bool {
	if err := r.up(); err != nil {
		return false
	}
	n, err := r.singleton().FlushAll(r.ctx).Result()
	if err != nil {
		log.Errorf("Error trying to delete keys: %s", err.Error())
	}

	if n == "OK" {
		return true
	}

	return false
}

// DeleteRawKey will remove a key from the database without prefixing, assumes user knows what they are doing.
func (r *RedisCluster) DeleteRawKey(keyName string) bool {
	if err := r.up(); err != nil {
		return false
	}
	n, err := r.singleton().Del(r.ctx, keyName).Result()
	if err != nil {
		log.Errorf("Error trying to delete key: %s", err.Error())
	}

	return n > 0
}

// DeleteScanMatch will remove a group of keys in bulk.
func (r *RedisCluster) DeleteScanMatch(pattern string) bool {
	if err := r.up(); err != nil {
		return false
	}
	client := r.singleton()
	log.Debugf("Deleting: %s", pattern)

	fnScan := func(client *redis.Client) ([]string, error) {
		values := make([]string, 0)

		iter := client.Scan(r.ctx, 0, pattern, 0).Iterator()
		for iter.Next(r.ctx) {
			values = append(values, iter.Val())
		}

		if err := iter.Err(); err != nil {
			return nil, err
		}

		return values, nil
	}

	var err error
	var keys []string
	var values []string

	switch v := client.(type) {
	case *redis.ClusterClient:
		ch := make(chan []string)
		go func() {
			err = v.ForEachMaster(r.ctx, func(ctx context.Context, client *redis.Client) error {
				values, err = fnScan(client)
				if err != nil {
					return err
				}

				ch <- values

				return nil
			})
			close(ch)
		}()

		for vals := range ch {
			keys = append(keys, vals...)
		}
	case *redis.Client:
		keys, err = fnScan(v)
	}

	if err != nil {
		log.Errorf("SCAN command field with err: %s", err.Error())

		return false
	}

	if len(keys) > 0 {
		for _, name := range keys {
			log.Infof("Deleting: %s", name)
			err := client.Del(r.ctx, name).Err()
			if err != nil {
				log.Errorf("Error trying to delete key: %s - %s", name, err.Error())
			}
		}
		log.Infof("Deleted: %d records", len(keys))
	} else {
		log.Debug("RedisCluster called DEL - Nothing to delete")
	}

	return true
}

// DeleteKeys will remove a group of keys in bulk.
func (r *RedisCluster) DeleteKeys(keys []string) bool {
	if err := r.up(); err != nil {
		return false
	}
	if len(keys) > 0 {
		for i, v := range keys {
			keys[i] = r.fixKey(v)
		}

		log.Debugf("Deleting: %v", keys)
		client := r.singleton()
		switch v := client.(type) {
		case *redis.ClusterClient:
			{
				pipe := v.Pipeline()
				for _, k := range keys {
					pipe.Del(r.ctx, k)
				}

				if _, err := pipe.Exec(r.ctx); err != nil {
					log.Errorf("Error trying to delete keys: %s", err.Error())
				}
			}
		case *redis.Client:
			{
				_, err := v.Del(r.ctx, keys...).Result()
				if err != nil {
					log.Errorf("Error trying to delete keys: %s", err.Error())
				}
			}
		}
	} else {
		log.Debug("RedisCluster called DEL - Nothing to delete")
	}

	return true
}

// StartPubSubHandler will listen for a signal and run the callback for
// every subscription and message event.
func (r *RedisCluster) StartPubSubHandler(channel string, callback func(interface{})) error {
	if err := r.up(); err != nil {
		return err
	}
	client := r.singleton()
	if client == nil {
		return errors.New("redis connection failed")
	}

	pubsub := client.Subscribe(r.ctx, channel)
	defer pubsub.Close()

	if _, err := pubsub.Receive(r.ctx); err != nil {
		log.Errorf("Error while receiving pubsub message: %s", err.Error())

		return err
	}

	for msg := range pubsub.Channel() {
		callback(msg)
	}

	return nil
}

// StartPubSubHandlerWithCancel will listen for a signal and run the callback for
// every subscription and message event, and will be canceled when the context is done.
func (r *RedisCluster) StartPubSubHandlerWithCancel(ctx context.Context, channel string, callback func(interface{})) (chan struct{}, error) {
	if err := r.up(); err != nil {
		return nil, err
	}
	client := r.singleton()
	if client == nil {
		return nil, errors.New("redis connection failed")
	}

	ready := make(chan struct{})

	pubsub := client.Subscribe(r.ctx, channel)

	go func() {
		defer pubsub.Close()
		defer close(ready)

		if _, err := pubsub.Receive(r.ctx); err != nil {
			log.Errorf("Error while receiving pubsub message: %s", err.Error())
			return
		}

		ready <- struct{}{}

		ch := pubsub.Channel()
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				callback(msg)
			case <-ctx.Done():
				return
			}
		}
	}()

	return ready, nil
}

// Publish publish a message to the specify channel.
func (r *RedisCluster) Publish(channel, message string) error {
	if err := r.up(); err != nil {
		return err
	}
	err := r.singleton().Publish(r.ctx, channel, message).Err()
	if err != nil {
		log.Errorf("Error trying to set value: %s", err.Error())

		return err
	}

	return nil
}

// GetAndDeleteSet get and delete a key.
func (r *RedisCluster) GetAndDeleteSet(keyName string) []interface{} {
	log.Debugf("Getting raw key set: %s", keyName)
	if err := r.up(); err != nil {
		return nil
	}
	log.Debugf("keyName is: %s", keyName)
	fixedKey := r.fixKey(keyName)
	log.Debugf("Fixed keyname is: %s", fixedKey)

	client := r.singleton()

	var lrange *redis.StringSliceCmd
	_, err := client.TxPipelined(r.ctx, func(pipe redis.Pipeliner) error {
		lrange = pipe.LRange(r.ctx, fixedKey, 0, -1)
		pipe.Del(r.ctx, fixedKey)

		return nil
	})
	if err != nil {
		log.Errorf("Multi command failed: %s", err.Error())

		return nil
	}

	vals := lrange.Val()
	log.Debugf("Analytics returned: %d", len(vals))
	if len(vals) == 0 {
		return nil
	}

	log.Debugf("Unpacked vals: %d", len(vals))
	result := make([]interface{}, len(vals))
	for i, v := range vals {
		result[i] = v
	}

	return result
}

// AppendToSet append a value to the key set.
func (r *RedisCluster) AppendToSet(keyName, value string) {
	fixedKey := r.fixKey(keyName)
	log.Debug("Pushing to raw key list", log.String("keyName", keyName))
	log.Debug("Appending to fixed key list", log.String("fixedKey", fixedKey))
	if err := r.up(); err != nil {
		return
	}
	if err := r.singleton().RPush(r.ctx, fixedKey, value).Err(); err != nil {
		log.Errorf("Error trying to append to set keys: %s", err.Error())
	}
}

// Exists check if keyName exists.
func (r *RedisCluster) Exists(keyName string) (bool, error) {
	fixedKey := r.fixKey(keyName)
	log.Debug("Checking if exists", log.String("keyName", fixedKey))

	exists, err := r.singleton().Exists(r.ctx, fixedKey).Result()
	if err != nil {
		log.Errorf("Error trying to check if key exists: %s", err.Error())

		return false, err
	}
	if exists == 1 {
		return true, nil
	}

	return false, nil
}

// LPush set the value of the given key.
func (r *RedisCluster) LPush(key string, values ...interface{}) error {
	if err := r.up(); err != nil {
		return err
	}

	fixedKey := r.fixKey(key)
	err := r.singleton().LPush(r.ctx, fixedKey, values).Err()
	if err != nil {
		log.Errorf("Error trying to LPush value: %s", err.Error())

		return err
	}

	return nil
}

// LLen get the length of the given key.
func (r *RedisCluster) LLen(key string) (int64, error) {
	if err := r.up(); err != nil {
		return 0, err
	}

	fixedKey := r.fixKey(key)
	length, err := r.singleton().LLen(r.ctx, fixedKey).Result()
	if err != nil {
		log.Errorf("Error trying to LLen value: %s", err.Error())

		return 0, err
	}

	return length, nil
}

// RPop remove and get the last element in a list
func (r *RedisCluster) RPop(key string) (string, error) {
	if err := r.up(); err != nil {
		return "", err
	}

	fixedKey := r.fixKey(key)
	val, err := r.singleton().RPop(r.ctx, fixedKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrKeyNotFound
		}
		log.Errorf("Error trying to RPop value: %s", err.Error())
		return "", err
	}

	return val, nil
}

// BRPop remove and get the last element in a list, block until one is available
func (r *RedisCluster) BRPop(key string, timeout time.Duration) (string, error) {
	if err := r.up(); err != nil {
		return "", err
	}

	fixedKey := r.fixKey(key)
	results, err := r.singleton().BRPop(r.ctx, timeout, fixedKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrKeyNotFound
		}
		log.Errorf("Error trying to BRPop value: %s", err.Error())
		return "", err
	}

	return results[1], nil // BRPop returns [key, element]
}

// RemoveFromList delete an value from a list idetinfied with the keyName.
func (r *RedisCluster) RemoveFromList(keyName, value string) error {
	fixedKey := r.fixKey(keyName)

	log.Debug(
		"Removing value from list",
		log.String("keyName", keyName),
		log.String("fixedKey", fixedKey),
		log.String("value", value),
	)

	if err := r.singleton().LRem(r.ctx, fixedKey, 0, value).Err(); err != nil {
		log.Error(
			"LREM command failed",
			log.String("keyName", keyName),
			log.String("fixedKey", fixedKey),
			log.String("value", value),
			log.String("error", err.Error()),
		)

		return err
	}

	return nil
}

// GetListRange gets range of elements of list identified by keyName.
func (r *RedisCluster) GetListRange(keyName string, from, to int64) ([]string, error) {
	fixedKey := r.fixKey(keyName)

	elements, err := r.singleton().LRange(r.ctx, fixedKey, from, to).Result()
	if err != nil {
		log.Error(
			"LRANGE command failed",
			log.String(
				"keyName",
				keyName,
			),
			log.String("fixedKey", fixedKey),
			log.Int64("from", from),
			log.Int64("to", to),
			log.String("error", err.Error()),
		)

		return nil, err
	}

	return elements, nil
}

// AppendToSetPipelined append values to redis pipeline.
func (r *RedisCluster) AppendToSetPipelined(key string, values [][]byte) {
	if len(values) == 0 {
		return
	}

	fixedKey := r.fixKey(key)
	if err := r.up(); err != nil {
		log.Debug(err.Error())

		return
	}
	client := r.singleton()

	pipe := client.Pipeline()
	for _, val := range values {
		pipe.RPush(r.ctx, fixedKey, val)
	}

	if _, err := pipe.Exec(r.ctx); err != nil {
		log.Errorf("Error trying to append to set keys: %s", err.Error())
	}

	// if we need to set an expiration time
	if storageExpTime := int64(viper.GetDuration("analytics.storage-expiration-time")); storageExpTime != int64(-1) {
		// If there is no expiry on the analytics set, we should set it.
		exp, _ := r.GetExp(key)
		if exp == -1 {
			_ = r.SetExp(key, time.Duration(storageExpTime)*time.Second)
		}
	}
}

// GetSet return key set value.
func (r *RedisCluster) GetSet(keyName string) (map[string]string, error) {
	log.Debugf("Getting from key set: %s", keyName)
	log.Debugf("Getting from fixed key set: %s", r.fixKey(keyName))
	if err := r.up(); err != nil {
		return nil, err
	}
	val, err := r.singleton().SMembers(r.ctx, r.fixKey(keyName)).Result()
	if err != nil {
		log.Errorf("Error trying to get key set: %s", err.Error())

		return nil, err
	}

	result := make(map[string]string)
	for i, value := range val {
		result[strconv.Itoa(i)] = value
	}

	return result, nil
}

// AddToSet add value to key set.
func (r *RedisCluster) AddToSet(keyName string, values ...interface{}) error {
	log.Debugf("Pushing to raw key set: %s", keyName)
	log.Debugf("Pushing to fixed key set: %s", r.fixKey(keyName))
	if err := r.up(); err != nil {
		return err
	}
	err := r.singleton().SAdd(r.ctx, r.fixKey(keyName), values...).Err()
	if err != nil {
		log.Errorf("Error trying to append keys: %s", err.Error())

		return err
	}

	return nil
}

func (r *RedisCluster) AddSet(keyName string, values ...interface{}) (int64, error) {
	log.Debugf("AddSet to raw key set: %s", keyName)
	log.Debugf("AddSet to fixed key set: %s", r.fixKey(keyName))
	if err := r.up(); err != nil {
		return 0, err
	}
	res := r.singleton().SAdd(r.ctx, r.fixKey(keyName), values...)
	last, err := res.Result()
	if err != nil {
		log.Errorf("Error trying to append keys: %s", err.Error())

		return 0, err
	}
	return last, nil
}

// RemoveFromSet remove a value from key set.
func (r *RedisCluster) RemoveFromSet(keyName, value string) error {
	log.Debugf("Removing from raw key set: %s", keyName)
	log.Debugf("Removing from fixed key set: %s", r.fixKey(keyName))
	if err := r.up(); err != nil {
		log.Debug(err.Error())

		return err
	}
	err := r.singleton().SRem(r.ctx, r.fixKey(keyName), value).Err()
	if err != nil {
		log.Errorf("Error trying to remove keys: %s", err.Error())
	}
	return err
}

// RemoveAnyFromSet remove a value from key set.
func (r *RedisCluster) RemoveAnyFromSet(keyName string, value interface{}) {
	log.Debugf("Removing any from raw key set: %s", keyName)
	log.Debugf("Removing any from fixed key set: %s", r.fixKey(keyName))
	if err := r.up(); err != nil {
		log.Debug(err.Error())

		return
	}
	err := r.singleton().SRem(r.ctx, r.fixKey(keyName), value).Err()
	if err != nil {
		log.Errorf("Error trying to remove keys: %s", err.Error())
	}
}

// PopFromSet pop a value from key set.
func (r *RedisCluster) PopFromSet(keyName string) (string, error) {
	if err := r.up(); err != nil {
		log.Debug(err.Error())

		return "", err
	}
	val, err := r.singleton().SPop(r.ctx, r.fixKey(keyName)).Result()
	if err != nil {
		log.Errorf("Error trying to pop from set: %s", err.Error())

		return "", err
	}

	return val, nil
}

// IsMemberOfSet return whether the given value belong to key set.
func (r *RedisCluster) IsMemberOfSet(keyName, value string) (bool, error) {
	if err := r.up(); err != nil {
		log.Debug(err.Error())

		return false, err
	}
	val, err := r.singleton().SIsMember(r.ctx, r.fixKey(keyName), value).Result()
	if err != nil {
		log.Errorf("Error trying to check set member: %s", err.Error())

		return false, err
	}

	log.Debugf("SISMEMBER %s %s %v %v", keyName, value, val, err)

	return val, nil
}

// GetSetSize get the size of the key set.
func (r *RedisCluster) GetSetSize(keyName string) (int64, error) {
	if err := r.up(); err != nil {
		return 0, err
	}
	return r.singleton().SCard(r.ctx, r.fixKey(keyName)).Result()
}

// SetRollingWindow will append to a sorted set in redis and extract a timed window of values.
func (r *RedisCluster) SetRollingWindow(
	keyName string,
	per int64,
	valueOverride string,
	pipeline bool,
) (int, []interface{}) {
	log.Debugf("Incrementing raw key: %s", keyName)
	if err := r.up(); err != nil {
		log.Debug(err.Error())

		return 0, nil
	}
	log.Debugf("keyName is: %s", keyName)
	now := time.Now()
	log.Debugf("Now is: %v", now)
	onePeriodAgo := now.Add(time.Duration(-1*per) * time.Second)
	log.Debugf("Then is: %v", onePeriodAgo)

	client := r.singleton()
	var zrange *redis.StringSliceCmd

	pipeFn := func(pipe redis.Pipeliner) error {
		pipe.ZRemRangeByScore(r.ctx, keyName, "-inf", strconv.Itoa(int(onePeriodAgo.UnixNano())))
		zrange = pipe.ZRange(r.ctx, keyName, 0, -1)

		element := redis.Z{
			Score: float64(now.UnixNano()),
		}

		if valueOverride != "-1" {
			element.Member = valueOverride
		} else {
			element.Member = strconv.Itoa(int(now.UnixNano()))
		}

		pipe.ZAdd(r.ctx, keyName, &element)
		pipe.Expire(r.ctx, keyName, time.Duration(per)*time.Second)

		return nil
	}

	var err error
	if pipeline {
		_, err = client.Pipelined(r.ctx, pipeFn)
	} else {
		_, err = client.TxPipelined(r.ctx, pipeFn)
	}

	if err != nil {
		log.Errorf("Multi command failed: %s", err.Error())

		return 0, nil
	}

	values := zrange.Val()

	// Check actual value
	if values == nil {
		return 0, nil
	}

	intVal := len(values)
	result := make([]interface{}, len(values))

	for i, v := range values {
		result[i] = v
	}

	log.Debugf("Returned: %d", intVal)

	return intVal, result
}

// GetRollingWindow return rolling window.
func (r *RedisCluster) GetRollingWindow(keyName string, per int64, pipeline bool) (int, []interface{}) {
	if err := r.up(); err != nil {
		log.Debug(err.Error())

		return 0, nil
	}
	now := time.Now()
	onePeriodAgo := now.Add(time.Duration(-1*per) * time.Second)

	client := r.singleton()
	var zrange *redis.StringSliceCmd

	pipeFn := func(pipe redis.Pipeliner) error {
		pipe.ZRemRangeByScore(r.ctx, keyName, "-inf", strconv.Itoa(int(onePeriodAgo.UnixNano())))
		zrange = pipe.ZRange(r.ctx, keyName, 0, -1)

		return nil
	}

	var err error
	if pipeline {
		_, err = client.Pipelined(r.ctx, pipeFn)
	} else {
		_, err = client.TxPipelined(r.ctx, pipeFn)
	}
	if err != nil {
		log.Errorf("Multi command failed: %s", err.Error())

		return 0, nil
	}

	values := zrange.Val()

	// Check actual value
	if values == nil {
		return 0, nil
	}

	intVal := len(values)
	result := make([]interface{}, intVal)
	for i, v := range values {
		result[i] = v
	}

	log.Debugf("Returned: %d", intVal)

	return intVal, result
}

// GetKeyPrefix returns storage key prefix.
func (r *RedisCluster) GetKeyPrefix() string {
	return r.KeyPrefix
}

// AddToSortedSetAndTrimPipelined 添加到有序集合并修剪
func (r *RedisCluster) AddToSortedSetAndTrimPipelined(key string, values map[float64]string, maxElements int64) error {
	if len(values) == 0 {
		return nil
	}

	fixedKey := r.fixKey(key)
	if err := r.up(); err != nil {
		return err
	}

	pipe := r.singleton().Pipeline()
	for score, val := range values {
		pipe.ZAdd(r.ctx, fixedKey, &redis.Z{Score: score, Member: val})
	}

	pipe.ZRemRangeByRank(r.ctx, fixedKey, 0, -(maxElements + 1))

	if _, err := pipe.Exec(r.ctx); err != nil {
		return err
	}

	return nil
}

// AddToSortedSet adds value with given score to sorted set identified by keyName.
func (r *RedisCluster) AddToSortedSet(keyName, value string, score float64) (int64, error) {
	fixedKey := r.fixKey(keyName)

	log.Debug("Pushing raw key to sorted set", log.String("keyName", keyName), log.String("fixedKey", fixedKey))

	if err := r.up(); err != nil {
		log.Debug(err.Error())

		return 0, err
	}
	member := redis.Z{Score: score, Member: value}
	val, err := r.singleton().ZAdd(r.ctx, fixedKey, &member).Result()
	if err != nil {
		log.Error(
			"ZADD command failed",
			log.String("keyName", keyName),
			log.String("fixedKey", fixedKey),
			log.String("error", err.Error()),
		)
		return 0, err
	}
	return val, nil
}

// GetSortedSetRange gets range of elements of sorted set identified by keyName.
func (r *RedisCluster) GetSortedSetRange(keyName, scoreFrom, scoreTo string) ([]string, []float64, error) {
	fixedKey := r.fixKey(keyName)
	log.Debug(
		"Getting sorted set range",
		log.String(
			"keyName",
			keyName,
		),
		log.String("fixedKey", fixedKey),
		log.String("scoreFrom", scoreFrom),
		log.String("scoreTo", scoreTo),
	)

	args := redis.ZRangeBy{Min: scoreFrom, Max: scoreTo}
	values, err := r.singleton().ZRangeByScoreWithScores(r.ctx, fixedKey, &args).Result()
	if err != nil {
		log.Error(
			"ZRANGEBYSCORE command failed",
			log.String(
				"keyName",
				keyName,
			),
			log.String("fixedKey", fixedKey),
			log.String("scoreFrom", scoreFrom),
			log.String("scoreTo", scoreTo),
			log.String("error", err.Error()),
		)

		return nil, nil, err
	}

	if len(values) == 0 {
		return nil, nil, nil
	}

	elements := make([]string, len(values))
	scores := make([]float64, len(values))

	for i, v := range values {
		elements[i] = fmt.Sprint(v.Member)
		scores[i] = v.Score
	}

	return elements, scores, nil
}

// GetZSetSize get the size of the key set.
func (r *RedisCluster) GetZSetSize(keyName string) (int64, error) {
	if err := r.up(); err != nil {
		return 0, err
	}
	return r.singleton().ZCard(r.ctx, r.fixKey(keyName)).Result()
}

// ZSCAN  to retrieve all data from sorted set
func (r *RedisCluster) ZSCAN(keyName string) ([]string, []float64, error) {
	fixedKey := r.fixKey(keyName)
	log.Debug(
		"Executing ZSCAN operation to retrieve all data",
		log.String("keyName", keyName),
		log.String("fixedKey", fixedKey),
	)

	var cursor uint64
	var elements []string
	var scores []float64

	for {
		keys, nextCursor, err := r.singleton().ZScan(r.ctx, fixedKey, cursor, "", 0).Result()
		if err != nil {
			log.Error(
				"ZSCAN command execution failed",
				log.String("keyName", keyName),
				log.String("fixedKey", fixedKey),
				log.String("error", err.Error()),
			)
			return nil, nil, err
		}

		for i := 0; i < len(keys); i += 2 {
			elements = append(elements, keys[i])
			score, err := strconv.ParseFloat(keys[i+1], 64)
			if err != nil {
				log.Error(
					"Failed to parse score",
					log.String("keyName", keyName),
					log.String("fixedKey", fixedKey),
					log.String("score", keys[i+1]),
					log.String("error", err.Error()),
				)
				return nil, nil, err
			}
			scores = append(scores, score)
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return elements, scores, nil
}

// RemoveSortedSetRange removes range of elements from sorted set identified by keyName.
func (r *RedisCluster) RemoveSortedSetRange(keyName, scoreFrom, scoreTo string) error {
	fixedKey := r.fixKey(keyName)

	log.Debug(
		"Removing sorted set range",
		log.String(
			"keyName",
			keyName,
		),
		log.String("fixedKey", fixedKey),
		log.String("scoreFrom", scoreFrom),
		log.String("scoreTo", scoreTo),
	)
	if err := r.singleton().ZRemRangeByScore(r.ctx, fixedKey, scoreFrom, scoreTo).Err(); err != nil {
		log.Debug(
			"ZREMRANGEBYSCORE command failed",
			log.String("keyName", keyName),
			log.String("fixedKey", fixedKey),
			log.String("scoreFrom", scoreFrom),
			log.String("scoreTo", scoreTo),
			log.String("error", err.Error()),
		)

		return err
	}

	return nil
}

func (r *RedisCluster) RemoveSortedSetByMember(keyName string, member ...string) (int64, error) {
	fixedKey := r.fixKey(keyName)

	log.Debug(
		"Removing sorted set range",
		log.String(
			"keyName",
			keyName,
		),
		log.String("fixedKey", fixedKey),
		log.Strings("member", member),
	)
	val, err := r.singleton().ZRem(r.ctx, fixedKey, member).Result()
	if err != nil {
		log.Debug(
			"ZRem command failed",
			log.String("keyName", keyName),
			log.String("fixedKey", fixedKey),
			log.Strings("member", member),
			log.String("error", err.Error()),
		)
		return 0, err
	}

	return val, nil
}

// messages, err := rdb.ZRevRange(ctx, key, offset, offset+limit-1).Result()
// 有序集合中按分数从高到低（降序）获取指定范围内的成员
func (r *RedisCluster) GetRangeSortedByScoreDesc(keyName string, start, stop int64) (int, []interface{}) {
	fixedKey := r.fixKey(keyName)

	log.Debug(
		"Get range sorted order by score desc",
		log.String(
			"keyName",
			keyName,
		),
		log.String("fixedKey", fixedKey),
		log.Int64("start", start),
		log.Int64("stop", stop),
	)
	var zrange *redis.StringSliceCmd
	zrange = r.singleton().ZRevRange(r.ctx, fixedKey, start, stop)
	// 	ZRevRange(ctx context.Context, key string, start, stop int64) *StringSliceCmd
	if err := zrange.Err(); err != nil {
		log.Debug(
			"ZRevRange command failed",
			log.String("keyName", keyName),
			log.String("fixedKey", fixedKey),
			log.Int64("start", start),
			log.Int64("stop", stop),
			log.String("error", err.Error()),
		)
		return 0, nil
	}

	values := zrange.Val()

	// Check actual value
	if values == nil {
		return 0, nil
	}

	intVal := len(values)
	result := make([]interface{}, intVal)
	for i, v := range values {
		result[i] = v
	}

	log.Debugf("Returned: %d", intVal)

	return intVal, result
}

// messages, err := rdb.ZRange(ctx, key, offset, offset+limit-1).Result()
// 有序集合中按分数从高到低（降序）获取指定范围内的成员
func (r *RedisCluster) GetRangeSortedByScoreAsc(keyName string, start, stop int64) (int, []interface{}) {
	fixedKey := r.fixKey(keyName)

	log.Debug(
		"Get range sorted order by score asc",
		log.String(
			"keyName",
			keyName,
		),
		log.String("fixedKey", fixedKey),
		log.Int64("start", start),
		log.Int64("stop", stop),
	)
	var zrange *redis.StringSliceCmd
	zrange = r.singleton().ZRange(r.ctx, fixedKey, start, stop)
	// ZRange(ctx context.Context, key string, start, stop int64) *StringSliceCmd
	if err := zrange.Err(); err != nil {
		log.Debug(
			"ZRange command failed",
			log.String("keyName", keyName),
			log.String("fixedKey", fixedKey),
			log.Int64("start", start),
			log.Int64("stop", stop),
			log.String("error", err.Error()),
		)
		return 0, nil
	}

	values := zrange.Val()

	// Check actual value
	if values == nil {
		return 0, nil
	}

	intVal := len(values)
	result := make([]interface{}, intVal)
	for i, v := range values {
		result[i] = v
	}

	log.Debugf("Returned: %d", intVal)

	return intVal, result
}

func (r *RedisCluster) MSet(kv map[string]string, duration time.Duration) error {
	if err := r.up(); err != nil {
		return err
	}
	cluster := r.singleton()
	switch v := cluster.(type) {
	case *redis.ClusterClient:
		{
			pip := v.TxPipeline()
			defer pip.Close()
			for key, value := range kv {
				pip.Set(r.ctx, key, value, duration)
			}
			exec, err := pip.Exec(r.ctx)
			if err != nil {
				return err
			}
			log.Infof("exec result size %d", len(exec))
			for _, cmder := range exec {
				status := cmder.(*redis.StatusCmd)
				if status.Val() != "OK" {
					log.Infof("exec result %v", cmder)
				}
			}
		}
	case *redis.Client:
		{
			result, err := v.MSet(r.ctx, kv).Result()
			log.Infof("exec result %v,%v", result, err)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *RedisCluster) HDel(key string, fields ...string) int64 {
	if err := r.up(); err != nil {
		return 0
	}

	fixedKey := r.fixKey(key)

	log.Debug(
		"Deleting hash fields",
		log.String("key", key),
		log.String("fixedKey", fixedKey),
		log.Strings("fields", fields),
	)

	n, err := r.singleton().HDel(r.ctx, fixedKey, fields...).Result()
	if err != nil {
		log.Error(
			"HDel command failed",
			log.String("key", key),
			log.String("fixedKey", fixedKey),
			log.Strings("fields", fields),
			log.String("error", err.Error()),
		)
		return 0
	}

	return n
}

func (r *RedisCluster) HSet(key string, values ...interface{}) error {
	if err := r.up(); err != nil {
		return err
	}

	fixedKey := r.fixKey(key)

	log.Debug(
		"Setting hash fields",
		log.String("key", key),
		log.String("fixedKey", fixedKey),
	)

	_, err := r.singleton().HSet(r.ctx, fixedKey, values...).Result()
	if err != nil {
		log.Error(
			"HSet command failed",
			log.String("key", key),
			log.String("fixedKey", fixedKey),
			log.String("error", err.Error()),
		)
		return err
	}

	return nil
}

// HIncrBy increments the integer value of a hash field by the given number
func (r *RedisCluster) HIncrBy(key, field string, incr int64) (int64, error) {
	if err := r.up(); err != nil {
		return 0, err
	}

	fixedKey := r.fixKey(key)

	log.Debug(
		"Incrementing hash field",
		log.String("key", key),
		log.String("fixedKey", fixedKey),
		log.String("field", field),
		log.Int64("increment", incr),
	)

	value, err := r.singleton().HIncrBy(r.ctx, fixedKey, field, incr).Result()
	if err != nil {
		log.Error(
			"HIncrBy command failed",
			log.String("key", key),
			log.String("fixedKey", fixedKey),
			log.String("field", field),
			log.Int64("increment", incr),
			log.String("error", err.Error()),
		)
		return 0, err
	}

	return value, nil
}

func (r *RedisCluster) HGet(key, field string) (string, error) {
	if err := r.up(); err != nil {
		return "", err
	}

	fixedKey := r.fixKey(key)

	log.Debug(
		"Getting hash field",
		log.String("key", key),
		log.String("fixedKey", fixedKey),
		log.String("field", field),
	)

	value, err := r.singleton().HGet(r.ctx, fixedKey, field).Result()
	if err != nil {
		if err == redis.Nil {
			log.Debug(
				"Hash field does not exist",
				log.String("key", key),
				log.String("fixedKey", fixedKey),
				log.String("field", field),
			)
			return "", ErrKeyNotFound
		}
		log.Error(
			"HGet command failed",
			log.String("key", key),
			log.String("fixedKey", fixedKey),
			log.String("field", field),
			log.String("error", err.Error()),
		)
		return "", err
	}

	return value, nil
}

// HMGet retrieves the values associated with the specified fields in the hash stored at key
func (r *RedisCluster) HMGet(key string, fields ...string) ([]interface{}, error) {
	if err := r.up(); err != nil {
		return nil, err
	}

	fixedKey := r.fixKey(key)

	log.Debug(
		"Getting multiple hash fields",
		log.String("key", key),
		log.String("fixedKey", fixedKey),
		log.Strings("fields", fields),
	)

	values, err := r.singleton().HMGet(r.ctx, fixedKey, fields...).Result()
	if err != nil {
		log.Error(
			"HMGet command failed",
			log.String("key", key),
			log.String("fixedKey", fixedKey),
			log.Strings("fields", fields),
			log.String("error", err.Error()),
		)
		return nil, err
	}

	return values, nil
}

func (r *RedisCluster) HGetAll(key string) (map[string]string, error) {
	if err := r.up(); err != nil {
		return nil, err
	}

	fixedKey := r.fixKey(key)

	log.Debug(
		"Getting all hash fields",
		log.String("key", key),
		log.String("fixedKey", fixedKey),
	)

	values, err := r.singleton().HGetAll(r.ctx, fixedKey).Result()
	if err != nil {
		log.Error(
			"HGetAll command failed",
			log.String("key", key),
			log.String("fixedKey", fixedKey),
			log.String("error", err.Error()),
		)
		return nil, err
	}

	return values, nil
}

const (
	LockSuccess = iota
	LockFail
	LockTimeout
	LockContention
)
const (
	maxAttempts       = 10
	baseDelay         = 10 * time.Millisecond
	maxJitterRatio    = 0.3 // TryLock 重试间隔抖动比例
	expiryJitterRatio = 0.2 // MutexLock 过期时间抖动比例，与 TryLock 同款：base + [0, ratio*base]
	lockChannelPrefix = "{%s}:lock:notify"
	lockScript        = `if redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2]) then
        return 1
    else
        return 0
    end`
	unlockLuaScript = `
    if redis.call("get", KEYS[1]) == ARGV[1] then
        redis.call("del", KEYS[1])
        redis.call("publish", KEYS[2], "1")
        return 1
    end
    return 0`
)

func (r *RedisCluster) Lock(ctx context.Context, key, threadId string) bool {
	/*i := r.IncrememntWithExpire(key, 30)
	if i == 1 {
		return true
	}
	return false*/

	// 默认锁过期时间30秒
	expireTime := 30000
	result, err := r.Eval(lockScript,
		[]string{key},
		[]interface{}{threadId, expireTime},
	)

	if err != nil {
		log.L(ctx).Errorf("加锁失败: %s %v", threadId, err)
		return false
	}
	return result.(int64) == 1
}
func (r *RedisCluster) Unlock(cxt context.Context, key, threadId string) error {
	/*r.DeleteRawKey(key)
	return nil*/
	// 原子化释放锁并发送通知
	result, err := r.Eval(unlockLuaScript, []string{
		key,
		fmt.Sprintf(lockChannelPrefix, key),
	}, threadId)
	if err != nil {
		log.L(cxt).Errorf("解锁失败: %s %v", threadId, err)
		return fmt.Errorf("解锁失败: %v", err)
	}
	if result.(int64) == 0 {
		log.L(cxt).Errorf("锁持有者不匹配: %s %v", threadId, err)
		return errors.New("锁持有者不匹配")
	}
	return nil
}

// expiryWithJitter 对 expiry 施加动态抖动，与 TryLock 同款公式：base + random(0, ratio)*base
// 公式: expiry + rand.Float64() * expiryJitterRatio * expiry，结果在 [expiry, expiry*(1+ratio)]
func expiryWithJitter(expiry time.Duration) time.Duration {
	if expiry <= 0 {
		return expiry
	}
	base := float64(expiry)
	src := rand.NewSource(time.Now().UnixNano())
	ran := rand.New(src)
	jitter := time.Duration(ran.Float64() * expiryJitterRatio * base)
	return expiry + jitter
}

func (r *RedisCluster) MutexLock(ctx context.Context, key, threadId string, expiry time.Duration) *redsync.Mutex {
	// Create a pool with go-redis (or redigo) which is the pool redisync will
	// use while communicating with Redis. This can also be any pool that
	// implements the `redis.Pool` interface.

	client := r.singleton()
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)

	// Create an instance of redisync to be used to obtain a mutual exclusion
	// lock.
	rs := redsync.New(pool)

	// 对 expiry 施加动态抖动，避免大量锁同时过期
	exp := expiryWithJitter(expiry)

	// Obtain a new mutex by using the same name for all instances wanting the
	// same lock.
	log.L(ctx).Infof("--> MutexLock key：%+v", key)
	mutex := rs.NewMutex(key, redsync.WithExpiry(exp))
	log.L(ctx).Infof("--> MutexLock start：%+v", mutex)
	return mutex
}
func (r *RedisCluster) TryLock(ctx context.Context, key, threadId string) (int, error) {
	if err := r.up(); err != nil {
		return LockFail, err
	}
	pubsub := r.singleton().Subscribe(ctx, fmt.Sprintf(lockChannelPrefix, key))
	defer pubsub.Close()
	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return LockTimeout, context.Cause(ctx)
		default:
			if acquired := r.Lock(ctx, key, threadId); acquired {
				return LockSuccess, nil
			}
			// 动态抖动计算
			base := float64(baseDelay) * math.Pow(2, float64(attempt))
			src := rand.NewSource(time.Now().UnixNano())
			ran := rand.New(src)
			jitter := time.Duration(ran.Float64() * maxJitterRatio * base)
			delay := time.Duration(base) + jitter
			// 休眠
			select {
			case <-time.After(delay):
				// 正常等待
				log.L(ctx).Infof("锁竞争第%d 次", attempt)
			case <-ctx.Done():
				return LockTimeout, context.Cause(ctx)
			case <-pubsub.Channel(): // 等待期间收到通知
				attempt--
				continue
			}
		}
	}
	//锁竞争次数耗尽
	return LockContention, fmt.Errorf("exhausted %d attempts", maxAttempts)
}

func (r *RedisCluster) SetContext(ctx context.Context) {
	r.ctx = ctx
}

func (r *RedisCluster) Eval(script string, keys []string, args ...interface{}) (interface{}, error) {
	if err := r.up(); err != nil {
		return nil, err
	}

	return r.singleton().Eval(r.ctx, script, keys, args...).Result()
}

func (r *RedisCluster) PFAdd(key string, values ...interface{}) error {
	if err := r.up(); err != nil {
		return err
	}

	fixedKey := r.fixKey(key)

	_, err := r.singleton().PFAdd(r.ctx, fixedKey, values...).Result()
	if err != nil {
		log.Error(
			"PFAdd command failed",
			log.String("key", key),
			log.String("fixedKey", fixedKey),
			log.String("error", err.Error()),
		)
		return err
	}

	return nil
}

func (r *RedisCluster) PFCount(key string) (int64, error) {
	if err := r.up(); err != nil {
		return 0, err
	}

	fixedKey := r.fixKey(key)

	count, err := r.singleton().PFCount(r.ctx, fixedKey).Result()
	if err != nil {
		log.Error(
			"PFCount command failed",
			log.String("key", key),
			log.String("fixedKey", fixedKey),
			log.String("error", err.Error()),
		)
		return 0, err
	}

	return count, nil
}
