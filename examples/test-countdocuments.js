import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
  let count = client.countDocuments("testdb", "testcollection", {correlationId: `test--mongodb`});
  console.log(`Number of documents with correlationId 'test--mongodb': ${count}`);
}
