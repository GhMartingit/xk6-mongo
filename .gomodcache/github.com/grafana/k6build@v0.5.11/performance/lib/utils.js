// returns a random element from an array
export function randomFromArray(items) {
        if (items.length == 0) {
                return null
        }
        return items[Math.floor(Math.random()*items.length)];
}

// return a random element from the object
export function randomFromDict(dict) {
        return randomFromArray(Object.keys(dict))
}