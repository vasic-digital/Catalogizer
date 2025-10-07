# 🎬 Advanced Media Detection & Metadata System

A comprehensive media content detection, analysis, and metadata management system that automatically identifies, categorizes, and enriches your media collection with external metadata from multiple sources.

## 🌟 **Key Features**

### 🔍 **Universal Content Detection**
- **50+ Media Types**: Movies, TV shows, anime, music, games, software, training materials, YouTube content, podcasts, ebooks, and more
- **Multi-Pattern Detection**: Filename patterns, directory structure analysis, file content inspection
- **AI-Powered Classification**: Advanced algorithms with confidence scoring
- **Real-time Analysis**: Automatic detection as files change on SMB shares

### 📊 **Rich Metadata Integration**
- **15+ External Providers**: IMDB, TMDB, TVDB, MusicBrainz, Spotify, Steam, YouTube, GitHub, etc.
- **Comprehensive Details**: Covers, trailers, reviews, ratings, cast/crew, technical specs
- **Auto-Enrichment**: Automatically fetches and updates metadata
- **Multi-Source Aggregation**: Combines data from multiple providers for complete information

### 💎 **Quality Analysis & Version Management**
- **Quality Detection**: 4K/UHD, 1080p, 720p, DVD, BluRay, WEB-DL, HDR support
- **Version Tracking**: Multiple qualities, languages, formats per media item
- **Duplicate Detection**: Finds duplicate content across different qualities
- **Smart Recommendations**: Suggests missing qualities or better versions

### 🔄 **Real-Time Monitoring**
- **Multi-Protocol Change Detection**: Monitors file system changes in real-time across all supported protocols
- **Auto-Reanalysis**: Automatically updates catalog when files change
- **Intelligent Debouncing**: Avoids excessive processing during bulk operations
- **Change Logging**: Complete audit trail of all modifications

### 🔐 **Enterprise Security**
- **SQLCipher Encryption**: Database encrypted at rest
- **Secure API**: JWT authentication, CORS support
- **Audit Logging**: Complete activity tracking
- **Backup & Recovery**: Encrypted backup capabilities

## 🏗️ **System Architecture**

```
┌─────────────────────────────────────────────────────────────────┐
│                     REST API Layer                              │
├─────────────────────────────────────────────────────────────────┤
│  Media Handler  │  Search  │  Analytics  │  Real-time Updates  │
├─────────────────────────────────────────────────────────────────┤
│                    Media Manager                                │
├─────────────────────────────────────────────────────────────────┤
│ Detection Engine │ Analyzer │ Providers │ Change Watcher       │
├─────────────────────────────────────────────────────────────────┤
│              SQLCipher Encrypted Database                      │
├─────────────────────────────────────────────────────────────────┤
│                   SMB File System                              │
└─────────────────────────────────────────────────────────────────┘
```

### **Core Components**

1. **Detection Engine** (`internal/media/detector/`)
   - Pattern-based content type detection
   - Multi-method analysis (filename, structure, content)
   - Confidence scoring and validation

2. **Metadata Providers** (`internal/media/providers/`)
   - TMDB, IMDB, TVDB for movies/TV
   - MusicBrainz, Spotify for music
   - IGDB, Steam for games
   - Custom providers for specialized content

3. **Content Analyzer** (`internal/media/analyzer/`)
   - Real-time directory analysis
   - Quality detection and comparison
   - Version aggregation and tracking

4. **Change Watcher** (`internal/media/realtime/`)
   - File system monitoring
   - Real-time update processing
   - Intelligent change debouncing

5. **Encrypted Database** (`internal/media/database/`)
   - SQLCipher-encrypted SQLite
   - Comprehensive schema for media metadata
   - Performance-optimized indexes

## 🚀 **Supported Media Types**

### 🎬 **Video Content**
- **Movies**: Feature films, documentaries, short films
- **TV Shows**: Series, episodes, seasons, miniseries
- **Anime**: Japanese animation series and movies
- **YouTube**: Videos, channels, vlogs, tutorials
- **Sports**: Matches, tournaments, championships
- **Comedy**: Stand-up specials, comedy shows

### 🎵 **Audio Content**
- **Music**: Albums, singles, discographies, lossless audio
- **Podcasts**: Episodes, series, talk shows
- **Audiobooks**: Narrated books, unabridged content
- **Radio Shows**: Broadcasts, AM/FM content
- **Sound Effects**: SFX libraries, audio samples

### 🎮 **Gaming Content**
- **PC Games**: Steam, Epic, GOG releases
- **Console Games**: PlayStation, Xbox, Nintendo
- **Mobile Games**: Android, iOS applications
- **Game Mods**: Modifications, patches, expansions
- **Emulators**: Retro gaming, ROM collections

