import { createMedia, createMediaList } from './factories'

// Mock API responses
export const mockApiResponses = {
  // Media endpoints
  getMedia: {
    success: {
      data: createMediaList(10),
      total: 10,
      page: 1,
      limit: 50
    },
    error: {
      message: 'Failed to fetch media',
      code: 500
    }
  },
  
  searchMedia: {
    success: {
      data: createMediaList(5),
      query: 'test',
      total: 5,
      page: 1
    },
    error: {
      message: 'Search failed',
      code: 400
    }
  },
  
  getMediaStats: {
    success: {
      totalFiles: 1000,
      totalSize: 5000000000,
      byType: {
        image: 600,
        video: 200,
        audio: 150,
        document: 50
      },
      bySource: {
        local: 400,
        smb: 300,
        ftp: 200,
        nfs: 100
      }
    },
    error: {
      message: 'Stats unavailable',
      code: 503
    }
  },
  
  downloadMedia: {
    success: {
      url: 'https://example.com/download/media123',
      expires: '2023-12-31T23:59:59Z'
    },
    error: {
      message: 'Download failed',
      code: 404
    }
  },
  
  // Authentication endpoints
  login: {
    success: {
      token: 'mock-jwt-token',
      user: {
        id: 'user123',
        username: 'testuser',
        email: 'test@example.com',
        role: 'user'
      },
      expiresIn: 3600
    },
    error: {
      message: 'Invalid credentials',
      code: 401
    }
  },
  
  register: {
    success: {
      user: {
        id: 'newuser123',
        username: 'newuser',
        email: 'newuser@example.com',
        role: 'user'
      },
      message: 'Registration successful'
    },
    error: {
      message: 'User already exists',
      code: 409
    }
  },
  
  // Collections endpoints
  getCollections: {
    success: {
      data: [
        {
          id: 'collection1',
          name: 'My Photos',
          description: 'Photo collection',
          itemCount: 150,
          createdAt: '2023-01-01T00:00:00Z'
        },
        {
          id: 'collection2',
          name: 'Videos',
          description: 'Video collection',
          itemCount: 50,
          createdAt: '2023-02-01T00:00:00Z'
        }
      ],
      total: 2
    },
    error: {
      message: 'Failed to fetch collections',
      code: 500
    }
  }
}

// WebSocket mock events
export const mockWebSocketEvents = {
  mediaAdded: {
    type: 'media_added',
    data: createMedia()
  },
  
  mediaDeleted: {
    type: 'media_deleted',
    data: { id: 'media123' }
  },
  
  scanStarted: {
    type: 'scan_started',
    data: { sourceId: 'source123' }
  },
  
  scanCompleted: {
    type: 'scan_completed',
    data: {
      sourceId: 'source123',
      foundCount: 25,
      duration: 120
    }
  },
  
  systemStatus: {
    type: 'system_status',
    data: {
      activeConnections: 5,
      cpuUsage: 25.5,
      memoryUsage: 45.2,
      diskUsage: 60.8
    }
  }
}

// Mock API client implementation
export class MockApiClient {
  constructor() {
    this.responses = mockApiResponses
    this.shouldFail = false
    this.delay = 100
  }
  
  setFailureMode(shouldFail) {
    this.shouldFail = shouldFail
  }
  
  setDelay(delay) {
    this.delay = delay
  }
  
  async get(endpoint, params = {}) {
    return this.simulateApiCall(endpoint, 'GET', params)
  }
  
  async post(endpoint, data = {}) {
    return this.simulateApiCall(endpoint, 'POST', data)
  }
  
  async put(endpoint, data = {}) {
    return this.simulateApiCall(endpoint, 'PUT', data)
  }
  
  async delete(endpoint) {
    return this.simulateApiCall(endpoint, 'DELETE')
  }
  
  simulateApiCall(endpoint, method, data) {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        if (this.shouldFail) {
          reject(new Error('Simulated API failure'))
          return
        }
        
        // Default response based on endpoint
        let response = { data: null, success: true }
        
        switch (endpoint) {
          case '/media':
            if (method === 'GET') {
              response.data = this.responses.getMedia.success
            } else if (method === 'POST') {
              response.data = { id: 'new-media-id', ...data }
            }
            break
            
          case '/media/search':
            response.data = this.responses.searchMedia.success
            break
            
          case '/media/stats':
            response.data = this.responses.getMediaStats.success
            break
            
          case '/auth/login':
            response.data = this.responses.login.success
            break
            
          case '/auth/register':
            response.data = this.responses.register.success
            break
            
          case '/collections':
            response.data = this.responses.getCollections.success
            break
            
          default:
            response.data = { message: 'Mock response for ' + endpoint }
        }
        
        resolve(response)
      }, this.delay)
    })
  }
}

// Mock WebSocket implementation
export class MockWebSocket {
  constructor(url) {
    this.url = url
    this.readyState = WebSocket.CONNECTING
    this.onopen = null
    this.onmessage = null
    this.onclose = null
    this.onerror = null
    this.eventQueue = []
    
    // Simulate connection
    setTimeout(() => {
      this.readyState = WebSocket.OPEN
      if (this.onopen) {
        this.onopen({ type: 'open' })
      }
    }, 50)
  }
  
  send(data) {
    // Mock sending data - in tests this would be verified
    console.log('WebSocket send:', data)
  }
  
  close() {
    this.readyState = WebSocket.CLOSED
    if (this.onclose) {
      this.onclose({ type: 'close' })
    }
  }
  
  // Simulate receiving an event
  simulateEvent(eventType, eventData) {
    if (this.onmessage && this.readyState === WebSocket.OPEN) {
      this.onmessage({
        type: 'message',
        data: JSON.stringify({
          type: eventType,
          data: eventData
        })
      })
    }
  }
  
  // Simulate connection error
  simulateError() {
    this.readyState = WebSocket.CLOSED
    if (this.onerror) {
      this.onerror({ type: 'error' })
    }
  }
}