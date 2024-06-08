import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
  const aggregationPipeline = [
    {
      $match: { correlationId: "test--mongodb" }
    },
    {
      $group: { _id: "$locale", count: { $sum: 1} }
    }
  ];

  let result, error = client.aggregate("testdb", "testcollection", aggregationPipeline);
  if (error) 
    console.log(error.message);
  else 
    console.log(`Aggregation result: ${JSON.stringify(result)}`);
}