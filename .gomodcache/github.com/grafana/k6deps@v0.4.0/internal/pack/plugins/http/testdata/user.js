export function newUser(name) {
  return { name, id: Math.floor(Math.random() * Number.MAX_SAFE_INTEGER) };
}
