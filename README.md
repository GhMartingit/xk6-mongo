# xk6-mongo

A k6 extension for interacting with MongoDB while performance testing.

## Features

- **CRUD Operations**: Insert, InsertMany, Find, FindOne, FindAll, UpdateOne, UpdateMany, DeleteOne, DeleteMany
- **Advanced Operations**: Upsert, FindOneAndUpdate, Aggregate, Distinct, CountDocuments
- **Bulk Operations**: BulkWrite for mixed insert/update/delete operations
- **Collection Management**: DropCollection
- **Flexible Filters**: Complex query support for all filter parameters
- **Connection Management**: Automatic connection verification with timeout handling
- **Performance**: Built-in operation timeouts and cursor management

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

## Configuration

### Connection Timeouts

The extension includes built-in timeout handling:

- **Connection timeout**: 10 seconds (connection establishment and ping verification)
- **Operation timeout**: 30 seconds (default for all database operations)

### Connection Pooling

You can configure MongoDB connection pool settings via client options:

```js
const clientOptions = {
    "max_pool_size": 100,
    "min_pool_size": 10,
    "max_connecting": 10
};

const client = xk6_mongo.newClientWithOptions('mongodb://localhost:27017', clientOptions);
```

## Examples

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

If we need to pass extra options to the driver connection we can pass a plain JavaScript object.

Snake_case and camelCase keys are automatically translated to the underlying Go driver field names. In this example, we are specifying the use of the [stable api](https://www.mongodb.com/docs/drivers/go/v1.15/fundamentals/stable-api/), with strict compatibility. We are also setting the application name via the `app_name` property as `"k6-test-app"`, so the connection can be identified in the logs server-side.

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

// When using update helpers you can provide either a full update document
// (with operators like $set) or a plain object. Plain objects are
// automatically wrapped in $set before being sent to MongoDB.
```

### Complex filter example

```js
import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
    const result = client.findOne(
        "testdb",
        "testcollection",
        { score: { "$gte": 10 } }
    );
    console.log(result);
}
```

### Advanced Find with Options

The `findWithOptions` method provides fine-grained control over query behavior:

```js
import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
    const options = {
        limit: 100,
        skip: 10,
        batch_size: 50,  // Process results in batches of 50
        sort: { createdAt: -1 },
        projection: { name: 1, email: 1, _id: 0 }  // Only return specific fields
    };

    const results = client.findWithOptions(
        "testdb",
        "users",
        { active: true },
        options
    );
    console.log(`Found ${results.length} users`);
}
```

### Bulk Operations

```js
import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
    // BulkWrite returns insertedCount and modifiedCount
    const [inserted, modified] = client.bulkWrite(
        "testdb",
        "testcollection",
        [
            { insertOne: { document: { name: "Alice" } } },
            { updateOne: { filter: { name: "Bob" }, update: { $set: { age: 30 } } } },
            { deleteOne: { filter: { name: "Charlie" } } }
        ]
    );
    console.log(`Inserted: ${inserted}, Modified: ${modified}`);
}
```

## API Reference

### Connection Methods

- `newClient(uri)` - Create a new MongoDB client with default options
- `newClientWithOptions(uri, options)` - Create a client with custom connection options
- `disconnect()` - Close the connection to MongoDB

### CRUD Operations

- `insert(db, collection, document)` - Insert a single document
- `insertMany(db, collection, documents)` - Insert multiple documents
- `find(db, collection, filter, sort, limit)` - Find documents with basic options
- `findWithOptions(db, collection, filter, options)` - Find with advanced options (batch size, projection, skip)
- `findOne(db, collection, filter)` - Find a single document
- `findAll(db, collection)` - Find all documents in a collection
- `updateOne(db, collection, filter, update)` - Update a single document
- `updateMany(db, collection, filter, update)` - Update multiple documents
- `deleteOne(db, collection, filter)` - Delete a single document
- `deleteMany(db, collection, filter)` - Delete multiple documents

### Advanced Operations

- `upsert(db, collection, filter, document)` - Insert or update a document
- `findOneAndUpdate(db, collection, filter, update)` - Find and update atomically, returns updated document
- `aggregate(db, collection, pipeline)` - Run aggregation pipeline
- `distinct(db, collection, field, filter)` - Get distinct values for a field
- `countDocuments(db, collection, filter)` - Count documents matching filter
- `bulkWrite(db, collection, operations)` - Execute multiple write operations in one call

### Collection Management

- `dropCollection(db, collection)` - Drop a collection

## Performance Tips

1. **Use batch size** for large result sets to control memory usage
2. **Use projections** to retrieve only needed fields
3. **Configure connection pooling** for high-concurrency tests
4. **Use bulk operations** for multiple writes to reduce network overhead
5. **Add indexes** to your MongoDB collections for better query performance

## Error Handling

All operations include built-in validation and timeout handling:

```js
const error = client.insert("testdb", "testcol", doc);
if (error) {
    console.error(`Insert failed: ${error.message}`);
}
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Security

For security best practices and reporting vulnerabilities, see [SECURITY.md](SECURITY.md).

## Benchmarks

Run performance benchmarks:

```bash
go test -bench=. -benchmem ./...
```

Recent benchmark results show excellent performance:

- Input validation: ~24 ns/op with 0 allocations
- PascalCase conversion: ~230 ns/op
- Update document detection: <3 ns/op
- Pipeline detection: <1 ns/op

## License

This project is licensed under the same license as the k6 project.

## Support

- 📖 [Documentation](README.md)
- 🐛 [Issue Tracker](https://github.com/GhMartingit/xk6-mongo/issues)
- 💬 [Discussions](https://github.com/GhMartingit/xk6-mongo/discussions)
- 📝 [Changelog](CHANGELOG.md)

## Acknowledgments

Built with [k6](https://k6.io/) and the [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver).
