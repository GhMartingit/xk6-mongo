import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
  let result = client.findOne("testdb", "testcollection", {correlationId: `test--mongodb`});
  console.log(JSON.stringify(result));
}
