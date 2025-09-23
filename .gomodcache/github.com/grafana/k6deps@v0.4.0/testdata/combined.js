"use k6 > 0.54";
"use k6 with k6/x/faker > 0.4.0";
"use k6 with k6/x/sql >= 1.0.1";

import faker from "./faker.js";
import sql from "./sql.js";

export { setup, teardown } from "./sql.js";

export default () => {
  faker();
  sql();
};
