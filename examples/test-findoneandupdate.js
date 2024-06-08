import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');
export default () => {
  let result, error = client.findOneAndUpdate("testdb", "testcollection", {correlationId: `test--mongodb`}, { $set: { locale: 'it', title: 'Update Document'}})
  if (error) 
    console.log(error.message);
  else 
    console.log(`Updated Document: ${JSON.stringify(result)}`);
}