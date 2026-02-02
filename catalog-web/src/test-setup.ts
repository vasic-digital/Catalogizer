import { vi } from 'vitest';
import '@testing-library/jest-dom';

// Make jest.mock() and jest.fn() work in vitest
// Tests written with Jest conventions will work with this shim
(globalThis as any).jest = vi;
