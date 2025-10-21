# Communication System Tests - Quick Reference

## Test Overview

**Total: 39 passing tests across 4 test suites**

### Test Breakdown
- **Matcher Tests** (`matcher_test.go`): 17 tests - Pattern matching and subscription filtering
- **MessageService Tests** (`message_service_test.go`): 6 test suites - Direct messaging operations
- **PubSubService Tests** (`pubsub_service_test.go`): 5 test suites - Pub/sub operations
- **Repository Tests** (`repository_test.go`): 11 test suites - Database integration tests

## Running Tests

### Run All Tests (with ArangoDB)
```bash
ARANGO_PASSWORD=rootpassword go test -v ./internal/communication/
```

### Run Specific Test Suites

**Unit Tests Only (no database required)**
```bash
go test -v -run "TestMatches|TestMessageService|TestPubSubService" ./internal/communication/
```

**Repository Integration Tests (requires ArangoDB)**
```bash
ARANGO_PASSWORD=rootpassword go test -v -run TestRepository ./internal/communication/
```

**Pattern Matcher Tests**
```bash
go test -v -run TestMatches ./internal/communication/
```

**MessageService Tests**
```bash
go test -v -run TestMessageService ./internal/communication/
```

**PubSubService Tests**
```bash
go test -v -run TestPubSubService ./internal/communication/
```

## ArangoDB Setup

### Using Existing Container
```bash
# Start existing ArangoDB container
docker start asset-mgmt-arangodb

# Verify it's running
curl -s -u root:rootpassword http://localhost:8529/_api/version
```

### Environment Variables

Repository tests support the following environment variables:

| Variable          | Default               | Description        |
| ----------------- | --------------------- | ------------------ |
| `ARANGO_HOST`     | `localhost`           | ArangoDB host      |
| `ARANGO_PORT`     | `8529`                | ArangoDB port      |
| `ARANGO_TEST_DB`  | `codeval_cortex_test` | Test database name |
| `ARANGO_USER`     | `root`                | Database username  |
| `ARANGO_PASSWORD` | _(empty)_             | Database password  |

### Custom Configuration Example
```bash
ARANGO_HOST=192.168.1.100 \
ARANGO_PORT=8529 \
ARANGO_PASSWORD=mypassword \
ARANGO_TEST_DB=my_test_db \
go test -v ./internal/communication/
```

## Test Behavior

### Repository Tests
- **Automatically skip** if ArangoDB is not available
- Create test database if it doesn't exist
- Create required collections and indexes automatically
- Clean up test data after each test using `Truncate()`
- Safe to run multiple times

### Unit Tests (Matcher, MessageService, PubSubService)
- Use **mock repositories** - no database required
- Fast execution (< 1 second)
- Isolated - no external dependencies

## Test Coverage

### Messages
- ✅ Create and retrieve messages
- ✅ Get pending messages with filtering (by agent, status, expiration)
- ✅ Update message status and delivery timestamps
- ✅ Acknowledge messages
- ✅ Conversation history by correlation ID
- ✅ Delete expired messages
- ✅ Priority-based ordering

### Publications & Subscriptions
- ✅ Create and retrieve publications
- ✅ Create and retrieve subscriptions
- ✅ Get active subscriptions by agent
- ✅ Deactivate subscriptions
- ✅ Publication delivery tracking
- ✅ Pattern matching (glob-style)
- ✅ Publication type filtering
- ✅ Publisher filtering

### Pattern Matching
- ✅ Exact match
- ✅ Wildcard patterns (`*`, `state.*`, `*.completed`)
- ✅ Multi-criteria subscription matching
- ✅ Publication filtering

## Troubleshooting

### Tests Skip with "ArangoDB not available"
```bash
# Check if ArangoDB is running
docker ps | grep arango

# Start ArangoDB if stopped
docker start asset-mgmt-arangodb

# Wait a few seconds for startup
sleep 3

# Verify connection
curl -u root:rootpassword http://localhost:8529/_api/version
```

### Connection Refused Error
- Ensure ArangoDB container is running
- Check port 8529 is not blocked
- Verify password with `ARANGO_PASSWORD` environment variable

### Test Database Cleanup
```bash
# Drop test database if needed
curl -X DELETE -u root:rootpassword http://localhost:8529/_db/codeval_cortex_test
```

## Quick Test Commands

```bash
# Full test suite with coverage
ARANGO_PASSWORD=rootpassword go test -v -cover ./internal/communication/

# Fast unit tests only
go test -v -short ./internal/communication/

# Specific test
ARANGO_PASSWORD=rootpassword go test -v -run TestRepository_CreateAndGetMessage ./internal/communication/

# List all tests
go test -v -list . ./internal/communication/
```

## CI/CD Integration

For CI environments without ArangoDB:
```bash
# Unit tests will pass, repository tests will skip
go test ./internal/communication/
```

For CI environments with ArangoDB:
```bash
# All tests run
ARANGO_PASSWORD=$DB_PASSWORD go test ./internal/communication/
```
