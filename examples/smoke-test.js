// Minimal end-to-end smoke test used by CI to verify the extension actually
// loads and works inside a real k6 binary (as opposed to the Go unit tests,
// which exercise the Client directly and never touch k6's module system).
// Unlike the other examples, this one throws on any failure so a failing
// `k6 run` exits non-zero.
import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');
const db = 'smoketest';
const col = 'docs';

export default () => {
  const marker = `smoke-${Date.now()}-${Math.random()}`;

  let err = client.insert(db, col, { marker, value: 1 });
  if (err) {
    throw new Error(`insert failed: ${err.message}`);
  }

  const doc = client.findOne(db, col, { marker });
  if (!doc || doc.value !== 1) {
    throw new Error(`findOne returned unexpected document: ${JSON.stringify(doc)}`);
  }

  err = client.deleteOne(db, col, { marker });
  if (err) {
    throw new Error(`deleteOne failed: ${err.message}`);
  }
};
