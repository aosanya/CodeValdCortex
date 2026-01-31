# Flutter Migration Plan - CodeValdFortex Frontend

**Date**: January 31, 2026  
**Status**: Active Implementation  
**Author**: Architecture Team  
**Last Updated**: January 31, 2026

---

## Executive Summary

This document outlines the migration strategy from Templ+HTMX+Alpine.js to **Flutter cross-platform application (CodeValdFortex)**. This migration provides separation of concerns, multi-platform support, and modern reactive state management.

### Problem Statement
- **Current Issue**: Templ templates entangled with Go business logic
- **Backend Bloat**: Presentation layer mixed with domain logic
- **Platform Limitations**: Web-only, no native mobile/desktop
- **State Management**: Alpine.js insufficient for complex apps

### Solution Overview
- **Pure REST API Backend**: CodeValdCortex (Go + Gin)
- **Flutter Frontend**: CodeValdFortex (Web, iOS, Android, Desktop)
- **MVVM Architecture**: Riverpod for reactive state management
- **Incremental Migration**: Domain-by-domain rollout

---

## Architecture Overview

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Development Container                      │
│  /workspaces/                                              │
│  ├── CodeValdCortex/          (Go Backend - REST API)     │
│  │   ├── internal/api/        (REST endpoints)            │
│  │   ├── internal/agency/     (Business logic)            │
│  │   └── internal/agent/      (Runtime)                   │
│  │                                                          │
│  └── CodeValdFortex/          (Flutter Multi-Platform)    │
│      ├── lib/                                             │
│      │   ├── features/        (Feature modules)           │
│      │   ├── models/          (Data models)               │
│      │   ├── services/        (API clients)               │
│      │   ├── viewmodels/      (Riverpod providers)        │
│      │   └── widgets/         (Reusable UI)               │
│      └── pubspec.yaml                                      │
└─────────────────────────────────────────────────────────────┘

Runtime:
┌──────────────┐     HTTP/REST    ┌──────────────┐
│ Flutter App  │ ◄────────────────►│ Cortex API   │
│ (Fortex)     │  /api/v1/*        │ (Go Backend) │
│ Web/Mobile/  │                   │  Port: 8080  │
│  Desktop     │                   └──────┬───────┘
└──────────────┘                          │
                                          ▼
                                   ┌──────────────┐
                                   │  ArangoDB    │
                                   └──────────────┘
```

---

## Flutter Technology Stack

### Core Technologies
- **Language**: Dart 3.x
- **Framework**: Flutter 3.x (stable channel)
- **State Management**: Riverpod 2.x
- **HTTP Client**: Dio 5.x
- **Routing**: go_router 13.x
- **Storage**: flutter_secure_storage (tokens/credentials)

### UI & Design
- **Design System**: Material Design 3
- **Responsive**: Layout Builder with breakpoints
- **Charts**: fl_chart for data visualization
- **Icons**: Material Icons + FontAwesome

### Development Tools
- **Linting**: flutter_lints (strict mode)
- **Testing**: flutter_test + mockito
- **Code Generation**: freezed + json_serializable

---

## Migration Status

### ✅ Completed (Foundation)
- MVP-FL-001: Flutter project setup
- MVP-FL-002: Design system (Material Design)
- MVP-FL-003: Routing (go_router)
- MVP-FL-004: State management (Riverpod)
- MVP-FL-005: API client layer (Dio)
- MVP-FL-009: Authentication state management
- MVP-FL-010: Login/registration screens
- MVP-FL-011: Protected routes & permissions
- MVP-FL-101: Agency selection homepage
- MVP-FL-102: Create agency form
- MVP-FL-103: Agency Designer navigation shell

### 🔄 In Progress (Agency Designer)
- MVP-FL-104: Introduction section
- MVP-FL-105: Goals section
- MVP-FL-106: Work Items section
- MVP-FL-107: Roles section
- MVP-FL-108: RACI matrix
- MVP-FL-109: Workflows section
- MVP-FL-110: AI Policy section
- MVP-FL-111: Admin & Configuration

### 📋 Pending (Publishing & Instances)
- MVP-FL-112 through MVP-FL-116: Publishing system
- MVP-FL-117 through MVP-FL-119: Instance management
- MVP-FL-120 through MVP-FL-122: Work items (Kanban, Issue detail, File explorer)

---

## REST API Specification

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
- **Method**: JWT tokens
- **Storage**: flutter_secure_storage
- **Header**: `Authorization: Bearer <token>`
- **Refresh**: Auto-refresh on 401 responses

### Example Endpoints

#### Authentication
```http
POST /api/v1/auth/login
POST /api/v1/auth/register
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
```

#### Agencies
```http
GET    /api/v1/agencies
POST   /api/v1/agencies
GET    /api/v1/agencies/:id
PUT    /api/v1/agencies/:id
DELETE /api/v1/agencies/:id
```

#### Work Items
```http
GET    /api/v1/agencies/:id/work-items
POST   /api/v1/agencies/:id/work-items
GET    /api/v1/agencies/:id/work-items/:workItemId
PUT    /api/v1/agencies/:id/work-items/:workItemId
DELETE /api/v1/agencies/:id/work-items/:workItemId
```

### Response Format

**Success (200 OK):**
```json
{
  "data": { /* resource data */ },
  "meta": {
    "timestamp": "2026-01-31T10:00:00Z"
  }
}
```

**Error (4xx/5xx):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Title is required",
    "details": {
      "field": "title"
    }
  }
}
```

