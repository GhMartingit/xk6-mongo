import xk6_mongo from 'k6/x/mongo';

// Transactions require a MongoDB replica set or sharded cluster
const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
  const session = client.startSession();

  try {
    session.startTransaction();

    // Perform multiple operations within a single transaction
    session.insert("testdb", "accounts", { _id: "acc-1", name: "Alice", balance: 1000 });
    session.insert("testdb", "accounts", { _id: "acc-2", name: "Bob", balance: 500 });

    // Transfer money: debit Alice, credit Bob
    session.updateOne("testdb", "accounts", { _id: "acc-1" }, { $inc: { balance: -200 } });
    session.updateOne("testdb", "accounts", { _id: "acc-2" }, { $inc: { balance: 200 } });

    // Verify balances within the transaction
    const alice = session.findOne("testdb", "accounts", { _id: "acc-1" });
    const bob = session.findOne("testdb", "accounts", { _id: "acc-2" });
    console.log(`Alice balance: ${alice.balance}, Bob balance: ${bob.balance}`);

    session.commitTransaction();
    console.log("Transaction committed successfully");
  } catch (e) {
    session.abortTransaction();
    console.log(`Transaction aborted: ${e.message}`);
  } finally {
    session.endSession();
  }
}

export function teardown() {
  client.dropCollection("testdb", "accounts");
}
