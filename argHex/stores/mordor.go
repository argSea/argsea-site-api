package stores

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mordor struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewMordor(col *mongo.Collection, ctx context.Context) *Mordor {
	return &Mordor{
		collection: col,
		ctx:        ctx,
	}
}

func (m *Mordor) Get(field string, value interface{}, decoder interface{}) error {
	log.Printf("field: %v; value %v; decoder: %+v\n", field, value, decoder)
	if nil == m.collection {
		return errors.New("connection not setup")
	}

	if "key" == field {
		field = "_id"
	}

	if "_id" == field {
		id, idErr := primitive.ObjectIDFromHex(value.(string))

		if nil != idErr {
			return errors.New("invalid key")
		}

		value = id
	}

	err := m.collection.FindOne(m.ctx, bson.M{field: value}).Decode(decoder)

	return err
}

func (m *Mordor) GetMany(field string, value interface{}, limit int64, offset int64, sort interface{}, decoder interface{}) (int64, error) {
	log.Printf("field: %v; value %v; limit %v; offset %v; sort %+v; decoder: %+v\n", field, value, limit, offset, sort, decoder)
	if nil == m.collection {
		return 0, errors.New("connection not setup")
	}

	count, cErr := m.collection.EstimatedDocumentCount(m.ctx, nil)

	if nil != cErr {
		return 0, cErr
	}

	findOpts := options.Find()
	findOpts.SetLimit(limit)
	findOpts.SetSkip(offset)
	findOpts.SetSort(sort)
	cursor, err := m.collection.Find(m.ctx, bson.M{field: value}, findOpts)

	log.Printf("cursor: %+v\n", cursor)
	log.Printf("err: %+v\n", err)

	if nil != err {
		return 0, err
	}

	cursor.All(m.ctx, decoder)

	return count, nil
}

func (m *Mordor) GetAll(limit int64, offset int64, sort interface{}, decoder interface{}) (int64, error) {
	log.Printf("limit: %v; offset %v; sort: %+v; decoder: %+v;\n", limit, offset, sort, decoder)
	if nil == m.collection {
		return 0, errors.New("connection not setup")
	}

	count, cErr := m.collection.EstimatedDocumentCount(m.ctx, nil)

	if nil != cErr {
		return 0, cErr
	}

	findOpts := options.Find()
	findOpts.SetLimit(limit)
	findOpts.SetSkip(offset)
	findOpts.SetSort(sort)
	cursor, err := m.collection.Find(m.ctx, bson.D{}, findOpts)

	log.Printf("cursor: %+v\n", cursor)
	log.Printf("err: %+v\n", err)

	if nil != err {
		return 0, err
	}

	cursor.All(m.ctx, decoder)

	return count, nil
}

func (m *Mordor) Write(data interface{}) (string, error) {
	log.Printf("value %+v;\n", data)

	if nil == m.collection {
		return "", errors.New("connection not setup")
	}

	result, err := m.collection.InsertOne(m.ctx, data)

	if nil != err {
		return "", err
	}

	id, ok := result.InsertedID.(primitive.ObjectID)

	if !ok {
		return "", errors.New("unable to parse InsertID")
	}

	return id.Hex(), nil
}

func (m *Mordor) Update(key string, newData interface{}) error {
	log.Printf("key: %v; new_data %+v\n", key, newData)
	if nil == m.collection {
		return errors.New("connection not setup")
	}

	id, idErr := primitive.ObjectIDFromHex(key)

	if nil != idErr {
		return errors.New("invalid key")
	}

	_, err := m.collection.UpdateOne(
		m.ctx,
		bson.M{"_id": id},
		bson.D{
			{Key: "$set", Value: newData},
		},
	)

	if nil != err {
		return err
	}

	return nil
}

func (m *Mordor) Delete(key string) error {
	log.Printf("key: %v;\n", key)
	if nil == m.collection {
		return errors.New("connection not setup")
	}

	id, idErr := primitive.ObjectIDFromHex(key)

	if nil != idErr {
		return errors.New("invalid key")
	}

	_, err := m.collection.DeleteOne(m.ctx, bson.M{"_id": id})

	if nil != err {
		return err
	}

	return nil
}
