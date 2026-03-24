package xk6_mongo

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestValidateDatabaseAndCollection(t *testing.T) {
	tests := []struct {
		name       string
		database   string
		collection string
		wantErr    bool
	}{
		{"valid names", "testdb", "testcol", false},
		{"empty database", "", "testcol", true},
		{"empty collection", "testdb", "", true},
		{"database with slash", "test/db", "testcol", true},
		{"database with backslash", "test\\db", "testcol", true},
		{"database with dot", "test.db", "testcol", true},
		{"database with quote", "test\"db", "testcol", true},
		{"database with dollar", "test$db", "testcol", true},
		{"database with space", "test db", "testcol", true},
		{"valid collection with special chars", "testdb", "test_col-123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDatabaseAndCollection(tt.database, tt.collection)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDatabaseAndCollection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"snake_case", "SnakeCase"},
		{"kebab-case", "KebabCase"},
		{"camelCase", "Camelcase"},   // Current behavior: lowercases then capitalizes
		{"PascalCase", "Pascalcase"}, // Current behavior: lowercases then capitalizes
		{"app_name", "AppName"},
		{"server_api_version", "ServerAPIVersion"}, // API is recognized acronym
		{"api_key", "APIKey"},
		{"user_id", "UserID"},
		{"tls_config", "TLSConfig"},
		{"ocsp_enabled", "OCSPEnabled"},
		{"", ""},
		{"single", "Single"},
		{"with spaces", "WithSpaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestClientValidation(t *testing.T) {
	t.Run("empty connection URI", func(t *testing.T) {
		client := new(Mongo).NewClient("")
		if client != nil {
			t.Error("Expected nil client for empty URI")
		}
	})

	t.Run("invalid connection URI", func(t *testing.T) {
		// This should fail due to connection timeout or invalid URI
		client := new(Mongo).NewClient("mongodb://nonexistent:27017")
		if client != nil {
			t.Error("Expected nil client for unreachable URI")
			_ = client.Disconnect()
		}
	})
}

func TestInsertValidation(t *testing.T) {
	client := &Client{} // Mock client without real connection

	t.Run("nil document", func(t *testing.T) {
		err := client.Insert("db", "col", nil)
		if err != errDocumentNil {
			t.Errorf("Expected errDocumentNil, got %v", err)
		}
	})

	t.Run("empty database", func(t *testing.T) {
		err := client.Insert("", "col", map[string]any{"key": "value"})
		if err == nil {
			t.Error("Expected error for empty database")
		}
	})

	t.Run("empty collection", func(t *testing.T) {
		err := client.Insert("db", "", map[string]any{"key": "value"})
		if err == nil {
			t.Error("Expected error for empty collection")
		}
	})
}

func TestInsertManyValidation(t *testing.T) {
	client := &Client{}

	t.Run("empty documents array", func(t *testing.T) {
		err := client.InsertMany("db", "col", []any{})
		if err != errDocsEmpty {
			t.Errorf("Expected errDocsEmpty, got %v", err)
		}
	})
}

func TestFindValidation(t *testing.T) {
	client := &Client{}

	t.Run("negative limit", func(t *testing.T) {
		_, err := client.Find("db", "col", map[string]any{}, nil, -1)
		if err != errLimitNeg {
			t.Errorf("Expected errLimitNeg, got %v", err)
		}
	})
}

func TestUpdateValidation(t *testing.T) {
	client := &Client{}

	t.Run("nil filter for UpdateOne", func(t *testing.T) {
		err := client.UpdateOne("db", "col", nil, map[string]any{"key": "value"})
		if err != errFilterNil {
			t.Errorf("Expected errFilterNil, got %v", err)
		}
	})

	t.Run("nil filter for UpdateMany", func(t *testing.T) {
		err := client.UpdateMany("db", "col", nil, map[string]any{"key": "value"})
		if err != errFilterNil {
			t.Errorf("Expected errFilterNil, got %v", err)
		}
	})
}

func TestUpsertValidation(t *testing.T) {
	client := &Client{}

	t.Run("nil filter", func(t *testing.T) {
		err := client.Upsert("db", "col", nil, map[string]any{"key": "value"})
		if err != errFilterNil {
			t.Errorf("Expected errFilterNil, got %v", err)
		}
	})
}

func TestAggregateValidation(t *testing.T) {
	client := &Client{}

	t.Run("nil pipeline", func(t *testing.T) {
		_, err := client.Aggregate("db", "col", nil)
		if err != errPipelineNil {
			t.Errorf("Expected errPipelineNil, got %v", err)
		}
	})
}

func TestDistinctValidation(t *testing.T) {
	client := &Client{}

	t.Run("empty field name", func(t *testing.T) {
		_, err := client.Distinct("db", "col", "", map[string]any{})
		if err == nil {
			t.Error("Expected error for empty field name")
		}
	})
}

func TestFindOneAndUpdateValidation(t *testing.T) {
	client := &Client{}

	t.Run("nil filter", func(t *testing.T) {
		_, err := client.FindOneAndUpdate("db", "col", nil, map[string]any{"key": "value"})
		if err != errFilterNil {
			t.Errorf("Expected errFilterNil, got %v", err)
		}
	})
}

func TestBulkWriteValidation(t *testing.T) {
	client := &Client{}

	t.Run("empty operations array", func(t *testing.T) {
		_, _, err := client.BulkWrite("db", "col", []mongo.WriteModel{})
		if err == nil {
			t.Error("Expected error for empty operations array")
		}
	})
}

func TestCreateIndexValidation(t *testing.T) {
	client := &Client{}

	t.Run("nil keys", func(t *testing.T) {
		_, err := client.CreateIndex("db", "col", nil, nil)
		if err != errKeysNil {
			t.Errorf("Expected errKeysNil, got %v", err)
		}
	})

	t.Run("empty database", func(t *testing.T) {
		_, err := client.CreateIndex("", "col", bson.M{"field": 1}, nil)
		if err == nil {
			t.Error("Expected error for empty database")
		}
	})

	t.Run("empty collection", func(t *testing.T) {
		_, err := client.CreateIndex("db", "", bson.M{"field": 1}, nil)
		if err == nil {
			t.Error("Expected error for empty collection")
		}
	})
}

func TestDropIndexValidation(t *testing.T) {
	client := &Client{}

	t.Run("empty index name", func(t *testing.T) {
		err := client.DropIndex("db", "col", "")
		if err != errIndexNameEmpty {
			t.Errorf("Expected errIndexNameEmpty, got %v", err)
		}
	})

	t.Run("empty database", func(t *testing.T) {
		err := client.DropIndex("", "col", "idx_name")
		if err == nil {
			t.Error("Expected error for empty database")
		}
	})
}

func TestListIndexesValidation(t *testing.T) {
	client := &Client{}

	t.Run("empty database", func(t *testing.T) {
		_, err := client.ListIndexes("", "col")
		if err == nil {
			t.Error("Expected error for empty database")
		}
	})

	t.Run("empty collection", func(t *testing.T) {
		_, err := client.ListIndexes("db", "")
		if err == nil {
			t.Error("Expected error for empty collection")
		}
	})
}

func TestDropDatabaseValidation(t *testing.T) {
	client := &Client{}

	t.Run("empty database", func(t *testing.T) {
		err := client.DropDatabase("")
		if err != errDatabaseEmpty {
			t.Errorf("Expected errDatabaseEmpty, got %v", err)
		}
	})
}

func TestListCollectionsValidation(t *testing.T) {
	client := &Client{}

	t.Run("empty database", func(t *testing.T) {
		_, err := client.ListCollections("")
		if err != errDatabaseEmpty {
			t.Errorf("Expected errDatabaseEmpty, got %v", err)
		}
	})
}

func TestNormalizeKeys(t *testing.T) {
	t.Run("simple map", func(t *testing.T) {
		input := map[string]any{"max_pool_size": 100, "min_pool_size": 10}
		result := normalizeKeys(input).(map[string]any)
		if _, ok := result["MaxPoolSize"]; !ok {
			t.Error("Expected MaxPoolSize key after normalization")
		}
		if _, ok := result["MinPoolSize"]; !ok {
			t.Error("Expected MinPoolSize key after normalization")
		}
	})

	t.Run("nested map", func(t *testing.T) {
		input := map[string]any{
			"server_api_options": map[string]any{
				"server_api_version": "1",
			},
		}
		result := normalizeKeys(input).(map[string]any)
		nested, ok := result["ServerAPIOptions"].(map[string]any)
		if !ok {
			t.Fatal("Expected nested map after normalization")
		}
		if _, ok := nested["ServerAPIVersion"]; !ok {
			t.Error("Expected ServerAPIVersion key in nested map")
		}
	})

	t.Run("bson.M input", func(t *testing.T) {
		input := bson.M{"app_name": "test"}
		result := normalizeKeys(input).(map[string]any)
		if _, ok := result["AppName"]; !ok {
			t.Error("Expected AppName key after normalization")
		}
	})

	t.Run("slice input", func(t *testing.T) {
		input := []any{
			map[string]any{"key_name": "value"},
		}
		result := normalizeKeys(input).([]any)
		inner := result[0].(map[string]any)
		if _, ok := inner["KeyName"]; !ok {
			t.Error("Expected KeyName key after normalization of array element")
		}
	})

	t.Run("primitive input", func(t *testing.T) {
		result := normalizeKeys("hello")
		if result != "hello" {
			t.Errorf("Expected primitive to pass through, got %v", result)
		}
	})
}

func TestHasOperatorInKeysD(t *testing.T) {
	t.Run("with operator", func(t *testing.T) {
		d := bson.D{{Key: "$set", Value: bson.M{"name": "test"}}}
		if !hasOperatorInKeysD(d) {
			t.Error("Expected true for document with $ operator")
		}
	})

	t.Run("without operator", func(t *testing.T) {
		d := bson.D{{Key: "name", Value: "test"}, {Key: "age", Value: 30}}
		if hasOperatorInKeysD(d) {
			t.Error("Expected false for document without $ operator")
		}
	})

	t.Run("empty document", func(t *testing.T) {
		d := bson.D{}
		if hasOperatorInKeysD(d) {
			t.Error("Expected false for empty document")
		}
	})

	t.Run("mixed keys", func(t *testing.T) {
		d := bson.D{{Key: "name", Value: "test"}, {Key: "$inc", Value: bson.M{"count": 1}}}
		if !hasOperatorInKeysD(d) {
			t.Error("Expected true when at least one key has $ operator")
		}
	})
}

func TestClientOptionsFromMap(t *testing.T) {
	t.Run("basic options", func(t *testing.T) {
		raw := map[string]any{
			"app_name": "test-app",
		}
		opts, err := clientOptionsFromMap("mongodb://localhost:27017", raw)
		if err != nil {
			t.Fatalf("clientOptionsFromMap failed: %v", err)
		}
		if opts == nil {
			t.Fatal("Expected non-nil client options")
		}
	})

	t.Run("empty map", func(t *testing.T) {
		opts, err := clientOptionsFromMap("mongodb://localhost:27017", map[string]any{})
		if err != nil {
			t.Fatalf("clientOptionsFromMap failed: %v", err)
		}
		if opts == nil {
			t.Fatal("Expected non-nil client options")
		}
	})
}

func TestPrepareClientOptions(t *testing.T) {
	t.Run("nil opts", func(t *testing.T) {
		opts, err := prepareClientOptions("mongodb://localhost:27017", nil)
		if err != nil {
			t.Fatalf("prepareClientOptions failed: %v", err)
		}
		if opts == nil {
			t.Fatal("Expected non-nil client options")
		}
	})

	t.Run("map opts", func(t *testing.T) {
		opts, err := prepareClientOptions("mongodb://localhost:27017", map[string]any{"app_name": "test"})
		if err != nil {
			t.Fatalf("prepareClientOptions failed: %v", err)
		}
		if opts == nil {
			t.Fatal("Expected non-nil client options")
		}
	})

	t.Run("bson.M opts", func(t *testing.T) {
		opts, err := prepareClientOptions("mongodb://localhost:27017", bson.M{"app_name": "test"})
		if err != nil {
			t.Fatalf("prepareClientOptions failed: %v", err)
		}
		if opts == nil {
			t.Fatal("Expected non-nil client options")
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		_, err := prepareClientOptions("mongodb://localhost:27017", "invalid")
		if err == nil {
			t.Error("Expected error for unsupported type")
		}
	})
}