---

## Flutter Project Structure

```
/workspaces/CodeValdFortex/
├── lib/
│   ├── main.dart                      # Entry point
│   │
│   ├── core/                          # Core utilities
│   │   ├── constants/
│   │   ├── theme/
│   │   ├── router/
│   │   └── utils/
│   │
│   ├── models/                        # Data models
│   │   ├── agency.dart
│   │   ├── work_item.dart
│   │   ├── user.dart
│   │   └── *.freezed.dart            # Generated
│   │
│   ├── services/                      # API services
│   │   ├── api_client.dart            # Dio instance
│   │   ├── auth_service.dart
│   │   ├── agency_service.dart
│   │   └── work_items_service.dart
│   │
│   ├── providers/                     # Riverpod providers
│   │   ├── auth_provider.dart
│   │   ├── agency_provider.dart
│   │   └── work_items_provider.dart
│   │
│   ├── features/                      # Feature modules
│   │   ├── auth/
│   │   │   ├── views/
│   │   │   ├── viewmodels/
│   │   │   └── widgets/
│   │   │
│   │   ├── agencies/
│   │   │   ├── views/
│   │   │   │   ├── agency_list_view.dart
│   │   │   │   └── create_agency_view.dart
│   │   │   ├── viewmodels/
│   │   │   │   └── agency_viewmodel.dart
│   │   │   └── widgets/
│   │   │       └── agency_card.dart
│   │   │
│   │   ├── agency_designer/
│   │   │   ├── views/
│   │   │   │   ├── designer_shell.dart
│   │   │   │   ├── introduction_view.dart
│   │   │   │   ├── goals_view.dart
│   │   │   │   └── ...
│   │   │   ├── viewmodels/
│   │   │   └── widgets/
│   │   │
│   │   └── work_items/
│   │       ├── views/
│   │       │   ├── kanban_board.dart
│   │       │   └── issue_detail.dart
│   │       ├── viewmodels/
│   │       └── widgets/
│   │
│   └── widgets/                       # Shared widgets
│       ├── buttons/
│       ├── cards/
│       ├── forms/
│       └── dialogs/
│
├── test/                              # Tests
│   ├── unit/
│   ├── widget/
│   └── integration/
│
├── assets/                            # Assets
│   ├── images/
│   ├── fonts/
│   └── icons/
│
├── web/                               # Web platform
├── ios/                               # iOS platform
├── android/                           # Android platform
├── macos/                             # macOS platform
├── linux/                             # Linux platform
├── windows/                           # Windows platform
│
├── pubspec.yaml                       # Dependencies
└── analysis_options.yaml              # Lint rules
```

