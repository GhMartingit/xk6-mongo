# Examples

This folder contains simple k6 scripts that exercise the xk6-mongo extension end-to-end against a local MongoDB instance.

## Prerequisites

- Docker Desktop (or any Docker-compatible runtime) running
- Built k6 binary with the extension in the project root (`./k6`)

Build the k6 binary:

```bash
make build
# or: xk6 build --with github.com/GhMartingit/xk6-mongo
```

Start MongoDB locally:

```bash
docker run -d --rm --name xk6-mongo -p 27017:27017 mongo:7
```

Stop MongoDB when done:

```bash
docker stop xk6-mongo
```

## Suggested run order

The scripts are independent, but this order produces readable output:

1. `test-dropcollection.js` (ensure a clean slate)
2. `test-insert.js` (insert a single document)
3. `test-find.js` (find a document)
4. `test-update.js` (update a single document)
5. `test-findoneandupdate.js` (update + return updated doc; logs the updated document)
6. `test-insertmany.js` (insert a batch of documents)
7. `test-updatemany.js` (bulk update documents)
8. `test-countdocuments.js` (count matching documents)
9. `test-findall.js` (list all documents in collection)
10. `test-distinct.js` (distinct values for a field)
11. `test-aggregate.js` (aggregation pipeline)
12. `test-deletemany.js` (delete many)
13. `test-delete.js` (delete one)
14. `test-dropcollection.js` (clean up)

Run a script like this:

```bash
./k6 run examples/test-insert.js
```

## Notes

- Methods that return data (e.g., `findOne`, `findAll`, `distinct`, `aggregate`) return a single value to JS. Use a single assignment, for example:

  ```js
  const result = client.findOne("testdb", "testcollection", { correlationId: "test--mongodb" });
  console.log(result);
  ```

- Update helpers accept either a full update document with operators (e.g., `{ $set: { ... } }`) or a plain object, which is automatically wrapped in `$set`.
