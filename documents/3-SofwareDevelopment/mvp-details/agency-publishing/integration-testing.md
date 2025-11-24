# MVP-PUB-006: Publishing Integration & Testing

**Domain**: Agency Publishing & Tagging  
**Priority**: P0 (Critical)  
**Effort**: Medium  
**Dependencies**: ~~MVP-PUB-005~~ ✅ (Publishing UI Implementation)  
**Status**: In Progress

## Overview

End-to-end integration and testing of the complete agency publishing and tagging system. This task validates that all publishing components work together correctly, establishes baseline performance metrics, and creates comprehensive documentation for production deployment.

## Objectives

1. **End-to-End Workflow Testing**: Validate complete publish lifecycle from draft to active deployment
2. **Tag System Validation**: Test tag creation, restoration, and comparison workflows
3. **Lifecycle Management**: Verify activation, pause, resume, drain, and stop operations
4. **Performance Benchmarking**: Establish baseline metrics for publication and activation times
5. **Load Testing**: Validate system behavior with multiple concurrent agencies
6. **Documentation**: Create deployment guides and troubleshooting documentation
7. **Training Materials**: Develop user guides for agency publishing workflows

## Requirements

### Functional Requirements

#### FR-1: End-to-End Publish Workflow
- **FR-1.1**: Draft agency can be validated successfully
- **FR-1.2**: Validated agency can be published with version number
- **FR-1.3**: Published agency can be activated (agents spawn, workflows initialize)
- **FR-1.4**: Active agency can accept work and execute workflows
- **FR-1.5**: Active agency can be paused and resumed
- **FR-1.6**: Active agency can be drained gracefully
- **FR-1.7**: Active agency can be stopped (graceful or forced)
- **FR-1.8**: Stopped agency can be reactivated from publication

#### FR-2: Tag System Validation
- **FR-2.1**: Tag can be created from any agency state (except draft)
- **FR-2.2**: Tag types (Release, Snapshot, Experimental, Checkpoint) work correctly
- **FR-2.3**: Tag restoration reverts agency to exact tagged state
- **FR-2.4**: Tag comparison shows differences between two snapshots
- **FR-2.5**: Tag deletion removes tag but preserves agency
- **FR-2.6**: Tag-based publishing creates new publication from tag

#### FR-3: Validation System
- **FR-3.1**: Incomplete agencies fail validation with specific errors
- **FR-3.2**: Validation checks all required sections (introduction, goals, roles, workflows, RACI)
- **FR-3.3**: Validation prevents publishing of invalid agencies
- **FR-3.4**: Validation errors are user-friendly and actionable

#### FR-4: State Machine Integrity
- **FR-4.1**: State transitions follow defined rules (no invalid transitions)
- **FR-4.2**: Guards prevent unauthorized transitions
- **FR-4.3**: State changes trigger appropriate actions
- **FR-4.4**: Concurrent state transitions are handled safely

### Non-Functional Requirements

#### NFR-1: Performance
- **NFR-1.1**: Validation completes in < 2 seconds
- **NFR-1.2**: Publication creation completes in < 3 seconds
- **NFR-1.3**: Agent spawning completes in < 10 seconds (for 10 agents)
- **NFR-1.4**: Tag creation completes in < 1 second
- **NFR-1.5**: Tag restoration completes in < 2 seconds

#### NFR-2: Scalability
- **NFR-2.1**: System supports 50+ concurrent agencies
- **NFR-2.2**: System supports 100+ publications per agency
- **NFR-2.3**: System supports 50+ tags per agency
- **NFR-2.4**: Database queries remain performant (< 100ms p95)

#### NFR-3: Reliability
- **NFR-3.1**: Publication creation is atomic (all-or-nothing)
- **NFR-3.2**: Activation failures don't corrupt agency state
- **NFR-3.3**: Failed agent spawns are properly logged
- **NFR-3.4**: System recovers gracefully from database failures

#### NFR-4: Observability
- **NFR-4.1**: All lifecycle events are logged
- **NFR-4.2**: Metrics are exported for monitoring
- **NFR-4.3**: Audit trail captures all state changes
- **NFR-4.4**: Errors include actionable debugging information

## Test Scenarios

### Scenario 1: Happy Path - Draft to Active

**Steps**:
1. Create new agency (state: Draft)
2. Add introduction, goals, roles, work items, workflows, RACI matrix
3. Click "Validate" → state transitions to Validated
4. Click "Publish" with version v1.0.0 → state transitions to Published
5. Click "Activate" → agents spawn, workflows initialize, state transitions to Active
6. Verify agents are running and accepting work

**Expected Results**:
- All state transitions succeed
- Publication record created with deployment manifest
- Agents spawned match role definitions
- Workflows initialized correctly
- Audit trail captures all events

### Scenario 2: Tag Creation and Restoration

**Steps**:
1. Start with Active agency (from Scenario 1)
2. Click "Create Tag" → create "v1.0.0-stable" (Release tag)
3. Modify agency configuration (add new role)
4. Verify modified state
5. Click "Restore from Tag" → select "v1.0.0-stable"
6. Verify agency reverted to original configuration

**Expected Results**:
- Tag created successfully
- Tag snapshot captures complete agency state
- Restoration reverts all changes
- Original agents/workflows preserved

### Scenario 3: Lifecycle Management

**Steps**:
1. Start with Active agency
2. Click "Pause" → verify agents stop accepting work
3. Verify existing work continues
4. Click "Resume" → verify agents resume accepting work
5. Click "Drain" → verify no new work accepted
6. Wait for drain completion → state transitions to Stopped
7. Click "Activate" → verify agents respawn

