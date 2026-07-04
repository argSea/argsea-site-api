package stores

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	DB     *mongo.Database
	Client *mongo.Client
}

func NewMongoStore(user string, pass string, host string, db_name string) (*MongoDB, error) {
	mongoDB := new(MongoDB)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second+10)
	defer cancel()

	mongoDB.Client, _ = mongo.NewClient(options.Client().ApplyURI("mongodb://" + user + ":" + pass + "@" + host + "/?authSource=admin&readPreference=primary&ssl=false"))

	clientErr := mongoDB.Client.Connect(ctx)
	mongoDB.DB = mongoDB.Client.Database(db_name)

	return mongoDB, clientErr
}
