/// @ts-ignore
import { sleep } from "k6";
/// @ts-ignore
import { describe, expect } from "https://jslib.k6.io/k6chaijs/4.3.4.3/index.js";

export interface User {
  name: string;
  id: number;
}

class UserAccount implements User {
  name: string;
  id: number;

  constructor(name: string) {
    this.name = name;
    this.id = Math.floor(Math.random() * Number.MAX_SAFE_INTEGER);
  }
}

export function newUser(name: string): User {
  sleep(1);
  return new UserAccount(name);
}

export { describe, expect };
