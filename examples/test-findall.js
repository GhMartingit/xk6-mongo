import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
  let results, error = client.findAll("testdb", "testcollection");
  if (error) 
    console.log(error.message);
  else 
    console.log(`Number of documents: ${results.length}`);
}