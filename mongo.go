package xk6_mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"

	k6modules "go.k6.io/k6/js/modules"
)

// Register the extension on module initialization, available to
// import from JS as "k6/x/mongo".
func init() {
	k6modules.Register("k6/x/mongo", new(Mongo))
}

// Mongo is the k6 extension for a Mongo client.
type Mongo struct{}

// Client is the Mongo client wrapper.
type Client struct {
	client *mongo.Client
}

type UpsertOneModel struct {
	Query  interface{} `json:"query"`
	Update interface{} `json:"update"`
}

// NewClient represents the Client constructor (i.e. `new mongo.Client()`) and
// returns a new Mongo client object.
// connURI -> mongodb://username:password@address:port/db?connect=direct
func (*Mongo) NewClient(connURI string) interface{} {
	log.Print("start creating new client")

	clientOptions := options.Client().ApplyURI(connURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Print(err)
		return err
	}

	log.Print("created new client")
	return &Client{client: client}
}

const filter_is string = "filter is "

func (c *Client) Insert(database string, collection string, doc map[string]string) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	_, err := col.InsertOne(context.TODO(), doc)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) InsertMany(database string, collection string, docs []any) error {
	log.Printf("Insert multiple documents")
	db := c.client.Database(database)
	col := db.Collection(collection)
	_, err := col.InsertMany(context.TODO(), docs)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Upsert(database string, collection string, filter interface{}, upsert interface{}) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	opts := options.Update().SetUpsert(true)
	_, err := col.UpdateOne(context.TODO(), filter, upsert, opts)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) BulkUpsert(database string, collection string, upserts []UpsertOneModel) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	var models []mongo.WriteModel
	for _, upsert := range upserts {
		model := mongo.NewUpdateOneModel()
		model.SetFilter(upsert.Query)
		model.SetUpdate(upsert.Update)
		model.SetUpsert(true)
		models = append(models, model)
	}
	_, err := col.BulkWrite(context.TODO(), models)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (c *Client) Find(database string, collection string, filter interface{}, sort interface{}, limit int64) []bson.M {
	db := c.client.Database(database)
	col := db.Collection(collection)
	opts := options.Find().SetSort(sort).SetLimit(limit)
	log.Print(filter_is, filter)
	cur, err := col.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Fatal(err)
	}
	var results []bson.M
	if err = cur.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	return results
}

func (c *Client) Aggregate(database string, collection string, pipeline interface{}) []bson.M {
	db := c.client.Database(database)
	col := db.Collection(collection)
	log.Print(filter_is, pipeline)
	cur, err := col.Aggregate(context.TODO(), pipeline)
	if err != nil {
		log.Fatal(err)
	}
	var results []bson.M
	if err = cur.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	return results
}

func (c *Client) FindOne(database string, collection string, filter map[string]string) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	var result bson.M
	opts := options.FindOne().SetSort(bson.D{{"_id", 1}})
	log.Print(filter_is, filter)
	err := col.FindOne(context.TODO(), filter, opts).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("found document %v", result)
	return nil
}

func (c *Client) FindAll(database string, collection string) []bson.M {
	log.Printf("Find all documents")
	db := c.client.Database(database)
	col := db.Collection(collection)
	cur, err := col.Find(context.TODO(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}
	var results []bson.M
	if err = cur.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	return results
}

func (c *Client) DeleteOne(database string, collection string, filter map[string]string) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	opts := options.Delete().SetHint(bson.D{{"_id", 1}})
	log.Print(filter_is, filter)
	result, err := col.DeleteOne(context.TODO(), filter, opts)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Deleted documents %v", result)
	return nil
}

func (c *Client) DeleteMany(database string, collection string, filter map[string]string) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	opts := options.Delete().SetHint(bson.D{{"_id", 1}})
	log.Print(filter_is, filter)
	result, err := col.DeleteMany(context.TODO(), filter, opts)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Deleted documents %v", result)
	return nil
}

func (c *Client) Distinct(database string, collection string, field string, filter interface{}) []interface{} {
	db := c.client.Database(database)
	col := db.Collection(collection)
	results, err := col.Distinct(context.TODO(), field, filter)
	if err != nil {
		log.Fatal(err)
	}

	return results
}

func (c *Client) DropCollection(database string, collection string) error {
	log.Printf("Delete collection if present")
	db := c.client.Database(database)
	col := db.Collection(collection)
	err := col.Drop(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (c *Client) Disconnect() {
	log.Printf("Disconnecting from Mongo database")
	err := c.client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}
