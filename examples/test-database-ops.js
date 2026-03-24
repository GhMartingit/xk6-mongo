import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export function setup() {
  // Create some collections by inserting data
  client.insert("testdb", "collection1", { name: "doc1" });
  client.insert("testdb", "collection2", { name: "doc2" });
  client.insert("testdb", "collection3", { name: "doc3" });
}

export default () => {
  // List all collections in a database
  const collections = client.listCollections("testdb");
  console.log(`Collections: ${JSON.stringify(collections.map(c => c.name))}`);

  // Drop a specific collection
  client.dropCollection("testdb", "collection3");

  // Verify the collection was dropped
  const remaining = client.listCollections("testdb");
  console.log(`Remaining collections: ${JSON.stringify(remaining.map(c => c.name))}`);
}

export function teardown() {
  // Drop the entire test database
  client.dropDatabase("testdb");
}
