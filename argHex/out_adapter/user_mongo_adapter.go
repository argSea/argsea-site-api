package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
	"github.com/argSea/portfolio_blog_api/argHex/stores"
)

type userMongoAdapter struct {
	store *stores.Mordor
}

func NewUserMongoAdapter(store *stores.Mordor) out_port.UserRepo {
	u := userMongoAdapter{
		store: store,
	}

	return u
}

func (u userMongoAdapter) GetAll(limit int64, offset int64, sort interface{}) domain.Users {
	var users domain.Users
	count, err := u.store.GetAll(limit, offset, nil, &users)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return []domain.User{}
	}

	fmt.Printf("count: %v\n", count)

	return users
}

func (u userMongoAdapter) Get(id string) domain.User {
	var user domain.User
	err := u.store.Get("_id", id, &user)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.User{}
	}

	return user
}

func (u userMongoAdapter) GetByUserName(username string) domain.User {
	var user domain.User
	err := u.store.Get("userName", username, &user)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.User{}
	}

	return user
}

func (u userMongoAdapter) Set(user domain.User) error {
	key := user.Id
	user.Id = "" //unset so mongo doesn't try to set it

	err := u.store.Update(key, user)

	return err
}

func (u userMongoAdapter) Add(user domain.User) (string, error) {
	user.Id = "" //make sure it wasn't set

	new_id, err := u.store.Write(user)

	return new_id, err
}

func (u userMongoAdapter) Remove(user domain.User) error {
	user_id := user.Id

	err := u.store.Delete(user_id)

	return err
}
