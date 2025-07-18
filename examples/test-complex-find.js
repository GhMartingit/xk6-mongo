import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
  let result, error = client.findOne(
    "testdb",
    "testcollection",
    { score: { "$gte": 10 } }
  );
  if (error) {
    console.log(error.message);
  } else {
    console.log(result);
  }
};
