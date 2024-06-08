import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
  let result, error = client.distinct("testdb", "testcollection", "correlationId", {});
  if (error) 
    console.log(error.message);
  else 
    console.log(`Distinct correlationId values: ${result}`);
}