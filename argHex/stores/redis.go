package stores

// import redis
import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisStore(redis_host string, redis_port string, redis_user string, redis_pass string, redis_db int) (*Redis, error) {
	redisDB := new(Redis)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second+10)

	defer cancel()

	//

	redisDB.client = redis.NewClient(&redis.Options{
		Addr:         redis_host + ":" + redis_port,
		Username:     redis_user,
		Password:     redis_pass,
		DB:           redis_db,
		PoolSize:     100,
		PoolTimeout:  time.Second * 10,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		MaxIdleConns: 10,
	})

	redisDB.ctx = ctx

	// test connection
	_, err := redisDB.client.Ping(ctx).Result()

	if nil != err {
		return nil, err
	}

	return redisDB, nil
}
