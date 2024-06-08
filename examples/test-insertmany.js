import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');
const batchsize = 50;

export default () => {
  let docobjs = []

  for (let i = 0; i < batchsize; i++) {
    docobjs.push(getRecord());
  }

  let error = client.insertMany("test", "test", docobjs);
  if (error) 
    console.log(error.message);
}

function getRecord() {
  return {
    _id: `${makeId(15)}`,
    correlationId: `test--couchbase`,
    title: 'Perf test experiment',
    url: 'example.com',
    locale: 'en',
    time: `${new Date(Date.now()).toISOString()}`
  };
}

function makeId(length) {
  let result = ''; 
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  const charactersLength = characters.length;
  let counter = 0;
  while (counter < length) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength));
    counter += 1;
  }

  return result;
}