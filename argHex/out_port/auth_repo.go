package out_port

import "time"

type AuthRepo interface {
	Store(token string, expires time.Duration, data interface{}) error
	Get(id string) string
	Remove(id string) error
}
