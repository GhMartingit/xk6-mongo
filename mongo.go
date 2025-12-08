package xk6_mongo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

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
	client         *mongo.Client
	defaultTimeout time.Duration
	retryWrites    bool
	retryReads     bool
}

type UpsertOneModel struct {
	Query  any `json:"query"`
	Update any `json:"update"`
}

const (
	defaultConnectionTimeout = 10 * time.Second
	defaultOperationTimeout  = 30 * time.Second
)

// NewClient represents the Client constructor (i.e. `new mongo.Client()`) and
// returns a new Mongo client object.
// connURI -> mongodb://username:password@address:port/db?connect=direct
func (m *Mongo) NewClient(connURI string) *Client {
	return m.NewClientWithOptions(connURI, nil)
}

func (*Mongo) NewClientWithOptions(connURI string, opts any) *Client {
	log.Print("start creating new client")

	if connURI == "" {
		log.Printf("Error: connection URI cannot be empty")
		return nil
	}

	clientOptions, err := prepareClientOptions(connURI, opts)
	if err != nil {
		log.Printf("Error while preparing client options: %v", err)
		return nil
	}

	// Create context with timeout for connection
	ctx, cancel := context.WithTimeout(context.Background(), defaultConnectionTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Printf("Error while establishing a connection to MongoDB: %v", err)
		return nil
	}

	// Verify connection with ping
	if err := client.Ping(ctx, nil); err != nil {
		log.Printf("Error while pinging MongoDB: %v", err)
		// Attempt to disconnect on ping failure
		_ = client.Disconnect(context.Background())
		return nil
	}

	log.Print("created new client and verified connection")

	// Enable retry writes and reads by default (can be overridden in client options)
	retryWrites := true
	retryReads := true

	return &Client{
		client:         client,
		defaultTimeout: defaultOperationTimeout,
		retryWrites:    retryWrites,
		retryReads:     retryReads,
	}
}

// getContext creates a context with the default timeout
func (c *Client) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.defaultTimeout)
}

// getCollection returns a collection and validates input
func (c *Client) getCollection(database, collection string) (*mongo.Collection, error) {
	if err := validateDatabaseAndCollection(database, collection); err != nil {
		return nil, err
	}
	return c.client.Database(database).Collection(collection), nil
}

// validateDatabaseAndCollection validates database and collection names
func validateDatabaseAndCollection(database, collection string) error {
	if database == "" {
		return errors.New("database name cannot be empty")
	}
	if collection == "" {
		return errors.New("collection name cannot be empty")
	}
	// MongoDB database names cannot contain certain characters
	invalidChars := []string{"/", "\\", ".", "\"", "$", " ", "\x00"}
	for _, char := range invalidChars {
		if strings.Contains(database, char) {
			return fmt.Errorf("database name contains invalid character: %s", char)
		}
	}
	return nil
}

func (c *Client) Insert(database string, collection string, doc any) error {
	if doc == nil {
		return errDocumentNil
	}

	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	_, err = col.InsertOne(ctx, doc)
	if err != nil {
		log.Printf(errInsertingDocument, err)
		return err
	}
	log.Print("Document inserted successfully")
	return nil
}

func (c *Client) InsertMany(database string, collection string, docs []any) error {
	if len(docs) == 0 {
		return errDocsEmpty
	}

	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		log.Printf(errInsertingDocuments, err)
		return err
	}
	return nil
}

func (c *Client) Upsert(database string, collection string, filter any, upsert any) error {
	if filter == nil {
		return errFilterNil
	}

	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return err
	}

	opts := options.Update().SetUpsert(true)

	updateDoc, err := prepareUpdateDocument(upsert)
	if err != nil {
		log.Printf(errPreparingUpsertDoc, err)
		return err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	_, err = col.UpdateOne(ctx, filter, updateDoc, opts)
	if err != nil {
		log.Printf(errPerformingUpsert, err)
		return err
	}
	return nil
}

