import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export function setup() {
  // Insert sample data
  client.insertMany("testdb", "testcollection", [
    { name: "Alice", email: "alice@example.com", age: 30 },
    { name: "Bob", email: "bob@example.com", age: 25 },
    { name: "Charlie", email: "charlie@example.com", age: 35 },
  ]);
}

export default () => {
  // Create a simple index on the "name" field
  let indexName = client.createIndex("testdb", "testcollection", { name: 1 }, {});
  console.log(`Created index: ${indexName}`);

  // Create a unique index on the "email" field
  indexName = client.createIndex("testdb", "testcollection", { email: 1 }, { unique: true });
  console.log(`Created unique index: ${indexName}`);

  // Create a compound index
  indexName = client.createIndex("testdb", "testcollection", { name: 1, age: -1 }, { name: "name_age_idx" });
  console.log(`Created compound index: ${indexName}`);

  // List all indexes
  const indexes = client.listIndexes("testdb", "testcollection");
  console.log(`Indexes: ${JSON.stringify(indexes)}`);

  // Drop an index by name
  const error = client.dropIndex("testdb", "testcollection", "name_age_idx");
  if (error) {
    console.log(`Error dropping index: ${error.message}`);
  }
}

export function teardown() {
  client.dropCollection("testdb", "testcollection");
}
