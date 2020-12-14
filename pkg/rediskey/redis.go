package rediskey

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
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

func SetVersionAndCommit(ctx context.Context, client *redis.Client, version, commit string) {
	err := client.Set(ctx, Version, version, 0).Err()
	if err != nil {
		log.Println(err)
	}

	err = client.Set(ctx, Commit, commit, 0).Err()
	if err != nil {
		log.Println(err)
	}
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

const TotalGameExpiration = time.Minute * 5
const TotalUsersExpiration = time.Minute * 5

const NotFound = -1

func GetTotalGames(ctx context.Context, client *redis.Client) int64 {
	v, err := client.Get(ctx, TotalGames).Int64()
	if err == nil {
		return v
	}
	return NotFound
}

func GetTotalUsers(ctx context.Context, client *redis.Client) int64 {
	v, err := client.Get(ctx, TotalUsers).Int64()
	if err == nil {
		return v
	}
	return NotFound
}

func RefreshTotalUsers(ctx context.Context, client *redis.Client, pool *pgxpool.Pool) int64 {
	v := queryTotalUsers(ctx, pool)
	if v != NotFound {
		err := client.Set(ctx, TotalUsers, v, TotalUsersExpiration).Err()
		if err != nil {
			log.Println(err)
		}
	}
	return v
}

func RefreshTotalGames(ctx context.Context, client *redis.Client, pool *pgxpool.Pool) int64 {
	v := queryTotalGames(ctx, pool)
	if v != NotFound {
		err := client.Set(ctx, TotalGames, v, TotalGameExpiration).Err()
		if err != nil {
			log.Println(err)
		}
	}
	return v
}
