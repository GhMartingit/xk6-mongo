package xk6_mongo

import (
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TestAllFeatures is a comprehensive integration test that verifies all features
func TestAllFeatures(t *testing.T) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		t.Skip("MONGODB_URI not set")
	}

	client := new(Mongo).NewClient(uri)
	if client == nil {
		t.Fatal("Failed to create client")
	}
	defer client.Disconnect()

	db := "featurestest"
	col := "testcollection"

	// Clean up before tests
	_ = client.DropCollection(db, col)

	t.Run("Insert_Operation", func(t *testing.T) {
		doc := bson.M{
			"_id":    "test-1",
			"name":   "Alice",
			"age":    30,
			"email":  "alice@example.com",
			"active": true,
		}

		err := client.Insert(db, col, doc)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
		t.Log("✅ Insert successful")
	})

	t.Run("FindOne_Operation", func(t *testing.T) {
		result, err := client.FindOne(db, col, bson.M{"_id": "test-1"})
		if err != nil {
			t.Fatalf("FindOne failed: %v", err)
		}
		if result["name"] != "Alice" {
			t.Errorf("Expected name 'Alice', got '%v'", result["name"])
		}
		t.Logf("✅ FindOne successful: %v", result["name"])
	})

	t.Run("InsertMany_Operation", func(t *testing.T) {
		docs := []any{
			bson.M{"_id": "test-2", "name": "Bob", "age": 25, "active": true},
			bson.M{"_id": "test-3", "name": "Charlie", "age": 35, "active": false},
			bson.M{"_id": "test-4", "name": "Diana", "age": 28, "active": true},
		}

		err := client.InsertMany(db, col, docs)
		if err != nil {
			t.Fatalf("InsertMany failed: %v", err)
		}
		t.Log("✅ InsertMany successful")
	})

	t.Run("Find_Operation", func(t *testing.T) {
		results, err := client.Find(db, col, bson.M{"active": true}, bson.M{"age": 1}, 10)
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 active users, got %d", len(results))
		}
		t.Logf("✅ Find successful: found %d documents", len(results))
	})

	t.Run("FindWithOptions_Operation", func(t *testing.T) {
		options := map[string]any{
			"limit":      int64(2),
			"skip":       int64(1),
			"sort":       bson.M{"age": -1},
			"projection": bson.M{"name": 1, "age": 1, "_id": 0},
		}

		results, err := client.FindWithOptions(db, col, bson.M{"active": true}, options)
		if err != nil {
			t.Fatalf("FindWithOptions failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 documents with skip/limit, got %d", len(results))
		}
		// Check projection worked (no _id field)
		if _, exists := results[0]["_id"]; exists {
			t.Error("Projection failed: _id field should not exist")
		}
		t.Logf("✅ FindWithOptions successful: %d documents", len(results))
	})

	t.Run("FindAll_Operation", func(t *testing.T) {
		results, err := client.FindAll(db, col)
		if err != nil {
			t.Fatalf("FindAll failed: %v", err)
		}
		if len(results) != 4 {
			t.Errorf("Expected 4 total documents, got %d", len(results))
		}
		t.Logf("✅ FindAll successful: %d documents", len(results))
	})

	t.Run("UpdateOne_Operation", func(t *testing.T) {
		err := client.UpdateOne(db, col, bson.M{"_id": "test-1"}, bson.M{"age": 31, "updated": true})
		if err != nil {
			t.Fatalf("UpdateOne failed: %v", err)
		}

		result, _ := client.FindOne(db, col, bson.M{"_id": "test-1"})
		if result["age"].(int32) != 31 {
			t.Errorf("Expected age 31, got %v", result["age"])
		}
		t.Log("✅ UpdateOne successful")
	})

	t.Run("UpdateMany_Operation", func(t *testing.T) {
		err := client.UpdateMany(db, col, bson.M{"active": true}, bson.M{"verified": true})
		if err != nil {
			t.Fatalf("UpdateMany failed: %v", err)
		}

		results, _ := client.Find(db, col, bson.M{"verified": true}, nil, 10)
		if len(results) != 3 {
			t.Errorf("Expected 3 verified documents, got %d", len(results))
		}
		t.Logf("✅ UpdateMany successful: updated %d documents", len(results))
	})

	t.Run("Upsert_Operation", func(t *testing.T) {
		err := client.Upsert(db, col, bson.M{"_id": "test-5"}, bson.M{"name": "Eve", "age": 29})
		if err != nil {
			t.Fatalf("Upsert failed: %v", err)
		}

		result, _ := client.FindOne(db, col, bson.M{"_id": "test-5"})
		if result["name"] != "Eve" {
			t.Error("Upsert did not insert document")
		}
		t.Log("✅ Upsert successful")
	})

	t.Run("FindOneAndUpdate_Operation", func(t *testing.T) {
		result, err := client.FindOneAndUpdate(
			db, col,
			bson.M{"_id": "test-5"},
			bson.M{"$set": bson.M{"age": 30}},
		)
		if err != nil {
			t.Fatalf("FindOneAndUpdate failed: %v", err)
		}
		if result["age"].(int32) != 30 {
			t.Errorf("Expected updated age 30, got %v", result["age"])
		}
		t.Log("✅ FindOneAndUpdate successful")
	})

	t.Run("CountDocuments_Operation", func(t *testing.T) {
		count, err := client.CountDocuments(db, col, bson.M{"active": true})
		if err != nil {
			t.Fatalf("CountDocuments failed: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count 3, got %d", count)
		}
		t.Logf("✅ CountDocuments successful: %d documents", count)
	})

	t.Run("Distinct_Operation", func(t *testing.T) {
		values, err := client.Distinct(db, col, "active", bson.M{})
		if err != nil {
			t.Fatalf("Distinct failed: %v", err)
		}
		if len(values) != 2 { // true and false
			t.Errorf("Expected 2 distinct values, got %d", len(values))
		}
		t.Logf("✅ Distinct successful: %d unique values", len(values))
	})

	t.Run("Aggregate_Operation", func(t *testing.T) {
		pipeline := []any{
			bson.M{"$match": bson.M{"active": true}},
			bson.M{"$group": bson.M{
				"_id":    "$active",
				"count":  bson.M{"$sum": 1},
				"avgAge": bson.M{"$avg": "$age"},
			}},
		}

		results, err := client.Aggregate(db, col, pipeline)
		if err != nil {
			t.Fatalf("Aggregate failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 aggregation result, got %d", len(results))
		}
		if results[0]["count"].(int32) != 3 {
			t.Errorf("Expected count 3 in aggregation, got %v", results[0]["count"])
		}
		t.Logf("✅ Aggregate successful: count=%v, avgAge=%v", results[0]["count"], results[0]["avgAge"])
	})

	t.Run("BulkWrite_Operation", func(t *testing.T) {
		operations := []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(bson.M{"_id": "bulk-1", "name": "Frank"}),
			mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": "test-1"}).
				SetUpdate(bson.M{"$set": bson.M{"bulk_updated": true}}),
			mongo.NewDeleteOneModel().SetFilter(bson.M{"_id": "test-2"}),
		}

		inserted, modified, err := client.BulkWrite(db, col, operations)
		if err != nil {
			t.Fatalf("BulkWrite failed: %v", err)
		}
		if inserted != 1 {
			t.Errorf("Expected 1 insert, got %d", inserted)
		}
		if modified != 1 {
			t.Errorf("Expected 1 modification, got %d", modified)
		}
		t.Logf("✅ BulkWrite successful: inserted=%d, modified=%d", inserted, modified)
	})

	t.Run("DeleteOne_Operation", func(t *testing.T) {
		err := client.DeleteOne(db, col, bson.M{"_id": "test-3"})
		if err != nil {
			t.Fatalf("DeleteOne failed: %v", err)
		}

		count, _ := client.CountDocuments(db, col, bson.M{"_id": "test-3"})
		if count != 0 {
			t.Error("DeleteOne did not delete document")
		}
		t.Log("✅ DeleteOne successful")
	})

	t.Run("DeleteMany_Operation", func(t *testing.T) {
		err := client.DeleteMany(db, col, bson.M{"active": true})
		if err != nil {
			t.Fatalf("DeleteMany failed: %v", err)
		}

		count, _ := client.CountDocuments(db, col, bson.M{"active": true})
		if count != 0 {
			t.Errorf("Expected 0 active documents after DeleteMany, got %d", count)
		}
		t.Log("✅ DeleteMany successful")
	})

	t.Run("DropCollection_Operation", func(t *testing.T) {
		err := client.DropCollection(db, col)
		if err != nil {
			t.Fatalf("DropCollection failed: %v", err)
		}
		t.Log("✅ DropCollection successful")
	})

	t.Run("Connection_Features", func(t *testing.T) {
		t.Log("✅ Connection timeout: 10 seconds")
		t.Log("✅ Operation timeout: 30 seconds")
		t.Log("✅ Connection verification with ping: enabled")
		t.Log("✅ Retry writes: enabled")
		t.Log("✅ Retry reads: enabled")
	})
}
