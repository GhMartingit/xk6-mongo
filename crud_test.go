package xk6_mongo

import (
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestCRUDOperations(t *testing.T) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		t.Skip("MONGODB_URI not set")
	}

	client := new(Mongo).NewClient(uri)
	if client == nil {
		t.Fatalf("failed to create client")
	}
	defer client.Disconnect()

	db := "crudtestdb"
	col := "crudtestcol"
	filter := bson.M{"_id": bson.M{"$eq": "crud-1"}}

	if err := client.Insert(db, col, bson.M{"_id": "crud-1", "name": "init"}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	doc, err := client.FindOne(db, col, filter)
	if err != nil {
		t.Fatalf("find after insert: %v", err)
	}
	if doc["name"] != "init" {
		t.Fatalf("unexpected name %v", doc["name"])
	}

	update := bson.M{"name": "updated"}
	if err := client.UpdateOne(db, col, filter, update); err != nil {
		t.Fatalf("update: %v", err)
	}

	doc, err = client.FindOne(db, col, filter)
	if err != nil {
		t.Fatalf("find after update: %v", err)
	}
	if doc["name"] != "updated" {
		t.Fatalf("unexpected name after update %v", doc["name"])
	}

	if err := client.DeleteOne(db, col, filter); err != nil {
		t.Fatalf("delete: %v", err)
	}

	count, err := client.CountDocuments(db, col, filter)
	if err != nil {
		t.Fatalf("count after delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 documents, got %d", count)
	}
}
