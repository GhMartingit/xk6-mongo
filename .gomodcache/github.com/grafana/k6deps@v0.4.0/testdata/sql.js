import sql from "k6/x/sql";
import driver from "k6/x/sql/driver/ramsql";

const db = sql.open(driver);

export function setup() {
  db.exec(`CREATE TABLE IF NOT EXISTS namevalues (
           id integer PRIMARY KEY AUTOINCREMENT,
           name varchar NOT NULL,
           value varchar);`);
}

export function teardown() {
  db.close();
}

export default function () {
  db.exec("INSERT INTO namevalues (name, value) VALUES('foo', 'bar');");
}
