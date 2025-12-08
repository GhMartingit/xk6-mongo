package xk6_mongo

import (
	"testing"

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
		{"camelCase", "Camelcase"},      // Current behavior: lowercases then capitalizes
		{"PascalCase", "Pascalcase"},    // Current behavior: lowercases then capitalizes
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
