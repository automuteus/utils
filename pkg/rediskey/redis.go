package rediskey

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

func GetVersionAndCommit(ctx context.Context, client *redis.Client) (string, string) {
	v, err := client.Get(ctx, Version).Result()
	if err != nil {
		log.Println(err)
	}
	c, err := client.Get(ctx, Commit).Result()
	if err != nil {
		log.Println(err)
	}
	return v, c
}

func GetGuildCounter(ctx context.Context, client *redis.Client) int64 {
	count, err := client.SCard(ctx, TotalGuildsSet).Result()
	if err != nil {
		log.Println(err)
		return 0
	}
	return count
}

func GetActiveGames(ctx context.Context, client *redis.Client, secs int64) int64 {
	now := time.Now()
	before := now.Add(-(time.Second * time.Duration(secs)))
	count, err := client.ZCount(ctx, ActiveGamesZSet, fmt.Sprintf("%d", before.Unix()), fmt.Sprintf("%d", now.Unix())).Result()
	if err != nil {
		log.Println(err)
		return 0
	}
	return count
}
