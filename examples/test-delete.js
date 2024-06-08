import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
  let error = client.deleteOne("testdb", "testcollection", {correlationId: `test--couchbase`});
  if (error) 
    console.log(error.message);
}