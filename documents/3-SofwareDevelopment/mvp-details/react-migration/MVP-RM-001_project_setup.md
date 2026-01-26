# MVP-RM-001: CodeValdFortex Project Setup

**Domain**: React Migration  
**Priority**: P1  
**Effort**: Low  
**Skills**: Frontend Dev, TypeScript, Build Tools  
**Dependencies**: None

---

## Overview

Initialize the CodeValdFortex repository as a new React + TypeScript + Vite project within the development container workspace. This establishes the foundation for the React-based SPA frontend.

---

## Requirements

### 1. Repository Creation
- Create `/workspaces/CodeValdFortex/` directory in the dev container
- Initialize as new Vite project with React + TypeScript template
- Configure Git repository with proper `.gitignore`

### 2. Dependency Installation
**Core Dependencies:**
```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "@reduxjs/toolkit": "^2.0.0",
    "react-redux": "^9.0.0",
    "react-router-dom": "^6.20.0",
    "axios": "^1.6.0",
    "bulma": "^1.0.2"
  },
  "devDependencies": {
    "@vitejs/plugin-react": "^4.2.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0",
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "eslint": "^8.55.0",
    "prettier": "^3.1.0",
    "vitest": "^1.0.0",
    "@testing-library/react": "^14.1.0"
  }
}
```

### 3. Project Structure
```
/workspaces/CodeValdFortex/
├── public/
│   └── favicon.ico
├── src/
│   ├── api/                # API client layer
│   │   └── client.ts
│   ├── app/                # App-level setup
│   │   ├── App.tsx
│   │   ├── store.ts
│   │   └── routes.tsx
│   ├── features/           # Feature-based organization
│   ├── shared/             # Shared utilities
│   │   ├── components/
│   │   ├── hooks/
│   │   └── utils/
│   ├── styles/             # Global styles
│   │   ├── bulma.scss
│   │   └── custom.scss
│   ├── main.tsx            # Entry point
│   └── vite-env.d.ts
├── .env.development
├── .env.production
├── .gitignore
├── index.html
├── package.json
├── tsconfig.json
├── tsconfig.node.json
├── vite.config.ts
└── README.md
```

### 4. Configuration Files

**TypeScript Config (`tsconfig.json`):**
```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    }
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

**Vite Config (`vite.config.ts`):**
```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
```

**Environment Variables:**
```bash
# .env.development
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_NAME=CodeValdFortex

# .env.production
VITE_API_BASE_URL=https://api.codevaldcortex.com
VITE_APP_NAME=CodeValdFortex
```

### 5. Basic App Structure

**Entry Point (`src/main.tsx`):**
```tsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import { Provider } from 'react-redux';
import { BrowserRouter } from 'react-router-dom';
import App from './app/App';
import { store } from './app/store';
import './styles/bulma.scss';
import './styles/custom.scss';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Provider store={store}>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </Provider>
  </React.StrictMode>
);
```

**App Component (`src/app/App.tsx`):**
```tsx
import React from 'react';
import { Routes, Route } from 'react-router-dom';

function App() {
  return (
    <div className="app">
      <h1>CodeValdFortex</h1>
      <p>React frontend is running!</p>
    </div>
  );
}

export default App;
```

**Redux Store (`src/app/store.ts`):**
```tsx
import { configureStore } from '@reduxjs/toolkit';

export const store = configureStore({
  reducer: {
    // Reducers will be added in subsequent tasks
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

**API Client (`src/api/client.ts`):**
```typescript
import axios from 'axios';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // For session cookies
});

export default apiClient;
```

### 6. Code Quality Tools

**ESLint Config (`.eslintrc.json`):**
```json
{
  "extends": [
    "eslint:recommended",
    "plugin:react/recommended",
    "plugin:@typescript-eslint/recommended"
  ],
  "parser": "@typescript-eslint/parser",
  "plugins": ["react", "@typescript-eslint"],
  "rules": {
    "react/react-in-jsx-scope": "off"
  }
}
```

**Prettier Config (`.prettierrc`):**
```json
{
  "semi": true,
  "trailingComma": "es5",
  "singleQuote": true,
  "printWidth": 100,
  "tabWidth": 2
}
```

### 7. Bulma CSS Setup

**Bulma Import (`src/styles/bulma.scss`):**
```scss
@charset "utf-8";

// Import only what you need from Bulma
@import "bulma/bulma.sass";
```

**Custom Styles (`src/styles/custom.scss`):**
```scss
// Custom theme overrides
:root {
  --primary-color: #3273dc;
  --dark-bg: #1a1a1a;
}

.app {
  min-height: 100vh;
}
```

---

## Acceptance Criteria

- [ ] CodeValdFortex repository created in `/workspaces/CodeValdFortex/`
- [ ] All dependencies installed successfully (`npm install` completes)
- [ ] TypeScript compiles without errors (`tsc --noEmit` passes)
- [ ] Development server runs (`npm run dev` works)
- [ ] Vite dev server accessible at `http://localhost:5173`
- [ ] Basic "Hello World" React app renders
- [ ] ESLint and Prettier configured and working
- [ ] Bulma CSS loads correctly
- [ ] Environment variables work (can access `import.meta.env.VITE_API_BASE_URL`)
- [ ] Redux store configured (even if empty)
- [ ] Axios client configured with proper base URL
- [ ] Git repository initialized with proper `.gitignore`
- [ ] README.md contains project setup instructions
- [ ] Can run build command (`npm run build` succeeds)

---

## Testing

### Manual Testing
1. Run `npm run dev` - server starts on port 5173
2. Open `http://localhost:5173` - app loads
3. Check browser console - no errors
4. Verify Bulma CSS is applied (inspect styles)

### Automated Testing
```bash
# TypeScript compilation
npm run tsc

# Linting
npm run lint

# Build
npm run build
```

---

## Implementation Notes

### File Size Guidelines
- Keep configuration files concise
- Use comments to explain non-obvious settings
- Split large configs into separate files if needed

### Development Workflow
```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Run linter
npm run lint

# Format code
npm run format
```

### Multi-Repo Development
- CodeValdFortex runs on port 5173
- CodeValdCortex runs on port 8080
- Vite proxy forwards `/api` requests to Go backend
- Both projects run simultaneously in dev container

---

## Dependencies

**None** - This is the foundation task

---

## Related Tasks

- **MVP-RM-002**: Backend API Infrastructure (creates `/api/v1` endpoints)
- **MVP-RM-004**: Work Items Redux Store (first feature slice)
- **MVP-RM-005**: Work Items UI Components (first React components)

---

## References

- [Vite Documentation](https://vitejs.dev/)
- [React + TypeScript Guide](https://react-typescript-cheatsheet.netlify.app/)
- [Redux Toolkit Documentation](https://redux-toolkit.js.org/)
- [Bulma Documentation](https://bulma.io/)

---

**Created**: January 26, 2026  
**Last Updated**: January 26, 2026