### 💾 **Software & Applications**
- **Applications**: Windows, macOS, Linux software
- **Operating Systems**: ISO images, distributions
- **Drivers**: Hardware drivers, firmware
- **Plugins**: Extensions, add-ons, libraries
- **Portable Apps**: Standalone applications

### 📚 **Educational & Reference**
- **Training Courses**: Professional development, certifications
- **Academic Content**: Lectures, university courses
- **Tutorials**: How-to guides, DIY content
- **Language Learning**: Rosetta Stone, Duolingo content
- **eBooks**: Digital books, novels, manuals
- **Research Papers**: Academic publications, journals

### 🎨 **Creative & Design**
- **Templates**: Design assets, PSD files
- **3D Models**: Assets for games and animation
- **Fonts**: Typography collections
- **Wallpapers**: Desktop backgrounds, art
- **Comics**: Digital comics, manga, graphic novels

## 📊 **Database Schema Highlights**

### **Core Tables**
- **`media_types`**: 50+ predefined content categories
- **`media_items`**: Detected media with aggregated metadata
- **`external_metadata`**: Data from IMDB, TMDB, etc.
- **`media_files`**: Individual file versions and qualities
- **`directory_analysis`**: Detection results and confidence scores
- **`change_log`**: Real-time change tracking

### **Advanced Features**
- **Collections**: Group related content (TV series, album discographies)
- **Quality Profiles**: Define and compare quality standards
- **User Metadata**: Personal ratings, watch status, notes
- **Detection Rules**: Customizable pattern matching

## 🛠️ **API Endpoints**

### **Media Discovery**
```bash
# Search all media types
GET /api/v1/media/search?query=avengers&year=2012&quality=1080p

# Get media details with all metadata
GET /api/v1/media/12345

# Browse by media type
GET /api/v1/media/search?media_types=movie,tv_show&genre=action
```

### **Content Analysis**
```bash
# Trigger directory analysis
POST /api/v1/media/analyze
{
  "directory_path": "/nas/movies/Marvel",
  "smb_root": "nas1",
  "priority": 8
}

# Get analysis results
GET /api/v1/media/analysis/status?directory=/nas/movies/Marvel
```

### **Quality Management**
```bash
# Compare file qualities
GET /api/v1/media/12345/versions

# Find missing qualities
GET /api/v1/media/12345/missing-qualities

# Get duplicate analysis
GET /api/v1/media/duplicates?min_size=1GB
```

### **Statistics & Analytics**
```bash
# Comprehensive statistics
GET /api/v1/media/stats

# Media type distribution
GET /api/v1/media/stats/distribution

# Quality analysis
GET /api/v1/media/stats/quality

# Recent activity
GET /api/v1/media/stats/activity?since=24h
```

## ⚡ **Performance Features**

### **Real-Time Processing**
- **Change Detection**: Sub-second file system monitoring
- **Queue Management**: Priority-based analysis queuing
- **Parallel Processing**: Multi-worker content analysis
- **Debouncing**: Intelligent batching of rapid changes

### **Scalability**
- **Concurrent Analysis**: Multiple directories processed simultaneously
- **Incremental Updates**: Only processes changed content
- **Memory Efficient**: Streaming processing for large collections
- **Database Optimization**: Indexed queries and views

### **Caching & Performance**
- **Metadata Caching**: Reduces API calls to external providers
- **Result Caching**: Stores detection results for quick retrieval
- **Smart Reanalysis**: Only re-processes when necessary
- **Batch Operations**: Efficient bulk processing

## 🔧 **Configuration Example**

```json
{
  "media": {
    "database_path": "media_catalog.db",
    "database_password": "secure_encryption_key",
    "enable_realtime": true,
    "analysis_workers": 4,
    "api_keys": {
      "tmdb": "your_tmdb_api_key",
      "imdb": "your_imdb_api_key",
      "spotify": "your_spotify_client_id",
      "youtube": "your_youtube_api_key"
    },
    "watch_paths": [
      {
        "smb_root": "nas1",
        "local_path": "/mnt/smb/nas1",
        "enabled": true
      },
      {
        "smb_root": "nas2",
        "local_path": "/mnt/smb/nas2",
        "enabled": true
      }
    ],
    "quality_profiles": {
      "preferred_video": ["4K/UHD", "1080p", "720p"],
      "preferred_audio": ["Lossless", "320k", "256k"],
      "minimum_rating": 6.0
    }
  }
}
```

## 🚀 **Quick Start**

