import xk6_mongo from 'k6/x/mongo';

const clientOptions = {
  "app_name": "k6-test-app", 
  
  "server_api_options": { 
      "server_api_version": "1",
      "strict": true 
  }
};

const client = xk6_mongo.newClientWithOptions('mongodb://localhost:27017', clientOptions);

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
