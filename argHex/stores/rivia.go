package stores

// import redis
import (
	"context"
	"log"
	"time"
)

type Rivia struct {
	redis *Redis
	ctx   context.Context
	db    int
}

func NewRivia(redis *Redis, db int) *Rivia {
	ctx := context.Background()
	return &Rivia{
		redis: redis,
		ctx:   ctx,
		db:    db,
	}
}

func (r *Rivia) Get(id string) string {
	data := r.redis.client.Get(r.ctx, id)
	log.Println(data)

	return data.Val()
}

func (r *Rivia) Set(id string, expires time.Duration, data interface{}) error {
	// change db to r.db
	r.redis.client.Conn().Select(r.ctx, r.db)

	err := r.redis.client.Set(r.ctx, id, data, expires).Err()

	if nil != err {
		return err
	}

	return nil
}

func (r *Rivia) Remove(id string) error {

	return nil
}
