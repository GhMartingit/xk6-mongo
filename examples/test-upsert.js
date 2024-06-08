import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');
const id = 'tbF3SqSMl6OpjgQ';
const db = "testdb";
const col = "testcollection";

export function setup() {
  let doc = {
    update_id: id,
    correlationId: 'test--mongodb',
    title: 'Perf test experiment',
    url: 'example.com',
    locale: 'en',
    time: `${new Date(Date.now()).toISOString()}`
  };
  
  let error = client.insert(db, col, doc);
  if (error) 
    console.log(error.message);
}

export default () => {
  let error = client.upsert(db, col, {update_id: id}, {$set: {locale: 'en', title: 'This is a new document'}})
  if (error) 
    console.log(error.message);
}