---

## MVVM Architecture Pattern

### ViewModel Example (Riverpod)

```dart
// providers/work_items_provider.dart
import 'package:riverpod_annotation/riverpod_annotation.dart';
import '../models/work_item.dart';
import '../services/work_items_service.dart';

part 'work_items_provider.g.dart';

@riverpod
class WorkItems extends _$WorkItems {
  @override
  Future<List<WorkItem>> build(String agencyId) async {
    final service = ref.read(workItemsServiceProvider);
    return service.listWorkItems(agencyId);
  }

  Future<void> createWorkItem(WorkItem item) async {
    final service = ref.read(workItemsServiceProvider);
    await service.createWorkItem(item);
    ref.invalidateSelf();
  }

  Future<void> updateWorkItem(String id, WorkItem item) async {
    final service = ref.read(workItemsServiceProvider);
    await service.updateWorkItem(id, item);
    ref.invalidateSelf();
  }

  Future<void> deleteWorkItem(String id) async {
    final service = ref.read(workItemsServiceProvider);
    await service.deleteWorkItem(id);
    ref.invalidateSelf();
  }
}
```

### View Example

```dart
// features/work_items/views/work_items_view.dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class WorkItemsView extends ConsumerWidget {
  final String agencyId;

  const WorkItemsView({required this.agencyId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final workItemsAsync = ref.watch(workItemsProvider(agencyId));

    return Scaffold(
      appBar: AppBar(title: Text('Work Items')),
      body: workItemsAsync.when(
        data: (items) => ListView.builder(
          itemCount: items.length,
          itemBuilder: (context, index) {
            final item = items[index];
            return WorkItemCard(item: item);
          },
        ),
        loading: () => Center(child: CircularProgressIndicator()),
        error: (error, stack) => Center(
          child: Text('Error: $error'),
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showCreateDialog(context, ref),
        child: Icon(Icons.add),
      ),
    );
  }

  void _showCreateDialog(BuildContext context, WidgetRef ref) {
    // Show create work item dialog
  }
}
```

---

## Development Workflow

### Running Applications

**Terminal 1 - Go Backend:**
```bash
cd /workspaces/CodeValdCortex
make run
# Runs on http://localhost:8080
```

**Terminal 2 - Flutter Web:**
```bash
cd /workspaces/CodeValdFortex
flutter run -d chrome
# Runs on http://localhost:<random_port>
```

**Flutter Mobile (with device/emulator):**
```bash
flutter run -d <device_id>
# Lists devices: flutter devices
```

### Code Generation

```bash
# Generate freezed/json_serializable code
flutter pub run build_runner build --delete-conflicting-outputs

# Watch mode (auto-regenerate)
flutter pub run build_runner watch
```

### Testing

```bash
# Unit tests
flutter test

# Widget tests
flutter test test/widget/

# Integration tests
flutter test integration_test/

# Coverage
flutter test --coverage
```

---

## Migration Timeline

### Phase 1: Foundation ✅ COMPLETE
**Duration**: Weeks 1-4 (Completed)
- Project setup & dependencies
- Design system & theming
- Routing & navigation
- State management foundation
- API client layer
- Authentication system

### Phase 2: Agency Management ✅ COMPLETE  
**Duration**: Weeks 5-6 (Completed)
- Agency list/selection
- Create agency form
- Agency Designer shell with tab navigation

