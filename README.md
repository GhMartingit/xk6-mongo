# MongoDB k6 extension

K6 extension to perform tests on mongo.

## Currently Supported Commands

- Supports inserting a document.
- Supports inserting document batch.
- Supports find a document based on filter.
- Supports find all documents of a collection.
- Supports upserting a document based on filter.
- Supports bulk upserting documents based on filters.
- Supports aggregation pipelines.
- Supports finding distinct values for a field in a collection based on a filter.
- Supports delete first document based on filter.
- Supports deleting all documents for a specific filter.
- Filter parameters for `findOne`, `deleteOne`, and `deleteMany` accept any object, enabling complex queries.
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
    xk6 build --with  github.com/GhMartingit/xk6-mongo
    ```

   This will create a k6 binary that includes the xk6-mongo extension in your local folder. This k6 binary can now run a k6 test.

### Development
To make development a little smoother, use the `Makefile` in the root folder. The default target will format your code, run tests, and create a `k6` binary with your local code rather than from GitHub.

```shell
git clone git@github.com/GhMartingit/xk6-mongo.git
cd xk6-mongo
make build
```

Using the `k6` binary with `xk6-mongo`, run the k6 test as usual:

```bash
./k6 run test.js

```

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


### Passing custom `clientOptions` to the mongo client

If we need to pass extra options to the driver connection we can pass a [ClientOptions](https://pkg.go.dev/go.mongodb.org/mongo-driver@v1.15.0/mongo/options#ClientOptions) object.

In this example, we are specifying the use of the [stable api](https://www.mongodb.com/docs/drivers/go/v1.15/fundamentals/stable-api/), with strict compatibility. We are also setting the application name via the `app_name` property as `"k6-test-app"`, so the connection can be identified in the logs server-side.

```js
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

    client.insert("testdb", "testcollection", doc);
}
```


### Complex filter example
```js
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
}
```
