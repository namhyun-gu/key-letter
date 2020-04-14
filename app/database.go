package app

import (
	"encoding/json"
	"time"
	
	"github.com/go-redis/redis"
)

type Database interface {
	CreateStore(key string, store Store, expiration time.Duration) error

	ReadStore(key string) (*Store, error)

	DeleteStore(key string) error
}

type DatabaseChannel interface {
	Subscribe(channel string) *redis.PubSub

	Publish(channel string, message interface{}) (int, error)
}

type RedisDatabase struct {
	Client *redis.Client
}

func (db *RedisDatabase) CreateStore(key string, store Store, expiration time.Duration) error {
	storeJson, err := json.Marshal(store)
	if err != nil {
		return err
	}
	err = db.Client.Set(key, string(storeJson), time.Second*expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (db *RedisDatabase) ReadStore(key string) (*Store, error) {
	var store Store
	storeJson, err := db.Client.Get(key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(storeJson), &store)
	if err != nil {
		return nil, err
	}
	return &store, err
}

func (db *RedisDatabase) DeleteStore(key string) error {
	return db.Client.Del(key).Err()
}

type RedisDatabaseChannel struct {
	Client *redis.Client
}

func (db *RedisDatabaseChannel) Subscribe(channel string) *redis.PubSub {
	return db.Client.Subscribe(channel)
}

func (db *RedisDatabaseChannel) Publish(channel string, message interface{}) (int, error) {
	publish := db.Client.Publish(channel, message)
	err := publish.Err()
	if err != nil {
		return 0, err
	}
	return int(publish.Val()), nil
}
