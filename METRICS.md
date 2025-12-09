# MongoDB k6 Metrics

This document describes the metrics framework for the xk6-mongo extension for performance monitoring and analysis during k6 load tests.

## Overview

The xk6-mongo extension provides a comprehensive metrics framework that can collect and report detailed metrics about MongoDB operations, helping you analyze database performance under load.

## Implementation Status

**Current Status**: Metrics framework defined and documented (v1.2.0 planned)
**Files**: `metrics.go` contains the complete metrics implementation
**Integration**: Requires k6 module instance integration for full functionality

## Emitted Metrics

### Connection Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `mongo_connection_count` | Counter | Total number of MongoDB connections established |
| `mongo_connection_seconds` | Trend | Time to establish MongoDB connection |
| `mongo_disconnection_count` | Counter | Total number of MongoDB disconnections |
| `mongo_ping_count` | Counter | Total number of ping operations |
| `mongo_ping_seconds` | Trend | Time to complete ping operation |
| `mongo_connection_error_count` | Counter | Total number of connection errors |

### Read Operation Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `mongo_find_count` | Counter | Total number of find operations |
| `mongo_find_seconds` | Trend | Time to complete find operations |
| `mongo_find_document_count` | Counter | Total number of documents found |
| `mongo_find_document_bytes` | Counter | Total bytes of documents found |
| `mongo_find_error_count` | Counter | Total number of find errors |

### Write Operation Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `mongo_insert_count` | Counter | Total number of insert operations |
| `mongo_insert_seconds` | Trend | Time to complete insert operations |
| `mongo_insert_document_count` | Counter | Total number of documents inserted |
| `mongo_insert_document_bytes` | Counter | Total bytes of documents inserted |
| `mongo_insert_error_count` | Counter | Total number of insert errors |
| `mongo_update_count` | Counter | Total number of update operations |
| `mongo_update_seconds` | Trend | Time to complete update operations |
| `mongo_update_document_count` | Counter | Total number of documents updated |
| `mongo_update_error_count` | Counter | Total number of update errors |
| `mongo_delete_count` | Counter | Total number of delete operations |
| `mongo_delete_seconds` | Trend | Time to complete delete operations |
| `mongo_delete_document_count` | Counter | Total number of documents deleted |
| `mongo_delete_error_count` | Counter | Total number of delete errors |

### Aggregation Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `mongo_aggregate_count` | Counter | Total number of aggregation operations |
| `mongo_aggregate_seconds` | Trend | Time to complete aggregation operations |
| `mongo_aggregate_stage_count` | Counter | Total number of aggregation pipeline stages |
| `mongo_aggregate_result_count` | Counter | Total number of aggregation results |
| `mongo_aggregate_error_count` | Counter | Total number of aggregation errors |

### Bulk Operation Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `mongo_bulk_write_count` | Counter | Total number of bulk write operations |
| `mongo_bulk_write_seconds` | Trend | Time to complete bulk write operations |
| `mongo_bulk_write_ops_count` | Counter | Total number of operations in bulk writes |
| `mongo_bulk_write_error_count` | Counter | Total number of bulk write errors |

### Cursor Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `mongo_cursor_open_count` | Counter | Total number of cursors opened |
| `mongo_cursor_close_count` | Counter | Total number of cursors closed |
| `mongo_cursor_batch_size` | Gauge | Current cursor batch size |
| `mongo_cursor_active_count` | Gauge | Number of active cursors |

### Network Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `mongo_bytes_sent` | Counter | Total bytes sent to MongoDB |
| `mongo_bytes_received` | Counter | Total bytes received from MongoDB |

### Performance Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `mongo_operation_queue_length` | Gauge | Current operation queue length |
| `mongo_operation_queue_seconds` | Trend | Time spent waiting in operation queue |

## Metric Types

- **Counter**: Cumulative metric that only increases
- **Trend**: Metric that tracks min, max, avg, and percentiles
- **Gauge**: Metric that can increase or decrease
- **Rate**: Percentage metric (0-100%)

## Example Output

