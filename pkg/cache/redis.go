package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"

	"github.com/abtasty/flagship-go-sdk/pkg/logging"
	"github.com/go-redis/redis/v8"
)

// RedisManager represents a redis db manager object
type RedisManager struct {
	client *redis.Client
}

// RedisOptions are the options necessary to make redis cache manager work
type RedisOptions struct {
	Host      string
	Username  string
	Password  string
	TLSConfig *tls.Config
	Db        int
}

var redisLogger = logging.CreateLogger("redis")
var rdb *redis.Client
var ctx = context.Background()

// WithRedisOptions configures redis options for manager
func WithRedisOptions(redisOptions RedisOptions) func(options *Options) {
	return func(options *Options) {
		options.cacheType = Redis
		options.RedisOptions = redisOptions
	}
}

func initRedisManager(options RedisOptions) (Manager, error) {
	redisLogger.Info("Connecting to server...")
	rdb = redis.NewClient(&redis.Options{
		Addr:      options.Host,
		Username:  options.Username,
		TLSConfig: options.TLSConfig,
		Password:  options.Password,
		DB:        options.Db,
	})
	_, err := rdb.Ping(ctx).Result()

	if err != nil {
		return nil, err
	}

	return &RedisManager{
		client: rdb,
	}, nil
}

// Set saves the campaigns in cache for this visitor
func (m *RedisManager) Set(visitorID string, campaignCache map[string]*CampaignCache) (err error) {
	if m.client == nil {
		return errors.New("Redis cache manager not initialized")
	}

	data, err := json.Marshal(campaignCache)
	if err != nil {
		return err
	}

	redisLogger.Info("Setting visitor cache")
	cmd := m.client.Set(ctx, visitorID, string(data), 0)
	_, err = cmd.Result()

	return err
}

// Get returns the campaigns in cache for this visitor
func (m *RedisManager) Get(visitorID string) (cache map[string]*CampaignCache, err error) {
	if m.client == nil {
		return nil, errors.New("Redis cache manager not initialized")
	}

	redisLogger.Info("Getting visitor cache")
	cmd := m.client.Get(ctx, visitorID)
	data, err := cmd.Bytes()

	if err != nil {
		return nil, err
	}

	cache = make(map[string]*CampaignCache)
	err = json.Unmarshal(data, &cache)

	return cache, err
}
