import { faker } from '@faker-js/faker'

// Media item factory
export const createMedia = (overrides = {}) => ({
  id: faker.datatype.uuid(),
  name: `${faker.system.fileName()}.${faker.system.fileType()}`,
  path: faker.system.filePath(),
  size: faker.datatype.number({ min: 1000, max: 10000000 }),
  mimeType: faker.system.mimeType(),
  createdAt: faker.date.past(),
  updatedAt: faker.date.recent(),
  checksum: faker.datatype.hexadecimal({ length: 32 }),
  thumbnail: faker.image.url(),
  duration: faker.datatype.number({ min: 0, max: 7200 }), // 0-2 hours in seconds
  resolution: {
    width: faker.datatype.number({ min: 640, max: 3840 }),
    height: faker.datatype.number({ min: 480, max: 2160 })
  },
  metadata: {
    title: faker.lorem.words(3),
    artist: faker.name.fullName(),
    album: faker.lorem.words(2),
    year: faker.datatype.number({ min: 1900, max: 2023 }),
    genre: faker.music.genre(),
    description: faker.lorem.paragraph()
  },
  tags: Array.from({ length: faker.datatype.number({ min: 0, max: 5 }) }, () => faker.lorem.word()),
  collectionId: faker.datatype.uuid(),
  ...overrides
})

// Create multiple media items
export const createMediaList = (count, overrides = {}) => 
  Array.from({ length: count }, () => createMedia(overrides))

// Create media with specific type
export const createMediaByType = (type) => {
  switch (type) {
    case 'image':
      return createMedia({
        mimeType: 'image/jpeg',
        name: `${faker.system.fileName()}.jpg`,
        thumbnail: faker.image.url()
      })
    
    case 'video':
      return createMedia({
        mimeType: 'video/mp4',
        name: `${faker.system.fileName()}.mp4`,
        duration: faker.datatype.number({ min: 60, max: 7200 }),
        resolution: {
          width: faker.datatype.number({ min: 1280, max: 3840 }),
          height: faker.datatype.number({ min: 720, max: 2160 })
        },
        thumbnail: faker.image.url()
      })
    
    case 'audio':
      return createMedia({
        mimeType: 'audio/mpeg',
        name: `${faker.system.fileName()}.mp3`,
        duration: faker.datatype.number({ min: 180, max: 600 }),
        metadata: {
          title: faker.music.songName(),
          artist: faker.name.fullName(),
          album: faker.lorem.words(2),
          year: faker.datatype.number({ min: 1900, max: 2023 }),
          genre: faker.music.genre()
        }
      })
    
    case 'document':
      return createMedia({
        mimeType: 'application/pdf',
        name: `${faker.system.fileName()}.pdf`,
        metadata: {
          title: faker.lorem.words(3),
          description: faker.lorem.paragraph(),
          author: faker.name.fullName()
        }
      })
    
    default:
      return createMedia()
  }
}

// User factory
export const createUser = (overrides = {}) => ({
  id: faker.datatype.uuid(),
  username: faker.internet.userName(),
  email: faker.internet.email(),
  password: faker.internet.password(),
  role: faker.helpers.arrayElement(['user', 'admin', 'moderator']),
  avatar: faker.image.avatar(),
  createdAt: faker.date.past(),
  updatedAt: faker.date.recent(),
  preferences: {
    theme: faker.helpers.arrayElement(['light', 'dark']),
    language: faker.helpers.arrayElement(['en', 'es', 'fr', 'de']),
    notifications: faker.datatype.boolean(),
    autoPlay: faker.datatype.boolean(),
    quality: faker.helpers.arrayElement(['low', 'medium', 'high'])
  },
  ...overrides
})

// Create admin user
export const createAdminUser = () => createUser({ role: 'admin' })

// Collection factory
export const createCollection = (overrides = {}) => ({
  id: faker.datatype.uuid(),
  name: faker.lorem.words(2),
  description: faker.lorem.sentence(),
  itemCount: faker.datatype.number({ min: 0, max: 1000 }),
  coverImage: faker.image.url(),
  createdAt: faker.date.past(),
  updatedAt: faker.date.recent(),
  userId: faker.datatype.uuid(),
  isPublic: faker.datatype.boolean(),
  tags: Array.from({ length: faker.datatype.number({ min: 0, max: 5 }) }, () => faker.lorem.word()),
  ...overrides
})

