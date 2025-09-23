package xk6_mongo

import (
	"context"
	"fmt"
	"log"
	"strings"

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

type UpsertOneModel struct {
	Query  any `json:"query"`
	Update any `json:"update"`
}

// NewClient represents the Client constructor (i.e. `new mongo.Client()`) and
// returns a new Mongo client object.
// connURI -> mongodb://username:password@address:port/db?connect=direct
func (m *Mongo) NewClient(connURI string) *Client {
	return m.NewClientWithOptions(connURI, nil)
}

func (*Mongo) NewClientWithOptions(connURI string, opts any) *Client {
	log.Print("start creating new client")

	clientOptions, err := prepareClientOptions(connURI, opts)
	if err != nil {
		log.Printf("Error while preparing client options: %v", err)
		return nil
	}

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Printf("Error while establishing a connection to MongoDB: %v", err)
		return nil
	}

	log.Print("created new client")
	return &Client{client: client}
}

func (c *Client) Insert(database string, collection string, doc any) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	_, err := col.InsertOne(context.Background(), doc)
	if err != nil {
		log.Printf("Error while inserting document: %v", err)
		return err
	}
	log.Print("Document inserted successfully")
	return nil
}

func (c *Client) InsertMany(database string, collection string, docs []any) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	_, err := col.InsertMany(context.Background(), docs)
	if err != nil {
		log.Printf("Error while inserting multiple documents: %v", err)
		return err
	}
	return nil
}

func (c *Client) Upsert(database string, collection string, filter any, upsert any) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	opts := options.Update().SetUpsert(true)
	_, err := col.UpdateOne(context.Background(), filter, upsert, opts)
	if err != nil {
		log.Printf("Error while performing upsert: %v", err)
		return err
	}
	return nil
}

const errDecodingDocuments = "Error while decoding documents: %v"

func (c *Client) Find(database string, collection string, filter any, sort any, limit int64) ([]bson.M, error) {
	db := c.client.Database(database)
	col := db.Collection(collection)
	opts := options.Find().SetSort(sort).SetLimit(limit)
	cur, err := col.Find(context.Background(), filter, opts)
	if err != nil {
		log.Printf("Error while finding documents: %v", err)
		return nil, err
	}
	var results []bson.M
	if err = cur.All(context.Background(), &results); err != nil {
		log.Printf(errDecodingDocuments, err)
		return nil, err
	}
	return results, nil
}

func (c *Client) Aggregate(database string, collection string, pipeline any) ([]bson.M, error) {
	db := c.client.Database(database)
	col := db.Collection(collection)
	cur, err := col.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Printf("Error while aggregating: %v", err)
		return nil, err
	}
	var results []bson.M
	if err = cur.All(context.Background(), &results); err != nil {
		log.Printf(errDecodingDocuments, err)
		return nil, err
	}
	return results, nil
}

func (c *Client) FindOne(database string, collection string, filter any) (bson.M, error) {
	db := c.client.Database(database)
	col := db.Collection(collection)
	var result bson.M
	err := col.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		log.Printf("Error while finding the document: %v", err)
		return nil, err
	}

	return result, nil
}

func (c *Client) UpdateOne(database string, collection string, filter any, data any) error {
	db := c.client.Database(database)
	col := db.Collection(collection)

	update, err := prepareUpdateDocument(data)
	if err != nil {
		log.Printf("Error while preparing update document: %v", err)
		return err
	}

	_, err = col.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Error while updating the document: %v", err)
		return err
	}

	return nil
}

func (c *Client) UpdateMany(database string, collection string, filter any, data any) error {
	db := c.client.Database(database)
	col := db.Collection(collection)

	update, err := prepareUpdateDocument(data)
	if err != nil {
		log.Printf("Error while preparing update document: %v", err)
		return err
	}

	_, err = col.UpdateMany(context.Background(), filter, update)
	if err != nil {
		log.Printf("Error while updating the documents: %v", err)
		return err
	}

	return nil
}

func (c *Client) FindAll(database string, collection string) ([]bson.M, error) {
	db := c.client.Database(database)
	col := db.Collection(collection)
	cur, err := col.Find(context.Background(), bson.D{{}})
	if err != nil {
		log.Printf("Error while finding documents: %v", err)
		return nil, err
	}

	var results []bson.M
	if err = cur.All(context.Background(), &results); err != nil {
		log.Printf(errDecodingDocuments, err)
		return nil, err
	}

	return results, nil
}

func (c *Client) DeleteOne(database string, collection string, filter any) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	_, err := col.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Error while deleting the document: %v", err)
		return err
	}

	return nil
}

func (c *Client) DeleteMany(database string, collection string, filter any) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	_, err := col.DeleteMany(context.Background(), filter)
	if err != nil {
		log.Printf("Error while deleting the documents: %v", err)
		return err
	}

	return nil
}

func (c *Client) Distinct(database string, collection string, field string, filter any) ([]any, error) {
	db := c.client.Database(database)
	col := db.Collection(collection)
	result, err := col.Distinct(context.Background(), field, filter)
	if err != nil {
		log.Printf("Error while getting distinct values: %v", err)
		return nil, err
	}

	return result, nil
}