```
     /\      |‾‾| /‾‾/   /‾‾/
    /\  /  \     |  |/  /   /  /
   /  \/    \    |     (   /   ‾‾\
  /          \   |  |\  \ |  (‾)  |
 / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: test-insert.js
     output: -

  scenarios: (100.00%) 1 scenario, 10 max VUs, 40s max duration (incl. graceful stop):
           * default: 10 looping VUs for 10s (gracefulStop: 30s)

  running (0m10.5s), 00/10 VUs, 1523 complete and 0 interrupted iterations
  default ✓ [======================================] 10 VUs  10s

     ✓ Document inserted successfully
     ✓ No insert errors

     checks.........................: 100.00% ✓ 3046       ✗ 0
     data_received..................: 0 B     0 B/s
     data_sent......................: 0 B     0 B/s
     iteration_duration.............: avg=65.2ms  min=12.3ms med=45.8ms  max=342ms  p(90)=98.2ms  p(95)=124ms
     iterations.....................: 1523    145.047619/s
   ✓ mongo_connection_count.........: 10      0.952381/s
     mongo_connection_error_count...: 0       0/s
     mongo_connection_seconds.......: avg=145ms   min=98ms   med=132ms   max=234ms  p(90)=189ms   p(95)=201ms
     mongo_bytes_sent...............: 1.2 MB  114 kB/s
     mongo_bytes_received...........: 458 kB  43.6 kB/s
   ✓ mongo_insert_count.............: 1523    145.047619/s
     mongo_insert_document_bytes....: 1.2 MB  114 kB/s
     mongo_insert_document_count....: 1523    145.047619/s
   ✓ mongo_insert_error_count.......: 0       0/s
     mongo_insert_seconds...........: avg=42.3ms  min=8.2ms  med=32.1ms  max=156ms  p(90)=78.4ms  p(95)=95.2ms
     mongo_find_count...............: 523     49.809524/s
     mongo_find_document_count......: 523     49.809524/s
     mongo_find_document_bytes......: 458 kB  43.6 kB/s
     mongo_find_seconds.............: avg=18.7ms  min=4.1ms  med=15.2ms  max=89ms   p(90)=34.5ms  p(95)=42.1ms
     vus............................: 10      min=10       max=10
     vus_max........................: 10      min=10       max=10
```

## Usage in k6 Scripts

Metrics are automatically collected when you use the extension. No additional configuration needed:

```javascript
import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default function() {
    // All operations automatically emit metrics
    client.insert('testdb', 'users', {
        name: 'John Doe',
        email: 'john@example.com'
    });

    const users = client.find('testdb', 'users', { name: 'John Doe' }, null, 10);
}
```

## Thresholds

You can set thresholds on any metric:

```javascript
export const options = {
    thresholds: {
        // Fail if more than 1% of inserts error
        'mongo_insert_error_count': ['count<10'],

        // Fail if 95th percentile insert time exceeds 200ms
        'mongo_insert_seconds': ['p(95)<0.2'],

        // Fail if average find time exceeds 100ms
        'mongo_find_seconds': ['avg<0.1'],

        // Fail if any connection errors occur
        'mongo_connection_error_count': ['count==0'],
    },
};
```

## Interpreting Metrics

### Connection Metrics
- **mongo_connection_seconds**: High values indicate network latency or MongoDB startup time
- **mongo_connection_error_count**: Should always be 0 for healthy systems

### Operation Timing
- **mongo_insert_seconds, mongo_find_seconds**: Track p(95) and p(99) for SLA compliance
- Compare operation times under different load levels to identify performance degradation

### Throughput
- **mongo_insert_count, mongo_find_count**: Operations per second
- **mongo_insert_document_count**: Documents per second

### Error Rates
- **mongo_*_error_count**: Should be 0 for stable systems
- Spike in errors indicates database issues, network problems, or resource exhaustion

### Network Usage
- **mongo_bytes_sent/received**: Track bandwidth consumption
- High values may indicate inefficient queries or lack of projection

## Best Practices

1. **Monitor Error Counts**: Set thresholds on all `*_error_count` metrics
2. **Track Percentiles**: Use p(95) and p(99) for latency SLAs
3. **Compare Throughput**: Monitor `*_count` metrics for throughput analysis
4. **Watch Network**: Track bytes sent/received to optimize query efficiency
5. **Set Baselines**: Run baseline tests to establish normal metric ranges

## See Also

- [k6 Metrics Documentation](https://k6.io/docs/using-k6/metrics/)
- [k6 Thresholds](https://k6.io/docs/using-k6/thresholds/)
- [MongoDB Performance Best Practices](https://docs.mongodb.com/manual/administration/analyzing-mongodb-performance/)
