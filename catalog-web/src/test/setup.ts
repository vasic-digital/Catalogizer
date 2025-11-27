import '@testing-library/jest-dom';

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(), // deprecated
    removeListener: jest.fn(), // deprecated
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

// Mock IntersectionObserver
global.IntersectionObserver = class IntersectionObserver {
  root = null;
  rootMargin = '';
  thresholds: ReadonlyArray<number> = [];

  constructor() {/* mock */}
  observe() {
    return null;
  }
  disconnect() {
    return null;
  }
  unobserve() {
    return null;
  }
} as any;

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  constructor() {/* mock */}
  observe() {
    return null;
  }
  disconnect() {
    return null;
  }
  unobserve() {
    return null;
  }
} as any;

// Mock WebSocket properly for TypeScript
class WebSocketMock {
  static CONNECTING = 0;
  static OPEN = 1;
  static CLOSING = 2;
  static CLOSED = 3;
  
  url: string;
  readyState: number;
  onopen: ((event: Event) => void) | null;
  onmessage: ((event: MessageEvent) => void) | null;
  onclose: ((event: CloseEvent) => void) | null;
  onerror: ((event: Event) => void) | null;
  protocol: string;
  extensions: string;
  bufferedAmount: number;
  binaryType: BinaryType;

  constructor(url: string) {
    this.url = url
    this.readyState = WebSocketMock.CONNECTING
    this.onopen = null
    this.onmessage = null
    this.onclose = null
    this.onerror = null
    this.protocol = ''
    this.extensions = ''
    this.bufferedAmount = 0
    this.binaryType = 'blob'
    
    // Simulate connection
    setTimeout(() => {
      this.readyState = WebSocketMock.OPEN
      if (this.onopen) {
        this.onopen({ type: 'open' } as Event)
      }
    }, 10)
  }
  
  send(data: string | ArrayBuffer | Blob) {
    // Mock sending data
  }
  
  close(code?: number, reason?: string) {
    this.readyState = WebSocketMock.CLOSED
    if (this.onclose) {
      this.onclose({ type: 'close' } as CloseEvent)
    }
  }
}

// Mock WebSocket global
Object.defineProperty(global, 'WebSocket', {
  value: WebSocketMock,
  writable: true
});

// Mock localStorage with all required properties
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
  length: 0,
  key: jest.fn(),
};
global.localStorage = localStorageMock

// Mock sessionStorage with all required properties
const sessionStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
  length: 0,
  key: jest.fn(),
}
global.sessionStorage = sessionStorageMock

// Mock fetch
global.fetch = jest.fn()

// Mock crypto.randomUUID
Object.defineProperty(global, 'crypto', {
  value: {
    randomUUID: () => 'mock-uuid-' + Math.random().toString(36).substr(2, 9)
  }
})

// Mock MediaSource APIs
global.HTMLMediaElement.prototype.play = jest.fn(() => Promise.resolve())
global.HTMLMediaElement.prototype.pause = jest.fn()
global.HTMLMediaElement.prototype.load = jest.fn()

// Mock canvas context with proper CanvasRenderingContext2D properties
HTMLCanvasElement.prototype.getContext = jest.fn().mockReturnValue({
  drawImage: jest.fn(),
  getImageData: jest.fn(() => ({ data: new Array(4) })),
  putImageData: jest.fn(),
  createImageData: jest.fn(() => ({ data: new Array(4) })),
  setTransform: jest.fn(),
  drawFocusIfNeeded: jest.fn(),
  createLinearGradient: jest.fn(() => ({
    addColorStop: jest.fn()
  })),
  createRadialGradient: jest.fn(() => ({
    addColorStop: jest.fn()
  })),
  // Add required CanvasRenderingContext2D properties
  canvas: document.createElement('canvas'),
  getContextAttributes: jest.fn(() => ({})),
  globalAlpha: 1,
  globalCompositeOperation: 'source-over',
  strokeStyle: '#000000',
  fillStyle: '#000000',
  lineWidth: 1,
  lineCap: 'butt',
  lineJoin: 'miter',
  miterLimit: 10,
  shadowOffsetX: 0,
  shadowOffsetY: 0,
  shadowBlur: 0,
  shadowColor: 'rgba(0, 0, 0, 0)',
  font: '10px sans-serif',
  textAlign: 'start',
  textBaseline: 'alphabetic',
  direction: 'ltr',
  closePath: jest.fn(),
  moveTo: jest.fn(),
  lineTo: jest.fn(),
  beginPath: jest.fn(),
  stroke: jest.fn(),
  fill: jest.fn(),
  rect: jest.fn(),
  clearRect: jest.fn(),
  fillRect: jest.fn(),
  measureText: jest.fn(() => ({ width: 100 })),
  save: jest.fn(),
  restore: jest.fn(),
  translate: jest.fn(),
  rotate: jest.fn(),
  scale: jest.fn(),
  transform: jest.fn(),
  resetTransform: jest.fn(),
  clip: jest.fn(),
  arc: jest.fn(),
  quadraticCurveTo: jest.fn(),
  bezierCurveTo: jest.fn(),
  createPattern: jest.fn(),
  isPointInPath: jest.fn(),
  isPointInStroke: jest.fn(),
  drawImageToFit: jest.fn(),
  getImageSmoothingEnabled: true,
  setImageSmoothingEnabled: jest.fn()
})

// Suppress console warnings for tests
const originalError = console.error
beforeAll(() => {
  console.error = (...args) => {
    if (
      typeof args[0] === 'string' &&
      args[0].includes('Warning: ReactDOM.render is deprecated')
    ) {
      return
    }
    originalError.call(console, ...args)
  }
})

afterAll(() => {
  console.error = originalError
})