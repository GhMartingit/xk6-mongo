import xk6_mongo from 'k6/x/mongo';

// Change streams require a MongoDB replica set or sharded cluster
const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
  // Watch for all changes on a collection for 5 seconds
  let events = client.watch("testdb", "testcollection", [], 5000);
  console.log(`Received ${events.length} change events`);

  // Watch with a pipeline filter (only insert operations)
  const pipeline = [
    { $match: { operationType: "insert" } }
  ];
  events = client.watch("testdb", "testcollection", pipeline, 3000);
  console.log(`Received ${events.length} insert events`);

  for (const event of events) {
    console.log(`Event type: ${event.operationType}, Document: ${JSON.stringify(event.fullDocument)}`);
  }
}