### 1. **Initialize the System**
```go
package main

import (
    "catalog-api/internal/config"
    "catalog-api/internal/media"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    cfg, _ := config.Load()

    // Create media manager
    mediaManager, err := media.NewMediaManager(cfg, logger)
    if err != nil {
        log.Fatal(err)
    }

    // Start all services
    if err := mediaManager.Start(); err != nil {
        log.Fatal(err)
    }
    defer mediaManager.Stop()

    // Trigger full analysis
    ctx := context.Background()
    mediaManager.AnalyzeAllDirectories(ctx)
}
```

### 2. **Integrate with Existing API**
```go
// Add to your main.go
mediaManager, _ := media.NewMediaManager(cfg, logger)
mediaManager.Start()

mediaHandler := handlers.NewMediaHandler(
    mediaManager.GetDatabase(),
    mediaManager.GetAnalyzer(),
    logger,
)

// Add routes
api.GET("/media/search", mediaHandler.SearchMedia)
api.GET("/media/:id", mediaHandler.GetMediaItem)
api.POST("/media/analyze", mediaHandler.AnalyzeDirectory)
api.GET("/media/stats", mediaHandler.GetMediaStats)
```

### 3. **Real-Time Monitoring Setup**
```go
// Configure SMB monitoring
changeWatcher := mediaManager.GetChangeWatcher()

// Add watch paths
changeWatcher.WatchSMBPath("nas1", "/mnt/smb/nas1")
changeWatcher.WatchSMBPath("nas2", "/mnt/smb/nas2")

// Monitor changes
stats, _ := changeWatcher.GetChangeStatistics(time.Now().Add(-24*time.Hour))
fmt.Printf("Changes in last 24h: %+v\n", stats)
```

## 📈 **Example Results**

### **Movie Detection Result**
```json
{
  "id": 12345,
  "title": "Avengers: Endgame",
  "year": 2019,
  "media_type": "movie",
  "external_metadata": [
    {
      "provider": "tmdb",
      "rating": 8.4,
      "cover_url": "https://image.tmdb.org/t/p/w500/...",
      "trailer_url": "https://youtube.com/watch?v=..."
    },
    {
      "provider": "imdb",
      "rating": 8.4,
      "review_url": "https://imdb.com/title/tt4154796/"
    }
  ],
  "files": [
    {
      "filename": "Avengers.Endgame.2019.2160p.BluRay.x265-TERMINAL.mkv",
      "quality_info": {
        "resolution": {"width": 3840, "height": 2160},
        "quality_profile": "4K/UHD",
        "source": "BluRay",
        "video_codec": "H.265/HEVC",
        "hdr": true,
        "quality_score": 100
      },
      "file_size": 15728640000,
      "direct_smb_link": "smb://nas1/movies/Avengers.Endgame.2019.2160p.BluRay.x265-TERMINAL.mkv",
      "virtual_smb_link": "virtual://nas1/12345"
    }
  ],
  "available_qualities": ["4K/UHD", "1080p", "720p"],
  "duplicate_count": 3,
  "total_size": 25769803776
}
```

### **Quality Analysis**
```json
{
  "media_item_id": 12345,
  "title": "Game of Thrones",
  "available_versions": [
    {"quality": "4K/UHD", "seasons": [1,2,3,4,5,6,7,8], "total_size": "245GB"},
    {"quality": "1080p", "seasons": [1,2,3,4,5,6,7,8], "total_size": "89GB"},
    {"quality": "720p", "seasons": [1,2,3,4,5,6], "total_size": "34GB"}
  ],
  "missing_qualities": ["DVD"],
  "recommended_upgrade": "Complete 4K collection",
  "duplicate_episodes": 12,
  "wasted_space": "15.2GB"
}
```

## 🎯 **Use Cases**

1. **Media Server Management**: Automatically organize and enrich Plex/Jellyfin libraries
2. **Content Archival**: Track and catalog large media collections
3. **Quality Control**: Monitor and upgrade media quality across collections
4. **Duplicate Management**: Find and remove redundant files
5. **Metadata Enrichment**: Enhance existing catalogs with rich external data
6. **Real-time Monitoring**: Track changes to shared media storage
7. **Collection Analytics**: Analyze viewing patterns and content distribution

## 🔮 **Future Enhancements**

- **ML-Based Detection**: Advanced machine learning for content classification
- **Video Analysis**: Frame-by-frame content analysis for scene detection
- **Audio Fingerprinting**: Acoustic identification for music and audio
- **Subtitle Management**: Automatic subtitle detection and matching
- **Smart Recommendations**: AI-powered content suggestions
- **Mobile Apps**: Native iOS/Android applications
- **Cloud Integration**: Support for cloud storage providers
- **Advanced Analytics**: Predictive analytics and trend analysis

This comprehensive media detection system transforms your file catalog into an intelligent, enriched media database with automatic real-time updates and rich metadata from dozens of external sources.