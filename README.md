# CouchBase k6 extension

K6 extension to perform tests on mongo.

## Currently Supported Commands

- Supports inserting a document.
- Support findOne (Fetch by primary key)
- Support checking query performance

## Examples: 
### Document Insertion Test
```js
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

    client.insert("testdb", "testcollection", doc);
}

```