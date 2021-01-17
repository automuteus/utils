package capture

import (
	"context"
	"encoding/json"
	"github.com/automuteus/utils/pkg/rediskey"
	"github.com/go-redis/redis/v8"
	"time"
)

type EventType int

const (
	Connection EventType = iota
	Lobby
	State
	Player
	GameOver
)

type Event struct {
	EventType EventType   `json:"type"`
	Payload   interface{} `json:"payload"`
}

const EventTTLSeconds = 3600

func PushEvent(ctx context.Context, redis *redis.Client, connCode string, jobType EventType, payload string) error {
	event := Event{
		EventType: jobType,
		Payload:   payload,
	}
	jBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	count, err := redis.RPush(ctx, rediskey.EventsNamespace+connCode, string(jBytes)).Result()
	if err == nil {
		notify(ctx, redis, connCode)
	}

	// new list
	if count < 2 {
		// log.Printf("Set TTL for List")
		redis.Expire(ctx, rediskey.EventsNamespace+connCode, EventTTLSeconds*time.Second)
	}

	return err
}

func notify(ctx context.Context, redis *redis.Client, connCode string) {
	redis.Publish(ctx, rediskey.EventsNamespace+connCode+":notify", true)
}

func Subscribe(ctx context.Context, redis *redis.Client, connCode string) *redis.PubSub {
	return redis.Subscribe(ctx, rediskey.EventsNamespace+connCode+":notify")
}

func PopEvent(ctx context.Context, redis *redis.Client, connCode string) (Event, error) {
	str, err := redis.LPop(ctx, rediskey.EventsNamespace+connCode).Result()

	j := Event{}
	if err != nil {
		return j, err
	}
	err = json.Unmarshal([]byte(str), &j)
	return j, err
}

func Ack(ctx context.Context, redis *redis.Client, connCode string) {
	redis.Publish(ctx, rediskey.EventsNamespace+connCode+":ack", true)
}

func AckSubscribe(ctx context.Context, redis *redis.Client, connCode string) *redis.PubSub {
	return redis.Subscribe(ctx, rediskey.EventsNamespace+connCode+":ack")
}
