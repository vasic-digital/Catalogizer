# Catalog Web

Modern React web application for Catalogizer media management.

## Tech Stack

- **React 18** with TypeScript
- **Vite** for build tooling
- **TanStack Query** (React Query) for server state
- **Zustand** for client state
- **React Router** for navigation
- **Tailwind CSS** for styling
- **Vitest** for testing (823 tests)

## Quick Start

```bash
# Install dependencies
npm install

# Start development server (http://localhost:5173)
npm run dev

# Run tests
npm run test

# Build for production
npm run build
```

## Available Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server |
| `npm run build` | Build for production |
| `npm run preview` | Preview production build |
| `npm run test` | Run tests (single run) |
| `npm run test:watch` | Run tests in watch mode |
| `npm run test:coverage` | Run tests with coverage |
| `npm run lint` | Run ESLint |
| `npm run lint:fix` | Fix ESLint issues |
| `npm run format` | Format with Prettier |
| `npm run type-check` | TypeScript type checking |

## Project Structure

```
src/
├── components/     # Reusable UI components
├── contexts/       # React contexts (Auth, WebSocket)
├── hooks/          # Custom React hooks
├── lib/            # API clients, utilities
├── pages/          # Route page components
├── styles/         # Global styles
└── types/          # TypeScript type definitions
```

## Environment

The web app connects to the catalog-api backend. Configure the API URL in your environment:

```env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
```

## Testing

Tests are written with Vitest and React Testing Library. Run all 823 tests:

```bash
npm run test
```

## Related Documentation

- [React Frontend Guide](/docs/architecture/REACT_FRONTEND_GUIDE.md)
- [Web App Guide](/docs/guides/WEB_APP_GUIDE.md)
- [API Documentation](/docs/api/API_DOCUMENTATION.md)
