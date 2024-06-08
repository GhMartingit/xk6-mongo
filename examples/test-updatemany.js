import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');
const db = "testdb";
const col = "testcollection";

export default () => {
  let error = client.updateMany(db, col, {correlationId: `test--mongodb`}, {locale: 'in', title: 'This is the change for all docs'})
  if (error) 
    console.log(error.message);
}