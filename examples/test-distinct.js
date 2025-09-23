import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
  let result = client.distinct("testdb", "testcollection", "correlationId", {});
  console.log(`Distinct correlationId values: ${result}`);
}