const (
	errDecodingDocuments     = "Error while decoding documents: %v"
	errValidatingCollection  = "Error validating collection: %v"
	errInsertingDocument     = "Error while inserting document: %v"
	errInsertingDocuments    = "Error while inserting multiple documents: %v"
	errFindingDocuments      = "Error while finding documents: %v"
	errFindingDocument       = "Error while finding the document: %v"
	errUpdatingDocument      = "Error while updating the document: %v"
	errUpdatingDocuments     = "Error while updating the documents: %v"
	errDeletingDocument      = "Error while deleting the document: %v"
	errDeletingDocuments     = "Error while deleting the documents: %v"
	errPerformingUpsert      = "Error while performing upsert: %v"
	errPreparingUpdateDoc    = "Error while preparing update document: %v"
	errPreparingUpsertDoc    = "Error while preparing upsert document: %v"
	errAggregating           = "Error while aggregating: %v"
	errGettingDistinctValues = "Error while getting distinct values: %v"
	errDroppingCollection    = "Error while dropping the collection: %v"
	errCountingDocuments     = "Error while counting documents: %v"
	errFindingAndUpdating    = "Error while finding and updating document: %v"
)

var (
	errFilterNil   = errors.New("filter cannot be nil")
	errDocumentNil = errors.New("document cannot be nil")
	errPipelineNil = errors.New("pipeline cannot be nil")
	errDocsEmpty   = errors.New("documents array cannot be empty")
	errLimitNeg    = errors.New("limit cannot be negative")
)

func (c *Client) Find(database string, collection string, filter any, sort any, limit int64) ([]bson.M, error) {
	if limit < 0 {
		return nil, errLimitNeg
	}

	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return nil, err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	opts := options.Find().SetSort(sort).SetLimit(limit)
	cur, err := col.Find(ctx, filter, opts)
	if err != nil {
		log.Printf(errFindingDocuments, err)
		return nil, err
	}
	defer cur.Close(ctx)

	var results []bson.M
	if err = cur.All(ctx, &results); err != nil {
		log.Printf(errDecodingDocuments, err)
		return nil, err
	}
	return results, nil
}

// FindWithOptions provides advanced find options including batch size control
func (c *Client) FindWithOptions(database string, collection string, filter any, findOptions map[string]any) ([]bson.M, error) {
	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return nil, err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	opts := options.Find()

	// Apply options from map
	if limit, ok := findOptions["limit"].(int64); ok && limit > 0 {
		opts.SetLimit(limit)
	}
	if skip, ok := findOptions["skip"].(int64); ok && skip > 0 {
		opts.SetSkip(skip)
	}
	if sort, ok := findOptions["sort"]; ok {
		opts.SetSort(sort)
	}
	if batchSize, ok := findOptions["batch_size"].(int32); ok && batchSize > 0 {
		opts.SetBatchSize(batchSize)
	}
	if projection, ok := findOptions["projection"]; ok {
		opts.SetProjection(projection)
	}

	cur, err := col.Find(ctx, filter, opts)
	if err != nil {
		log.Printf(errFindingDocuments, err)
		return nil, err
	}
	defer cur.Close(ctx)

	var results []bson.M
	if err = cur.All(ctx, &results); err != nil {
		log.Printf(errDecodingDocuments, err)
		return nil, err
	}
	return results, nil
}

func (c *Client) Aggregate(database string, collection string, pipeline any) ([]bson.M, error) {
	if pipeline == nil {
		return nil, errPipelineNil
	}

	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return nil, err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	cur, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf(errAggregating, err)
		return nil, err
	}
	defer cur.Close(ctx)

	var results []bson.M
	if err = cur.All(ctx, &results); err != nil {
		log.Printf(errDecodingDocuments, err)
		return nil, err
	}
	return results, nil
}

