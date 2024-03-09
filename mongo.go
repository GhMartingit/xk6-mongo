package xk6_mongo

import (
	"context"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.k6.io/k6/js/modules"
)

// Register the extension on module initialization, available to
// import from JS as "k6/x/mongo".
func init() {
	modules.Register("k6/x/mongo", new(Mongo))
}

// Mongo is the k6 extension for a Mongo client.
type Mongo struct{}

// Client is the Mongo client wrapper.
type Client struct {
	client *mongo.Client
}

// NewClient represents the Client constructor (i.e. `new mongo.Client()`) and
// returns a new Mongo client object.
// The `connURI` parameter in the `NewClient` function is used to specify the connection URI for the
// MongoDB client. It typically follows the format
// `mongodb://username:password@address:port/db?connect=direct`. This URI contains information such
// as the username, password, address, port, database name, and connection options.
// connURI -> mongodb://username:password@address:port/db?connect=direct
// connURI -> mongodb+srv://username:password@address:port/db?authSource=admin
func (*Mongo) NewClient(connURI string) interface{} {

	clientOptions := options.Client().ApplyURI(connURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Printf("Error on Connection")
		log.Printf("%+v", err)
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

func (c *Client) Find(database string, collection string, filter interface{}) []bson.M {
	db := c.client.Database(database)
	col := db.Collection(collection)
	log.Print(filter_is, filter)
	cur, err := col.Find(context.TODO(), filter)
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
	if err == mongo.ErrNoDocuments {
		log.Printf("No document was found for filter %v", filter)
		return nil
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("found document %v", result)
	return nil
}

func (c *Client) UpdateOne(database string, collection string, filter interface{}, data map[string]string) error {
	// var result bson.M
	db := c.client.Database(database)
	col := db.Collection(collection)
	update := bson.D{{"$set", data}}
	result, err := col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}

	// opts := options.FindOne().SetSort(bson.D{{"_id", 1}})
	// err = col.FindOne(context.TODO(), filter, opts).Decode(&result)
	// if err == mongo.ErrNoDocuments {
	// 	log.Printf("No document was found for filter %v", filter)
	// 	return nil
	// }
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

func (*Mongo) MongoEncode(input string) string {
	// https://www.mongodb.com/docs/manual/reference/connection-string/
	r := strings.NewReplacer("%", "%25", "$", "%24", ":", "%3A", "/", "%2F", "?", "%3F", "#", "%23", "[", "%5B", "]", "%5D", "@", "%40")
	res := r.Replace(input)
	return res
}