**Expected Results**:
- Pause prevents new work, existing work completes
- Resume restores full functionality
- Drain completes all pending work gracefully
- Stop cleans up all resources
- Reactivation spawns fresh agents

### Scenario 4: Validation Failures

**Steps**:
1. Create agency with incomplete configuration
2. Missing introduction → validate fails
3. Missing goals → validate fails
4. Missing roles → validate fails
5. Missing RACI matrix → validate fails
6. Add all required sections
7. Validate succeeds

**Expected Results**:
- Each validation failure shows specific error
- UI displays validation errors clearly
- Cannot publish until validation passes

### Scenario 5: Concurrent Operations

**Steps**:
1. Create 10 agencies
2. Validate all 10 concurrently
3. Publish all 10 concurrently with different versions
4. Activate all 10 concurrently
5. Monitor system resource usage
6. Verify all operations complete successfully

**Expected Results**:
- No race conditions or deadlocks
- All agencies transition correctly
- Database maintains consistency
- Performance remains acceptable

## Acceptance Criteria

- [ ] **AC-1**: All test scenarios pass successfully
- [ ] **AC-2**: Performance benchmarks meet NFR requirements
- [ ] **AC-3**: Load testing validates 50+ concurrent agencies
- [ ] **AC-4**: No critical bugs or security issues identified
- [ ] **AC-5**: Deployment documentation complete and tested
- [ ] **AC-6**: User guide created with screenshots and workflows
- [ ] **AC-7**: Troubleshooting guide created with common issues
- [ ] **AC-8**: Monitoring dashboards configured for production
- [ ] **AC-9**: Audit trail validates all expected events logged
- [ ] **AC-10**: Error messages are user-friendly and actionable

## Technical Specifications

### Test Infrastructure

**Manual Testing**:
- Use Agency Designer UI for end-to-end workflows
- Create test agencies with realistic configurations
- Document steps and results with screenshots

**API Testing** (optional for future):
- Postman collection for all publish/tag/lifecycle endpoints
- Automated API test suite
- Load testing with k6 or similar tool

**Database Validation**:
- Query collections to verify data integrity
- Check indexes are used efficiently
- Validate audit trail completeness

### Monitoring Setup

**Metrics to Track**:
- `agency_publications_total` - Counter of publications created
- `agency_activations_total` - Counter of activations
- `agency_activation_duration_seconds` - Histogram of activation times
- `agency_agents_spawned_total` - Counter of agents created
- `agency_tags_created_total` - Counter of tags
- `agencies_by_state` - Gauge showing distribution

**Dashboards**:
- Agency lifecycle overview (state distribution, transitions)
- Publication performance (publish/activate times)
- Agent spawning success rate
- Tag usage patterns

### Documentation Deliverables

**1. Deployment Guide** (`deployment-guide.md`):
- System requirements
- Database setup (collections, indexes)
- Configuration parameters
- Environment variables
- Initial setup steps
- Health check procedures

**2. User Guide** (`user-guide.md`):
- Publishing workflow walkthrough
- Tag creation and management
- Lifecycle operations (pause/resume/drain/stop)
- Best practices for versioning
- Common use cases with examples

**3. Troubleshooting Guide** (`troubleshooting.md`):
- Common errors and solutions
- Validation failure resolution
- Activation failure debugging
- Performance optimization tips
- Database query debugging

**4. Training Materials**:
- Step-by-step tutorial for first publish
- Video walkthrough (optional)
- FAQ document
- Quick reference card

## Implementation Plan

### Phase 1: Test Scenario Execution (2-3 hours)
1. Set up test environment with clean database
2. Execute Scenario 1 (Happy Path) and document results
3. Execute Scenario 2 (Tags) and validate restoration
4. Execute Scenario 3 (Lifecycle) and verify transitions
5. Execute Scenario 4 (Validation) and check error messages
6. Execute Scenario 5 (Concurrent) and measure performance

### Phase 2: Performance Benchmarking (1-2 hours)
1. Measure validation time (10+ agencies)
2. Measure publication creation time (10+ agencies)
3. Measure activation time (10+ agencies, various sizes)
4. Measure tag operations (create/restore/compare)
5. Document baseline metrics

### Phase 3: Documentation Creation (2-3 hours)
1. Write deployment guide
2. Write user guide with screenshots
3. Write troubleshooting guide
4. Create quick reference materials

### Phase 4: Final Validation (1 hour)
1. Review all acceptance criteria
2. Fix any identified issues
3. Re-run critical test scenarios
4. Sign off on completion

## Success Metrics

- **Test Coverage**: All 5 scenarios executed successfully
- **Performance**: All NFRs met or exceeded
- **Documentation**: 3+ guides created and reviewed
- **Stability**: Zero critical bugs in production path
- **Usability**: User guide validated by non-developer

## Known Limitations

- Manual testing only (no automated test suite)
- Load testing limited by development environment
- Monitoring dashboards not implemented (metrics collection only)
- No automated performance regression testing

## Future Enhancements

- Automated integration test suite
- Performance regression testing in CI/CD
- Canary deployment testing
- Blue-green deployment support
- A/B testing framework for versioned agencies

## Related Documentation

- [Architecture Spec](../../../2-SoftwareDesignAndArchitecture/agency-publishing-tagging-architecture.md)
- [State Machine Details](state-machine.md)
- [MVP-PUB-001: State Machine & Data Models](state-machine.md#mvp-pub-001)
- [Coding Sessions](../../coding_sessions/) - Implementation logs for MVP-PUB-001 through MVP-PUB-005