func (c *Client) FindOne(database string, collection string, filter any) (bson.M, error) {
	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return nil, err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	var result bson.M
	err = col.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		log.Printf(errFindingDocument, err)
		return nil, err
	}

	return result, nil
}

func (c *Client) UpdateOne(database string, collection string, filter any, data any) error {
	if filter == nil {
		return errFilterNil
	}

	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return err
	}

	update, err := prepareUpdateDocument(data)
	if err != nil {
		log.Printf(errPreparingUpdateDoc, err)
		return err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	_, err = col.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf(errUpdatingDocument, err)
		return err
	}

	return nil
}

func (c *Client) UpdateMany(database string, collection string, filter any, data any) error {
	if filter == nil {
		return errFilterNil
	}

	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return err
	}

	update, err := prepareUpdateDocument(data)
	if err != nil {
		log.Printf(errPreparingUpdateDoc, err)
		return err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	_, err = col.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Printf(errUpdatingDocuments, err)
		return err
	}

	return nil
}

func (c *Client) FindAll(database string, collection string) ([]bson.M, error) {
	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return nil, err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	// Use an empty filter to match all documents
	cur, err := col.Find(ctx, bson.D{})
	if err != nil {
		log.Printf(errFindingDocuments, err)
		return nil, err
	}
	defer cur.Close(ctx)

	var results []bson.M
	if err = cur.All(ctx, &results); err != nil {
		log.Printf(errDecodingDocuments, err)
		return nil, err
	}

	return results, nil
}

func (c *Client) DeleteOne(database string, collection string, filter any) error {
	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	_, err = col.DeleteOne(ctx, filter)
	if err != nil {
		log.Printf(errDeletingDocument, err)
		return err
	}

	return nil
}

func (c *Client) DeleteMany(database string, collection string, filter any) error {
	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	_, err = col.DeleteMany(ctx, filter)
	if err != nil {
		log.Printf(errDeletingDocuments, err)
		return err
	}

	return nil
}

func (c *Client) Distinct(database string, collection string, field string, filter any) ([]any, error) {
	if field == "" {
		return nil, errors.New("field name cannot be empty")
	}

	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return nil, err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	result, err := col.Distinct(ctx, field, filter)
	if err != nil {
		log.Printf(errGettingDistinctValues, err)
		return nil, err
	}

	return result, nil
}

func (c *Client) DropCollection(database string, collection string) error {
	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	err = col.Drop(ctx)
	if err != nil {
		log.Printf(errDroppingCollection, err)
		return err
	}

	return nil
}

func (c *Client) CountDocuments(database string, collection string, filter any) (int64, error) {
	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return 0, err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf(errCountingDocuments, err)
		return 0, err
	}
	return count, nil
}

func (c *Client) FindOneAndUpdate(database string, collection string, filter any, update any) (bson.M, error) {
	if filter == nil {
		return nil, errFilterNil
	}

	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return nil, err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var out bson.M
	err = col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&out)
	if err != nil {
		log.Printf(errFindingAndUpdating, err)
		return nil, err
	}
	return out, nil
}

func (c *Client) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := c.client.Disconnect(ctx)
	if err != nil {
		log.Printf("Error while disconnecting from the database: %v", err)
		return err
	}

	return nil
}

// BulkWrite executes multiple write operations in a single call
func (c *Client) BulkWrite(database string, collection string, operations []mongo.WriteModel) (int64, int64, error) {
	if len(operations) == 0 {
		return 0, 0, errors.New("operations array cannot be empty")
	}

	col, err := c.getCollection(database, collection)
	if err != nil {
		log.Printf(errValidatingCollection, err)
		return 0, 0, err
	}

	ctx, cancel := c.getContext()
	defer cancel()

	result, err := col.BulkWrite(ctx, operations)
	if err != nil {
		log.Printf("Error while performing bulk write: %v", err)
		return 0, 0, err
	}

	return result.InsertedCount, result.ModifiedCount, nil
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
