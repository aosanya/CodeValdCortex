# CodeValdCortex - Frontend Architecture

## Current Status: Deprecated

**Date**: January 31, 2026

**Reason**: The original React-based frontend architecture described in this document has been superseded by the Flutter-based frontend in CodeValdFortex.

## Architecture Evolution

### Previous Approach (Deprecated)
- **React + TypeScript** enterprise dashboard
- Full frontend implementation in CodeValdCortex
- See archived document: [archive/frontend-architecture-react-deprecated.md](archive/frontend-architecture-react-deprecated.md)

### Current Approach (Active)
- **Backend**: CodeValdCortex is now a **pure REST API backend** (Go + Gin)
- **Frontend**: CodeValdFortex provides the **Flutter-based UI** (cross-platform: Web, iOS, Android, Desktop)
- **Temporary UI**: Templ + HTMX templates exist for development/testing but will be removed (see MVP-CLEANUP-001 through MVP-CLEANUP-014)

## Current Frontend Strategy

### CodeValdCortex (Backend)
**Role**: REST API server only
- **Technology**: Go + Gin framework
- **Responsibilities**:
  - RESTful API endpoints (`/api/v1/*`)
  - Business logic and data persistence
  - Authentication/authorization (JWT)
  - Agent runtime and orchestration
  - Database operations (ArangoDB)
- **Temporary UI**: Templ templates for Agency Designer (to be removed after Flutter migration)

### CodeValdFortex (Frontend)
**Role**: Primary user interface
- **Technology**: Flutter (Dart)
- **Platforms**: Web, iOS, Android, Desktop
- **Architecture**: MVVM with Riverpod state management
- **API Client**: Dio HTTP client to CodeValdCortex REST API
- **See**: `/workspaces/CodeValdFortex/documents/2-SoftwareDesignAndArchitecture/`

## Migration Status

**Completed**:
- ✅ Flutter project setup (MVP-FL-001 through MVP-FL-005)
- ✅ Authentication system (MVP-FL-009 through MVP-FL-011)
- ✅ Agency management (MVP-FL-101, MVP-FL-102)
- ✅ Agency Designer navigation (MVP-FL-103)

**In Progress**:
- 🔄 Agency Designer sections (MVP-FL-104 through MVP-FL-111)
- 🔄 Publishing system (MVP-FL-112 through MVP-FL-116)
- 🔄 Instance management (MVP-FL-117 through MVP-FL-119)
- 🔄 Work items (MVP-FL-120 through MVP-FL-122)

**Pending Cleanup**:
- 📋 Remove Templ templates from Cortex after Flutter reaches feature parity (MVP-CLEANUP-001 through MVP-CLEANUP-014)

## References

- **Active Frontend Docs**: `/workspaces/CodeValdFortex/documents/2-SoftwareDesignAndArchitecture/`
- **Backend API Docs**: [backend-architecture.md](backend-architecture.md)
- **Migration Tracking**: `/workspaces/CodeValdCortex/documents/3-SofwareDevelopment/mvp.md` (P2: UI Migration & Cleanup section)
- **Archived React Architecture**: [archive/frontend-architecture-react-deprecated.md](archive/frontend-architecture-react-deprecated.md)
