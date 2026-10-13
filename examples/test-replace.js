import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');
const db = "testdb";
const col = "testcollection";
const id = 'tbF3SqSMl6OpjgQ-replace';

export function setup() {
  let doc = {
      replace_id: id,
      correlationId: 'test--mongodb',
      title: 'Original document',
      url: 'example.com',
      locale: 'en',
      time: `${new Date(Date.now()).toISOString()}`
    };
  let error = client.insert(db, col, doc);
  if (error)
      console.log(error.message);
}

export default () => {
  const replacement = {
      replace_id: id,
      correlationId: 'test--mongodb',
      title: 'Replaced document',
      status: 'replaced',
      updatedAt: `${new Date(Date.now()).toISOString()}`
    };
  let error = client.replaceOne(db, col, {replace_id: id}, replacement);
  if (error)
    console.log(error.message);
}