// Storage source factory
export const createStorageSource = (overrides = {}) => {
  const protocols = ['smb', 'ftp', 'nfs', 'webdav', 'local']
  const protocol = faker.helpers.arrayElement(protocols)
  
  return {
    id: faker.datatype.uuid(),
    name: faker.lorem.words(2),
    protocol,
    host: faker.internet.domainName(),
    path: faker.system.directoryPath(),
    port: faker.datatype.number({ min: 1, max: 65535 }),
    username: faker.internet.userName(),
    password: faker.internet.password(),
    enabled: faker.datatype.boolean(),
    lastConnected: faker.date.recent(),
    totalSize: faker.datatype.number({ min: 1000000000, max: 1000000000000 }),
    usedSize: faker.datatype.number({ min: 0, max: 1000000000 }),
    createdAt: faker.date.past(),
    updatedAt: faker.date.recent(),
    ...overrides
  }
}

// Search result factory
export const createSearchResult = (query, count = 5) => ({
  query,
  results: createMediaList(count),
  total: count,
  page: 1,
  totalPages: Math.ceil(count / 50),
  facets: {
    type: {
      image: faker.datatype.number({ min: 0, max: count }),
      video: faker.datatype.number({ min: 0, max: count }),
      audio: faker.datatype.number({ min: 0, max: count }),
      document: faker.datatype.number({ min: 0, max: count })
    },
    source: {
      local: faker.datatype.number({ min: 0, max: count }),
      smb: faker.datatype.number({ min: 0, max: count }),
      ftp: faker.datatype.number({ min: 0, max: count }),
      nfs: faker.datatype.number({ min: 0, max: count }),
      webdav: faker.datatype.number({ min: 0, max: count })
    },
    year: Array.from({ length: faker.datatype.number({ min: 1, max: 5 }) }, () => ({
      year: faker.datatype.number({ min: 2000, max: 2023 }),
      count: faker.datatype.number({ min: 0, max: count })
    }))
  },
  suggestions: Array.from({ length: faker.datatype.number({ min: 1, max: 5 }) }, () => faker.lorem.words(2))
})

// Analytics data factory
export const createAnalyticsData = () => ({
  overview: {
    totalFiles: faker.datatype.number({ min: 1000, max: 100000 }),
    totalSize: faker.datatype.number({ min: 1000000000, max: 1000000000000 }),
    totalCollections: faker.datatype.number({ min: 10, max: 100 }),
    totalUsers: faker.datatype.number({ min: 1, max: 100 })
  },
  byType: {
    image: faker.datatype.number({ min: 100, max: 10000 }),
    video: faker.datatype.number({ min: 50, max: 5000 }),
    audio: faker.datatype.number({ min: 200, max: 20000 }),
    document: faker.datatype.number({ min: 20, max: 2000 })
  },
  bySource: {
    local: faker.datatype.number({ min: 100, max: 10000 }),
    smb: faker.datatype.number({ min: 50, max: 5000 }),
    ftp: faker.datatype.number({ min: 20, max: 2000 }),
    nfs: faker.datatype.number({ min: 10, max: 1000 }),
    webdav: faker.datatype.number({ min: 5, max: 500 })
  },
  activity: Array.from({ length: 30 }, () => ({
    date: faker.date.recent({ days: 30 }),
    uploads: faker.datatype.number({ min: 0, max: 100 }),
    downloads: faker.datatype.number({ min: 0, max: 200 }),
    searches: faker.datatype.number({ min: 0, max: 500 })
  })),
  popularMedia: createMediaList(10),
  activeUsers: Array.from({ length: 5 }, () => createUser())
})

// Error response factory
export const createErrorResponse = (message, code = 500) => ({
  success: false,
  error: {
    message,
    code,
    timestamp: faker.date.recent().toISOString(),
    requestId: faker.datatype.uuid()
  }
})

// Success response wrapper
export const createSuccessResponse = (data, message = 'Success') => ({
  success: true,
  data,
  message,
  timestamp: faker.date.recent().toISOString()
})