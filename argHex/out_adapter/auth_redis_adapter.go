package out_adapter

import (
	"time"

	"github.com/argSea/portfolio_blog_api/argHex/out_port"
	"github.com/argSea/portfolio_blog_api/argHex/stores"
)

type authRedisAdapter struct {
	store *stores.Rivia
}

func NewAuthRedisAdapter(store *stores.Rivia) out_port.AuthRepo {
	a := authRedisAdapter{
		store: store,
	}

	return a
}

func (a authRedisAdapter) Get(id string) string {
	data := a.store.Get(id)

	return data
}

func (a authRedisAdapter) Store(token string, expires time.Duration, data interface{}) error {
	err := a.store.Set(token, expires, data)

	return err
}

func (a authRedisAdapter) Remove(token string) error {
	err := a.store.Remove(token)

	return err
}
