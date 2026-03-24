import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export function setup() {
  // Insert some initial documents for bulk operations
  client.insert("testdb", "testcollection", { name: "Alice", age: 30, active: true });
  client.insert("testdb", "testcollection", { name: "Bob", age: 25, active: true });
  client.insert("testdb", "testcollection", { name: "Charlie", age: 35, active: false });
}

export default () => {
  const [inserted, modified] = client.bulkWrite(
    "testdb",
    "testcollection",
    [
      { insertOne: { document: { name: "Diana", age: 28, active: true } } },
      { updateOne: { filter: { name: "Alice" }, update: { $set: { age: 31 } } } },
      { deleteOne: { filter: { name: "Charlie" } } }
    ]
  );
  console.log(`Inserted: ${inserted}, Modified: ${modified}`);
}

export function teardown() {
  client.dropCollection("testdb", "testcollection");
}
