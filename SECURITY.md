# Security Best Practices

This document outlines security best practices when using xk6-mongo for MongoDB performance testing.

## Table of Contents

- [Connection Security](#connection-security)
- [Authentication](#authentication)
- [Data Protection](#data-protection)
- [Testing Considerations](#testing-considerations)
- [Reporting Security Issues](#reporting-security-issues)

## Connection Security

### Use TLS/SSL Connections

Always use TLS when connecting to MongoDB in production or production-like environments:

```javascript
const clientOptions = {
    "tls": true,
    "tls_ca_file": "/path/to/ca.pem",
    "tls_certificate_key_file": "/path/to/client.pem"
};

const client = xk6_mongo.newClientWithOptions(
    'mongodb://mongodb.example.com:27017',
    clientOptions
);
```

### Connection String Security

**❌ DON'T** hardcode credentials in scripts:
```javascript
// BAD - Credentials exposed in code
const client = xk6_mongo.newClient('mongodb://admin:password123@localhost:27017');
```

**✅ DO** use environment variables:
```javascript
// GOOD - Credentials from environment
const uri = __ENV.MONGODB_URI || 'mongodb://localhost:27017';
const client = xk6_mongo.newClient(uri);
```

### Network Isolation

- Run tests in isolated networks
- Use VPNs for remote MongoDB connections
- Restrict MongoDB ports with firewalls
- Use IP whitelisting when possible

## Authentication

### Use Strong Authentication

Configure MongoDB with authentication enabled:

```javascript
const clientOptions = {
    "auth_source": "admin",
    "auth_mechanism": "SCRAM-SHA-256"  // Use SHA-256, not SHA-1
};

const client = xk6_mongo.newClientWithOptions(uri, clientOptions);
```

### Recommended Authentication Mechanisms

1. **SCRAM-SHA-256** (recommended): Strong password authentication
2. **X.509**: Certificate-based authentication
3. **AWS IAM**: For MongoDB Atlas on AWS

**Avoid:**
- ❌ MONGODB-CR (deprecated)
- ❌ SCRAM-SHA-1 (use SHA-256 instead)

### Database User Permissions

Follow the principle of least privilege:

```javascript
// Test user should have minimal permissions
db.createUser({
    user: "k6_test_user",
    pwd: "secure_password",
    roles: [
        { role: "read", db: "testdb" },      // Read-only for reads
        { role: "readWrite", db: "testdb" }  // Write only to test DB
    ]
})
```

**Don't grant:**
- ❌ `root` role
- ❌ `dbAdmin` on production databases
- ❌ `clusterAdmin` access

## Data Protection

### Sensitive Data in Tests

**Never** use production data for testing:

```javascript
// ❌ BAD - Testing with real user data
const users = client.findAll("production_db", "users");

// ✅ GOOD - Use synthetic test data
const testUser = {
    email: "test@example.com",
    name: "Test User",
    ssn: "XXX-XX-XXXX"  // Masked/fake data
};
```

### Data Anonymization

If you must use production-like data:

1. **Anonymize PII**: Remove/mask personal information
2. **Sanitize inputs**: Remove sensitive fields
3. **Use separate test database**: Never connect to production

```javascript
// Anonymize data before testing
function anonymizeUser(user) {
    return {
        ...user,
        email: `test${user.id}@example.com`,
        name: `Test User ${user.id}`,
        ssn: "XXX-XX-XXXX",
        creditCard: "XXXX-XXXX-XXXX-XXXX"
    };
}
```

### Secure Test Data Cleanup

Always clean up test data after testing:

```javascript
export function teardown() {
    // Clean up test data
    client.deleteMany("testdb", "test_users", {
        createdBy: "k6_test"
    });

    client.disconnect();
}
```

## Testing Considerations

### Rate Limiting

Implement rate limiting to avoid DoS-like behavior:

```javascript
import { sleep } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 10 },   // Ramp up slowly
        { duration: '1m', target: 50 },    // Moderate load
        { duration: '30s', target: 0 },    // Ramp down
    ],
};

export default function() {
    client.insert("testdb", "test_collection", doc);
    sleep(1);  // Avoid overwhelming the server
}
```

### Test Environment Isolation

- ✅ Use dedicated test databases
- ✅ Separate MongoDB instances for testing
- ✅ Use Docker containers for local testing
- ❌ Never test against production databases

### MongoDB Atlas Security

When using MongoDB Atlas:

```javascript
const clientOptions = {
    "tls": true,
    "tls_insecure_skip_verify": false,  // Always verify certificates
    "retry_writes": true,
    "retry_reads": true
};

const client = xk6_mongo.newClientWithOptions(
    'mongodb+srv://cluster.mongodb.net/?retryWrites=true&w=majority',
    clientOptions
);
```

### Connection String Best Practices

```javascript
// ✅ GOOD - Secure connection string with environment variables
const uri = `mongodb+srv://${__ENV.MONGO_USER}:${__ENV.MONGO_PASSWORD}@your-cluster.mongodb.net/?retryWrites=true&tls=true`;

// ❌ BAD - Hardcoded credentials (never do this!)
const uri = 'mongodb://admin:PASSWORD@localhost:27017/?tls=false';
```

## Input Validation

The extension validates inputs, but additional checks are recommended:

```javascript
function sanitizeInput(userInput) {
    // Prevent NoSQL injection
    if (typeof userInput !== 'object') {
        return userInput;
    }

    // Remove dangerous operators
    const dangerous = ['$where', '$function', '$accumulator'];
    for (const key of Object.keys(userInput)) {
        if (dangerous.includes(key)) {
            delete userInput[key];
        }
    }

    return userInput;
}

// Use sanitized input
const filter = sanitizeInput(userProvidedFilter);
const results = client.find("testdb", "collection", filter, null, 100);
```

## Secrets Management

### Store Secrets Securely

**❌ Don't:**
- Commit secrets to version control
- Store in plain text files
- Hardcode in scripts

**✅ Do:**
- Use environment variables
- Use secret management tools (HashiCorp Vault, AWS Secrets Manager)
- Use k6 Cloud secrets

```bash
# .env file (add to .gitignore!)
MONGODB_URI=mongodb+srv://USERNAME:PASSWORD@your-cluster.mongodb.net/
MONGODB_DATABASE=testdb

# Use with k6
k6 run --env MONGODB_URI="$MONGODB_URI" test.js
```

### k6 Cloud Secrets

When using k6 Cloud:

```javascript
// Reference cloud environment variables
const uri = __ENV.MONGODB_URI;  // Configured in k6 Cloud UI
const client = xk6_mongo.newClient(uri);
```

## Audit and Monitoring

### Enable MongoDB Audit Logging

Configure MongoDB to log access:

```yaml
# mongod.conf
security:
  authorization: enabled
auditLog:
  destination: file
  format: JSON
  path: /var/log/mongodb/audit.log
```

### Monitor Suspicious Activity

Watch for:
- Unusual query patterns
- Excessive connection attempts
- Large data exports
- Failed authentication attempts

### Testing Logs

Keep logs of test runs:

```javascript
export function setup() {
    console.log(`Test started at: ${new Date().toISOString()}`);
    console.log(`MongoDB URI: ${__ENV.MONGODB_URI.replace(/:[^:]*@/, ':****@')}`);
}
```

## Compliance Considerations

### GDPR/Privacy Regulations

- Don't store PII in test databases
- Implement data retention policies
- Document data processing activities
- Obtain consent for using real data

### Industry Standards

Follow relevant standards:
- PCI DSS (for payment card data)
- HIPAA (for healthcare data)
- SOC 2 (for service organizations)

## Reporting Security Issues

If you discover a security vulnerability:

1. **DO NOT** open a public issue
2. Email security concerns to: [your-email@example.com]
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will respond within 48 hours and work with you to address the issue.

## Security Checklist

Before running tests:

- [ ] Using TLS/SSL connections
- [ ] Credentials stored securely (environment variables/secrets manager)
- [ ] Test user has minimal required permissions
- [ ] Testing against non-production database
- [ ] No sensitive data in test scripts
- [ ] Rate limiting configured
- [ ] Audit logging enabled
- [ ] Test data cleanup implemented
- [ ] Monitoring in place
- [ ] Team notified of test schedule

## Additional Resources

- [MongoDB Security Checklist](https://docs.mongodb.com/manual/administration/security-checklist/)
- [OWASP NoSQL Injection](https://owasp.org/www-community/attacks/NoSQL_injection)
- [k6 Security Best Practices](https://k6.io/docs/misc/security/)
- [MongoDB Atlas Security](https://docs.atlas.mongodb.com/security/)

---

Last updated: December 2025
