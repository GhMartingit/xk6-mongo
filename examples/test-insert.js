import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');
export default ()=> {

  let doc = {
      correlationId: `test--mongodb`,
      title: 'Perf test experiment',
      url: 'example.com',
      locale: 'en',
      time: `${new Date(Date.now()).toISOString()}`
    };

    let error = client.insert("testdb", "testcollection", doc);
    if (error) 
        console.log(error.message);
}