### Phase 3: Agency Designer (Current) 🔄
**Duration**: Weeks 7-12 (In Progress)
- Introduction section (MVP-FL-104)
- Goals CRUD (MVP-FL-105)
- Work Items with deliverables tree (MVP-FL-106)
- Roles management (MVP-FL-107)
- RACI matrix editor (MVP-FL-108)
- Workflows visual designer (MVP-FL-109)
- AI Policy configuration (MVP-FL-110)
- Admin & configuration (MVP-FL-111)

### Phase 4: Publishing System
**Duration**: Weeks 13-16
- Publishing toolbar & validation (MVP-FL-112)
- Publish & activate dialogs (MVP-FL-113)
- Tag management UI (MVP-FL-114)
- Versions page (MVP-FL-115)
- Export system (MVP-FL-116)

### Phase 5: Instance Management
**Duration**: Weeks 17-19
- Instance creation UI (MVP-FL-117)
- Instance dashboard (MVP-FL-118)
- Lifecycle controls (MVP-FL-119)

### Phase 6: Work Items (Kanban)
**Duration**: Weeks 20-22
- Kanban board (MVP-FL-120)
- Issue detail panel (MVP-FL-121)
- File explorer (MVP-FL-122)

### Phase 7: Cleanup & Production
**Duration**: Weeks 23-24
- Remove Templ templates from Cortex (MVP-CLEANUP-001 through 014)
- Performance optimization
- Production deployment
- Monitoring setup

---

## Deployment Strategy

### Web Deployment
```bash
# Build for web
flutter build web --release

# Output: build/web/
# Deploy to: Vercel, Netlify, Firebase Hosting, or Nginx
```

### Mobile Deployment

**iOS (App Store):**
```bash
flutter build ios --release
# Code signing in Xcode
# Submit via App Store Connect
```

**Android (Play Store):**
```bash
flutter build apk --release
flutter build appbundle --release
# Submit via Google Play Console
```

### Desktop Deployment

**macOS:**
```bash
flutter build macos --release
```

**Windows:**
```bash
flutter build windows --release
```

**Linux:**
```bash
flutter build linux --release
```

---

## Success Criteria

### Performance
- [ ] Time to Interactive <2s (web)
- [ ] 60 FPS animations
- [ ] Bundle size <5MB (web gzipped)
- [ ] Memory usage <100MB (mobile)

### Quality
- [ ] Zero critical bugs
- [ ] Test coverage >80%
- [ ] Accessibility (screen readers, keyboard nav)
- [ ] Responsive on all screen sizes

### User Experience
- [ ] Consistent Material Design 3
- [ ] Smooth transitions
- [ ] Offline support (future)
- [ ] Multi-language support (future)

---

## Technology Stack Summary

| Category | Technology | Version | Purpose |
|----------|-----------|---------|---------|
| **Language** | Dart | 3.x | Type-safe language |
| **Framework** | Flutter | 3.x | UI framework |
| **State Management** | Riverpod | 2.x | Reactive state |
| **HTTP Client** | Dio | 5.x | API calls |
| **Routing** | go_router | 13.x | Navigation |
| **Storage** | flutter_secure_storage | 9.x | Secure storage |
| **Charts** | fl_chart | Latest | Data visualization |
| **Code Gen** | freezed | 2.x | Immutable models |
| **Code Gen** | json_serializable | 6.x | JSON parsing |
| **Testing** | flutter_test | Built-in | Unit/widget tests |
| **Testing** | mockito | 5.x | Mocking |

---

## Resources

- **Flutter Docs**: https://flutter.dev/docs
- **Riverpod Docs**: https://riverpod.dev
- **Material Design 3**: https://m3.material.io
- **Dio Package**: https://pub.dev/packages/dio
- **go_router**: https://pub.dev/packages/go_router
- **Fortex Architecture**: `/workspaces/CodeValdFortex/documents/2-SoftwareDesignAndArchitecture/`

---

**Status**: Active Implementation  
**Next Milestone**: Complete Agency Designer sections (Phase 3)  
**Review Date**: Weekly progress updates

