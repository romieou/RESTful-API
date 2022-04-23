package redis

import (
	"fmt"
	"rest/models"
	"rest/myerrors"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

// counter defines key under which counter is stored in redis
const counter = "counter"

type RedisCache struct {
	redisConn  *redis.Client
	expiration time.Duration
	mx         *sync.Mutex
}

// NewredisCache returns new redis client built upon expiration time passed.
// Users should pass in zero to indicate no expiration time.
func NewRedisCache(exp time.Duration) (models.RedisInterface, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	pong, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	fmt.Println(pong)
	return &RedisCache{
		redisConn:  client,
		expiration: exp,
		mx:         &sync.Mutex{},
	}, nil
}

// GetCounter retrieves current value of counter
func (r *RedisCache) GetCounter() (string, error) {
	value, err := r.redisConn.Get(counter).Result()
	if err != nil {
		if err == redis.Nil {
			return "0", r.redisConn.Set(counter, "0", r.expiration).Err()
		}
		return "", err
	}
	return value, nil
}

// SetCounter increments counter by value passed as argument
func (r *RedisCache) SetCounter(n int) (string, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	val, err := r.GetCounter()
	if err != nil {
		return "", err
	}
	curr, err := strconv.Atoi(val)
	if err != nil {
		return "", myerrors.ErrNonNumericCounter
	}
	add := curr + n
	if add < 0 {
		return "", myerrors.ErrNegativeCounter
	}
	res := strconv.Itoa(add)
	return res, r.redisConn.Set(counter, res, r.expiration).Err()
}

// Set sets value in redis for given key
func (r *RedisCache) Set(key string, value interface{}) error {
	return r.redisConn.Set(key, value, r.expiration).Err()
}

// Get gets value from redis using provided key
func (r *RedisCache) Get(key string) (string, error) {
	value, err := r.redisConn.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", myerrors.ErrNotFound
		}
		return "", err
	}
	return value, nil
}
