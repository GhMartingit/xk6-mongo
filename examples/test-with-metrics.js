import xk6_mongo from 'k6/x/mongo';
import { check, sleep } from 'k6';

// Test configuration
export const options = {
    stages: [
        { duration: '10s', target: 10 },  // Ramp up to 10 VUs
        { duration: '30s', target: 10 },  // Stay at 10 VUs
        { duration: '10s', target: 0 },   // Ramp down
    ],
    thresholds: {
        checks: ['rate>0.95'],
        // MongoDB-specific thresholds (will be available when metrics are integrated)
        // 'mongo_insert_error_count': ['count==0'],
        // 'mongo_insert_seconds': ['p(95)<0.2'],
        // 'mongo_find_seconds': ['avg<0.1'],
        // 'mongo_connection_error_count': ['count==0'],
    },
};

const uri = __ENV.MONGODB_URI || 'mongodb://localhost:27017';
const client = xk6_mongo.newClient(uri);

export function setup() {
    // Create test data
    console.log('Setting up test data...');

    // Clean up any existing test data
    try {
        client.deleteMany('metricstest', 'users', {});
    } catch (e) {
        // Ignore errors if collection doesn't exist
    }

    console.log('Setup complete');
}

export default function() {
    const userId = `user-${__VU}-${__ITER}`;

    // Test 1: Insert a document
    const insertStart = Date.now();
    const insertDoc = {
        _id: userId,
        name: `User ${__VU}`,
        email: `user${__VU}@example.com`,
        age: 20 + (__VU % 50),
        active: true,
        createdAt: new Date().toISOString(),
        data: {
            preferences: {
                theme: 'dark',
                language: 'en'
            },
            stats: {
                logins: __ITER,
                lastLogin: new Date().toISOString()
            }
        }
    };

    try {
        client.insert('metricstest', 'users', insertDoc);
        const insertDuration = Date.now() - insertStart;

        check(insertDuration, {
            'Insert completed': () => true,
            'Insert under 100ms': (d) => d < 100,
        });
    } catch (e) {
        check(null, {
            'Insert failed': () => false,
        });
        console.error(`Insert error: ${e}`);
    }

    // Test 2: Find the document
    const findStart = Date.now();
    try {
        const results = client.find(
            'metricstest',
            'users',
            { _id: userId },
            null,
            1
        );
        const findDuration = Date.now() - findStart;

        check(results, {
            'Find returned results': (r) => r && r.length > 0,
            'Find returned correct user': (r) => r && r.length > 0 && r[0]._id === userId,
            'Find under 50ms': () => findDuration < 50,
        });
    } catch (e) {
        check(null, {
            'Find failed': () => false,
        });
        console.error(`Find error: ${e}`);
    }

    // Test 3: Update the document
    const updateStart = Date.now();
    try {
        client.updateOne(
            'metricstest',
            'users',
            { _id: userId },
            {
                lastUpdated: new Date().toISOString(),
                'data.stats.logins': __ITER + 1
            }
        );
        const updateDuration = Date.now() - updateStart;

        check(updateDuration, {
            'Update completed': () => true,
            'Update under 100ms': (d) => d < 100,
        });
    } catch (e) {
        check(null, {
            'Update failed': () => false,
        });
        console.error(`Update error: ${e}`);
    }

    // Test 4: Find with filter (multiple docs)
    if (__ITER % 5 === 0) {
        const findAllStart = Date.now();
        try {
            const results = client.find(
                'metricstest',
                'users',
                { active: true },
                { age: 1 },
                100
            );
            const findAllDuration = Date.now() - findAllStart;

            check(results, {
                'FindAll returned results': (r) => r && r.length > 0,
                'FindAll under 200ms': () => findAllDuration < 200,
            });
        } catch (e) {
            check(null, {
                'FindAll failed': () => false,
            });
            console.error(`FindAll error: ${e}`);
        }
    }

    // Test 5: Aggregation (every 10 iterations)
    if (__ITER % 10 === 0) {
        const aggStart = Date.now();
        try {
            const pipeline = [
                { $match: { active: true } },
                {
                    $group: {
                        _id: '$active',
                        count: { $sum: 1 },
                        avgAge: { $avg: '$age' }
                    }
                }
            ];

            const results = client.aggregate('metricstest', 'users', pipeline);
            const aggDuration = Date.now() - aggStart;

            check(results, {
                'Aggregation completed': (r) => r && r.length > 0,
                'Aggregation under 300ms': () => aggDuration < 300,
            });
        } catch (e) {
            check(null, {
                'Aggregation failed': () => false,
            });
            console.error(`Aggregation error: ${e}`);
        }
    }

    // Small delay between iterations
    sleep(0.1);
}

export function teardown() {
    console.log('Cleaning up test data...');

    // Count final documents
    try {
        const count = client.countDocuments('metricstest', 'users', {});
        console.log(`Total documents created: ${count}`);

        // Clean up
        client.deleteMany('metricstest', 'users', {});
    } catch (e) {
        console.error(`Cleanup error: ${e}`);
    }

    client.disconnect();
    console.log('Teardown complete');
}