func (c *Client) DropCollection(database string, collection string) error {
	db := c.client.Database(database)
	col := db.Collection(collection)
	err := col.Drop(context.Background())
	if err != nil {
		log.Printf("Error while dropping the collection: %v", err)
		return err
	}

	return nil
}

func (c *Client) CountDocuments(database string, collection string, filter any) (int64, error) {
	db := c.client.Database(database)
	col := db.Collection(collection)
	count, err := col.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Printf("Error while counting documents: %v", err)
		return 0, err
	}
	return count, nil
}

func (c *Client) FindOneAndUpdate(database string, collection string, filter any, update any) (*mongo.SingleResult, error) {
	db := c.client.Database(database)
	col := db.Collection(collection)
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := col.FindOneAndUpdate(context.Background(), filter, update, opts)
	if result.Err() != nil {
		log.Printf("Error while finding and updating document: %v", result.Err())
		return nil, result.Err()
	}
	return result, nil
}

func (c *Client) Disconnect() error {
	err := c.client.Disconnect(context.Background())
	if err != nil {
		log.Printf("Error while disconnecting from the database: %v", err)
		return err
	}

	return nil
}

func prepareClientOptions(connURI string, opts any) (*options.ClientOptions, error) {
	switch v := opts.(type) {
	case nil:
		return options.Client().ApplyURI(connURI), nil
	case *options.ClientOptions:
		if v == nil {
			return nil, fmt.Errorf("client options cannot be nil")
		}
		v.ApplyURI(connURI)
		return v, nil
	case map[string]any:
		return clientOptionsFromMap(connURI, v)
	case bson.M:
		return clientOptionsFromMap(connURI, map[string]any(v))
	default:
		return nil, fmt.Errorf("unsupported client options type %T", opts)
	}
}

func clientOptionsFromMap(connURI string, raw map[string]any) (*options.ClientOptions, error) {
	normalized := normalizeKeys(raw)
	clientOptions := options.Client().ApplyURI(connURI)

	bsonBytes, err := bson.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal client options: %w", err)
	}

	if err := bson.Unmarshal(bsonBytes, clientOptions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client options: %w", err)
	}

	return clientOptions, nil
}

func normalizeKeys(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, val := range v {
			out[toPascalCase(key)] = normalizeKeys(val)
		}
		return out
	case bson.M:
		out := make(map[string]any, len(v))
		for key, val := range v {
			out[toPascalCase(key)] = normalizeKeys(val)
		}
		return out
	case []any:
		for i, val := range v {
			v[i] = normalizeKeys(val)
		}
		return v
	default:
		return value
	}
}

var knownAcronyms = map[string]string{
	"api":  "API",
	"id":   "ID",
	"uri":  "URI",
	"tls":  "TLS",
	"ocsp": "OCSP",
}

func toPascalCase(input string) string {
	if input == "" {
		return ""
	}
	segments := strings.FieldsFunc(input, func(r rune) bool {
		switch r {
		case '_', '-', ' ':
			return true
		default:
			return false
		}
	})
	for i, segment := range segments {
		lowerSegment := strings.ToLower(segment)
		if replacement, ok := knownAcronyms[lowerSegment]; ok {
			segments[i] = replacement
			continue
		}
		if len(segment) == 0 {
			continue
		}
		runes := []rune(strings.ToLower(segment))
		runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
		segments[i] = string(runes)
	}
	return strings.Join(segments, "")
}

func prepareUpdateDocument(data any) (any, error) {
	if data == nil {
		return nil, fmt.Errorf("update document cannot be nil")
	}

	if isPipelineUpdate(data) || updateDocumentHasOperator(data) {
		return data, nil
	}

	switch doc := data.(type) {
	case bson.D:
		return bson.D{{Key: "$set", Value: doc}}, nil
	case bson.M:
		return bson.M{"$set": doc}, nil
	case map[string]any:
		return bson.M{"$set": doc}, nil
	case bson.A:
		return bson.M{"$set": doc}, nil
	default:
		bytes, err := bson.Marshal(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal update document: %w", err)
		}
		var generic bson.M
		if err := bson.Unmarshal(bytes, &generic); err != nil {
			return nil, fmt.Errorf("failed to unmarshal update document: %w", err)
		}
		return bson.M{"$set": generic}, nil
	}
}

func updateDocumentHasOperator(doc any) bool {
	switch value := doc.(type) {
	case bson.D:
		return hasOperatorInKeysD(value)
	case bson.M:
		return hasOperatorInKeysM(value)
	case map[string]any:
		return hasOperatorInKeysMap(value)
	}
	return false
}

func hasOperatorInKeysD(d bson.D) bool {
	for _, elem := range d {
		if strings.HasPrefix(elem.Key, "$") {
			return true
		}
	}
	return false
}

func hasOperatorInKeysM(m bson.M) bool {
	for key := range m {
		if strings.HasPrefix(key, "$") {
			return true
		}
	}
	return false
}

func hasOperatorInKeysMap(m map[string]any) bool {
	for key := range m {
		if strings.HasPrefix(key, "$") {
			return true
		}
	}
	return false
}

func isPipelineUpdate(doc any) bool {
	switch doc.(type) {
	case []bson.D, []bson.M, []any, bson.A:
		return true
	}
	return false
}
