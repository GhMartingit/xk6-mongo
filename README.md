# MongoDB k6 extension

K6 extension to perform tests on mongo.

## Currently Supported Commands

- Supports inserting a document.
- Supports inserting document batch.
- Supports find a document based on filter.
- Supports find all documents of a collection.
- Supports delete first document based on filter.
- Supports deleting all documents for a specific filter.
- Supports dropping a collection.

# xk6-mongo

A k6 extension for interacting with mongoDb while testing.

## Build

To build a custom `k6` binary with this extension, first ensure you have the prerequisites:

- [Go toolchain](https://go101.org/article/go-toolchain.html)
- Git

1. Download [xk6](https://github.com/grafana/xk6):

    ```bash
    go install go.k6.io/xk6/cmd/xk6@latest
    ```

2. [Build the k6 binary](https://github.com/grafana/xk6#command-usage):

    ```bash
    xk6 build --with  github.com/alexkunde/xk6-mongo
    ```

   This will create a k6 binary that includes the xk6-mongo extension in your local folder. This k6 binary can now run a k6 test.

### Development

To make development a little smoother, use the `Makefile` in the root folder. The default target will format your code, run tests, and create a `k6` binary with your local code rather than from GitHub.

```shell
git clone git@github.com/alexkunde/xk6-mongo.git
cd xk6-mongo
make build
```

Using the `k6` binary with `xk6-mongo`, run the k6 test as usual:

```bash
./k6 run test.js

```

## Examples

### Connections

```js
import xk6_mongo from 'k6/x/mongo';

const mongodb_escaped_password = xk6_mongo.mongoEncode('All@Special%%And$$Cool');
const mongodb_escaped_user = xk6_mongo.mongoEncode('Also$%special');
// SRV Connection
const mongodb_connection_string = `mongodb+srv://${mongodb_escaped_user}:${mongodb_escaped_password}@localhost:27017/db?authSource=admin`;
const client = xk6_mongo.newClient(mongodb_connection_string);

// Direct Connection
const mongodb_connection_string = `mongodb://${mongodb_escaped_user}:${mongodb_escaped_password}@localhost:27017/db?connect=direct`;
const client = xk6_mongo.newClient(mongodb_connection_string);
```

### Retrieve Document

```js
import xk6_mongo from 'k6/x/mongo';

const mongodb_escaped_password = xk6_mongo.mongoEncode('All@Special%%And$$Cool');
const mongodb_escaped_user = xk6_mongo.mongoEncode('Also$%special');
const mongodb_connection_string = `mongodb://${mongodb_escaped_user}:${mongodb_escaped_password}@localhost:27017/db?connect=direct`;
const client = xk6_mongo.newClient(mongodb_connection_string);

export default ()=> {
    // json will be parsed to bson.D - for what is possible please google
    // if you like to help with examples, I'd be very happy
    let myDoc = client.findOne('database', 'collection', '{"correlationId": "test--mongodb"}');
    // result will be an object, so you can do stuff like
    let myTitle = MyDoc.title;
}
```

### Document Insertion Test

```js
import xk6_mongo from 'k6/x/mongo';

const mongodb_escaped_password = xk6_mongo.mongoEncode('All@Special%%And$$Cool');
const mongodb_escaped_user = xk6_mongo.mongoEncode('Also$%special');
const mongodb_connection_string = `mongodb://${mongodb_escaped_user}:${mongodb_escaped_password}@localhost:27017/db?connect=direct`;

export default ()=> {

    let doc = {
        correlationId: `test--mongodb`,
        title: 'Perf test experiment',
        url: 'example.com',
        requestId: {
            // this way you can define the type of requestId - in this case Int64
            $numberLong: "12345"
        },
        locale: 'en',
        time: `${new Date(Date.now()).toISOString()}`
    };
    client.insert("testdb", "testcollection", doc);
}

```
