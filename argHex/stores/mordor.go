package stores

import (
	"context"
	"errors"

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
	if nil == m.collection {
		return 0, errors.New("connection not setup")
	}

	count, cErr := m.collection.EstimatedDocumentCount(m.ctx, nil)

	if nil != cErr {
		// count is advisory
		count = 0
	}

	findOpts := options.Find()
	findOpts.SetLimit(limit)
	findOpts.SetSkip(offset)
	findOpts.SetSort(sort)
	cursor, err := m.collection.Find(m.ctx, bson.M{field: value}, findOpts)

	if nil != err {
		return 0, err
	}

	cursor.All(m.ctx, decoder)

	return count, nil
}

func (m *Mordor) GetAll(limit int64, offset int64, sort interface{}, decoder interface{}) (int64, error) {
	if nil == m.collection {
		return 0, errors.New("connection not setup")
	}

	count, cErr := m.collection.EstimatedDocumentCount(m.ctx, nil)

	if nil != cErr {
		// count is advisory
		count = 0
	}

	findOpts := options.Find()
	findOpts.SetLimit(limit)
	findOpts.SetSkip(offset)
	findOpts.SetSort(sort)
	cursor, err := m.collection.Find(m.ctx, bson.D{}, findOpts)

	if nil != err {
		return 0, err
	}

	cursor.All(m.ctx, decoder)

	return count, nil
}

// Find runs a caller-built filter with the same limit/offset/sort contract the
// other reads share. It exists for reads whose match is richer than GetMany's
// single-field equality, like a date-range window. A limit of 0 means no limit.
func (m *Mordor) Find(filter interface{}, limit int64, offset int64, sort interface{}, decoder interface{}) (int64, error) {
	if nil == m.collection {
		return 0, errors.New("connection not setup")
	}

	count, cErr := m.collection.EstimatedDocumentCount(m.ctx, nil)

	if nil != cErr {
		// count is advisory
		count = 0
	}

	findOpts := options.Find()
	findOpts.SetLimit(limit)
	findOpts.SetSkip(offset)
	findOpts.SetSort(sort)
	cursor, err := m.collection.Find(m.ctx, filter, findOpts)

	if nil != err {
		return 0, err
	}

	cursor.All(m.ctx, decoder)

	return count, nil
}

// CreateIndexes lands the given indexes on the collection. Mongo skips any that
// already exist, so it is safe to call on every boot to keep TTL and lookup
// indexes in place.
func (m *Mordor) CreateIndexes(models []mongo.IndexModel) error {
	if nil == m.collection {
		return errors.New("connection not setup")
	}

	_, err := m.collection.Indexes().CreateMany(m.ctx, models)

	return err
}

func (m *Mordor) Write(data interface{}) (string, error) {
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

// Replace swaps the stored document for newData wholesale (ReplaceOne, not
// $set). Content updates go through this so a field the caller cleared is
// actually cleared; $set over bson-omitempty fields silently merges instead.
// This exactly matches the snapshot/restore model: what you write is what the
// document becomes.
func (m *Mordor) Replace(key string, newData interface{}) error {
	if nil == m.collection {
		return errors.New("connection not setup")
	}

	id, idErr := primitive.ObjectIDFromHex(key)

	if nil != idErr {
		return errors.New("invalid key")
	}

	_, err := m.collection.ReplaceOne(m.ctx, bson.M{"_id": id}, newData)

	return err
}

// UpdateMany applies a $set to every document matching filter. Used to clear
// the "current" flag across an entity's other revisions after a new current
// revision lands.
func (m *Mordor) UpdateMany(filter interface{}, set interface{}) error {
	if nil == m.collection {
		return errors.New("connection not setup")
	}

	_, err := m.collection.UpdateMany(
		m.ctx,
		filter,
		bson.D{
			{Key: "$set", Value: set},
		},
	)

	return err
}

func (m *Mordor) Delete(key string) error {
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
