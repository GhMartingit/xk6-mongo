package xk6_mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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

type Options struct {
	Limit int64 `json:"limit"`
	Skip int64 `json:"skip"`
	Sort interface{} `json:"sort"`
}

func createOptions(opts map[string]any) Options {
	jsonStr, err := json.Marshal(opts)
    if err != nil {
        fmt.Println(err)
    }

    // Convert json string to struct
    var opt Options
    if err := json.Unmarshal(jsonStr, &opt); err != nil {
        log.Println(err)
    }
	return opt
}

func toFindOptions(opts Options) *options.FindOptions {
	opt := options.Find()

	if opts.Limit > 0 {
		opt.SetLimit(opts.Limit)
	}
	if opts.Skip > 0 {
		opt.SetSkip(opts.Skip)
	}

	if opts.Sort != nil  {
		opt.SetSort(opts.Sort)
	}
	return opt
}

func toCountOptions(opts Options) *options.CountOptions {
	opt := options.Count()
	if opts.Limit > 0 {
		opt.SetLimit(opts.Limit)
	}
	if opts.Skip > 0 {
		opt.SetSkip(opts.Skip)
	}
	return opt
}

// NewClient represents the Client constructor (i.e. `new mongo.Client()`) and
// returns a new Mongo client object.
// connURI -> mongodb://username:password@address:port/db?connect=direct
func (*Mongo) NewClient(connURI string) interface{} {

	clientOptions := options.Client().ApplyURI(connURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}

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

func (c *Client) Find(database string, collection string, filter interface{}, opts map[string]any) []bson.M {
	db := c.client.Database(database)
	col := db.Collection(collection)
	log.Print(filter_is, filter)

	opt := toFindOptions(createOptions(opts))
	log.Printf("%+v", opt)

	cur, err := col.Find(context.TODO(), filter, opt)
	if err != nil {
		log.Fatal(err)
	}
	var results []bson.M
	if err = cur.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	return results
}

func (c *Client) CountDocuments(database string, collection string, filter interface{}, opts map[string]any) int64 {
	db := c.client.Database(database)
	col := db.Collection(collection)
	log.Print(filter_is, filter)

	opt := toCountOptions(createOptions(opts))
	log.Printf("%+v", opt)

	cur, err := col.CountDocuments(context.TODO(), filter, opt)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		panic(err)
	}
	return cur
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