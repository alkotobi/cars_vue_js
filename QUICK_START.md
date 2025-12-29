# Quick Start: Monorepo with Shared Code

## ✅ Setup Complete!

Your monorepo is ready. Here's how to use it:

## 1. Install Dependencies

```bash
npm install
```

This installs dependencies for the root, shared package, and sub-app.

## 2. Run Apps

### Main App (existing)
```bash
npm run dev
# Runs on http://localhost:5173
```

### Sub-App (new)
```bash
npm run dev:sub-app
# Runs on http://localhost:5174
```

## 3. Using Shared Code

### In Main App (`src/`)

You can continue using existing imports OR switch to shared:

```javascript
// Option 1: Keep existing (still works)
import { useApi } from '../composables/useApi'

// Option 2: Use shared package (recommended for new code)
import { useApi } from '@cars/shared'
```

### In Sub-App (`packages/sub-app/src/`)

Always use shared package:

```javascript
import { useApi, useEnhancedI18n } from '@cars/shared'
import { useApi } from '@cars/shared/composables'
```

## 4. What's Shared?

Currently shared:
- ✅ All composables (`useApi`, `useI18n`, `useCarClientChat`, etc.)
- ✅ All utils (`i18nDevHelper`, etc.)

You can add more:
- Shared components
- Shared types/interfaces
- Shared constants
- Shared utilities

## 5. Project Structure

```
cars/
├── packages/
│   ├── shared/              # Shared code
│   │   └── src/
│   │       ├── composables/  # useApi, useI18n, etc.
│   │       ├── utils/        # Utilities
│   │       └── components/   # Shared components (add here)
│   └── sub-app/             # Your new sub app
│       └── src/
├── src/                      # Main app (existing)
└── package.json             # Root with workspaces
```

## 6. Adding New Shared Code

1. Add file to `packages/shared/src/composables/` (or appropriate folder)
2. Export it in `packages/shared/src/composables/index.js`
3. Use it in any app: `import { myFunction } from '@cars/shared'`

## Troubleshooting

### "Cannot find module '@cars/shared'"

Run `npm install` from the root directory to link workspaces.

### Changes not reflecting

- Restart dev server
- Check that file is exported in the appropriate `index.js`

### Port conflicts

Sub-app uses port 5174. Change it in `packages/sub-app/vite.config.js` if needed.

