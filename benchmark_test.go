package xk6_mongo

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

// Benchmark validation functions
func BenchmarkValidateDatabaseAndCollection(b *testing.B) {
	b.Run("valid", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = validateDatabaseAndCollection("testdb", "testcol")
		}
	})

	b.Run("invalid_with_dot", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = validateDatabaseAndCollection("test.db", "testcol")
		}
	})
}

func BenchmarkToPascalCase(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"snake_case", "max_pool_size"},
		{"kebab-case", "server-api-version"},
		{"with_acronym", "api_key"},
		{"simple", "timeout"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = toPascalCase(tc.input)
			}
		})
	}
}

// Benchmark update document preparation
func BenchmarkPrepareUpdateDocument(b *testing.B) {
	b.Run("plain_map", func(b *testing.B) {
		doc := map[string]any{"name": "test", "age": 30}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = prepareUpdateDocument(doc)
		}
	})

	b.Run("with_operators", func(b *testing.B) {
		doc := bson.M{"$set": bson.M{"name": "test", "age": 30}}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = prepareUpdateDocument(doc)
		}
	})

	b.Run("bson_d", func(b *testing.B) {
		doc := bson.D{{Key: "name", Value: "test"}, {Key: "age", Value: 30}}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = prepareUpdateDocument(doc)
		}
	})
}

// Benchmark key normalization
func BenchmarkNormalizeKeys(b *testing.B) {
	b.Run("simple_map", func(b *testing.B) {
		input := map[string]any{
			"max_pool_size": 100,
			"min_pool_size": 10,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = normalizeKeys(input)
		}
	})

	b.Run("nested_map", func(b *testing.B) {
		input := map[string]any{
			"server_api_options": map[string]any{
				"server_api_version": "1",
				"strict":             true,
			},
			"max_pool_size": 100,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = normalizeKeys(input)
		}
	})

	b.Run("with_arrays", func(b *testing.B) {
		input := map[string]any{
			"settings": []any{
				map[string]any{"key": "value1"},
				map[string]any{"key": "value2"},
			},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = normalizeKeys(input)
		}
	})
}

// Benchmark update document operator detection
func BenchmarkUpdateDocumentHasOperator(b *testing.B) {
	b.Run("with_operator_bson_M", func(b *testing.B) {
		doc := bson.M{"$set": bson.M{"name": "test"}}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = updateDocumentHasOperator(doc)
		}
	})

	b.Run("without_operator_bson_M", func(b *testing.B) {
		doc := bson.M{"name": "test", "age": 30}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = updateDocumentHasOperator(doc)
		}
	})

	b.Run("with_operator_bson_D", func(b *testing.B) {
		doc := bson.D{{Key: "$set", Value: bson.M{"name": "test"}}}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = updateDocumentHasOperator(doc)
		}
	})

	b.Run("without_operator_bson_D", func(b *testing.B) {
		doc := bson.D{{Key: "name", Value: "test"}, {Key: "age", Value: 30}}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = updateDocumentHasOperator(doc)
		}
	})

	b.Run("with_operator_map", func(b *testing.B) {
		doc := map[string]any{"$inc": map[string]any{"count": 1}}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = updateDocumentHasOperator(doc)
		}
	})
}

// Benchmark pipeline detection
func BenchmarkIsPipelineUpdate(b *testing.B) {
	b.Run("bson_D_slice", func(b *testing.B) {
		doc := []bson.D{
			{{Key: "$set", Value: bson.M{"name": "test"}}},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = isPipelineUpdate(doc)
		}
	})

	b.Run("any_slice", func(b *testing.B) {
		doc := []any{
			bson.M{"$set": bson.M{"name": "test"}},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = isPipelineUpdate(doc)
		}
	})

	b.Run("not_pipeline", func(b *testing.B) {
		doc := bson.M{"name": "test"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = isPipelineUpdate(doc)
		}
	})
}

// Memory allocation benchmarks
func BenchmarkMemoryAllocations(b *testing.B) {
	b.Run("validateDatabaseAndCollection", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = validateDatabaseAndCollection("testdb", "testcol")
		}
	})

	b.Run("toPascalCase", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = toPascalCase("max_pool_size")
		}
	})

	b.Run("prepareUpdateDocument_plain", func(b *testing.B) {
		doc := map[string]any{"name": "test", "age": 30}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = prepareUpdateDocument(doc)
		}
	})
}
