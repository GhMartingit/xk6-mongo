import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
  let results = client.findAll("testdb", "testcollection");
  console.log(`Number of documents: ${results.length}`);
}
