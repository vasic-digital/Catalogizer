# Catalogizer Video Course Scripts

## Course Catalog

This document contains video course scripts for Catalogizer users, developers, and integrators.

**Total Duration:** ~4 hours across 3 courses
**Courses:** 3 (User Onboarding, Developer Training, API Integration)
**Target Audience:** End users, Developers, Integrators

---

# COURSE 1: User Onboarding (5 videos - 32 minutes)

## Video 1.1: Getting Started with Catalogizer (5 min)

### Learning Objectives
- Install Catalogizer
- Complete initial setup wizard
- Navigate the main dashboard

### Script

**[00:00-00:30] - Introduction**
```
Welcome to Catalogizer! I'm excited to show you how to organize
your entire media collection in one place.

Whether you have movies, TV shows, music, games, books, or comics -
Catalogizer handles it all automatically.
```

**[00:30-02:00] - Installation**
```
Let's start with installation. You have several options:

Option 1: Download from our website (catalogizer.io)
- Available for Windows, macOS, and Linux
- Just download and run the installer

Option 2: Use Docker
```bash
docker run -p 3000:3000 -p 8080:8080 vasicdigital/catalogizer
```

Option 3: Build from source
- Clone the repository
- Follow the README instructions

For this tutorial, I'll use the desktop installer.
```

**[02:00-04:00] - Initial Setup**
```
When you first launch Catalogizer, you'll see the setup wizard.

Step 1: Create your admin account
- Choose a username and secure password
- This account has full access to everything

Step 2: Configure your first storage location
- This is where your media files are stored
- Can be local folders, network drives, or cloud storage
- You can add more locations later

Step 3: Set your preferences
- Default language
- Theme (light/dark)
- Media types to enable

**[04:00-05:00] - Dashboard Overview**
```
Setup complete! Welcome to your dashboard.

Here's what you see:
- Top: Search bar and quick actions
- Left sidebar: Navigation between media types
- Main area: Overview cards and recent items
- Bottom: Storage usage and system status

Key features:
- Total items count updates automatically
- Recently added shows your newest media
- Storage usage helps you manage space
- Quick scan button for instant updates
```

---

## Video 1.2: Understanding Media Types (7 min)

### Script

**[00:00-01:00] - Introduction to Media Types**
```
Catalogizer supports 11 media types, each with specialized
handling and metadata sources.

Let's explore them one by one.
```

**[01:00-02:00] - Movies & TV Shows**
```
MOVIES:
- Automatic detection from filenames
- Metadata from TMDB, IMDB, Rotten Tomatoes
- Posters, backdrops, cast information
- Trailer links and ratings

TV SHOWS:
- Full season/episode organization
- Episode-specific metadata
- Watch status tracking per episode
- "Up next" recommendations

Both support:
- 4K, HDR, Dolby Vision detection
- Multiple audio/subtitle tracks
- File quality indicators
```

**[02:00-03:00] - Music**
```
MUSIC:
- Artist and album organization
- Track-level metadata
- Genre classification
- Playlist creation and management
- Integration with MusicBrainz and Last.fm

Special features:
- Album art fetching
- Lyrics support
- Audio quality indicators (FLAC, MP3 bitrate)
- Compilation handling
```

**[03:00-04:00] - Games & Software**
```
GAMES:
- PC games (Steam, GOG, Epic, etc.)
- Console games (with emulator support)
- Mobile games (APK/IPA tracking)
- Metadata from IGDB and Steam

SOFTWARE:
- Application tracking
- Version management
- License key storage (encrypted)
- Installation package organization
```

**[04:00-05:00] - Books & Comics**
```
BOOKS:
- E-books (EPUB, MOBI, PDF)
- Audiobooks with chapter support
- Metadata from OpenLibrary, Goodreads
- Reading progress tracking

COMICS:
- CBZ/CBR archive support
- Page-by-page navigation
- Series and volume organization
- Metadata from ComicVine
```

**[05:00-06:00] - Other Media Types**
```
OTHER SUPPORTED TYPES:

PODCASTS:
- RSS feed integration
- Episode downloads
- Listening progress

PHOTOS:
- EXIF data extraction
- Location mapping
- Album organization

DOCUMENTS:
- PDF, Word, Excel support
- Full-text search
- Tag-based organization

ANIME:
- Special anime metadata fields
- MyAnimeList integration
- Episode tracking

YOUTUBE VIDEOS:
- Channel subscriptions
- Playlist import
- Offline downloads
```

**[06:00-07:00] - Summary**
```
With 11 media types, Catalogizer is your complete digital
library solution.

Each type has:
- Automatic metadata fetching
- Specialized organization
- Quality indicators
- Unique features tailored to the content

Next, let's see how to organize everything effectively.
```

---

## Video 1.3: Organizing Your Collection (6 min)

### Script

**[00:00-01:00] - Collections**
```
COLLECTIONS are the primary way to organize your media.

Think of them like playlists or folders, but more powerful.

Creating a collection:
1. Click "New Collection" button
2. Give it a name and description
3. Choose a cover image
4. Start adding items

Examples:
- "Favorite Movies"
- "Christmas Music"
- "Workout Playlist"
- "Retro Games"
```

**[01:00-02:30] - Smart Collections**
```
SMART COLLECTIONS automatically update based on rules.

Examples:
- "Unwatched Movies" - movies where play count = 0
- "4K Content" - files with 4K resolution
- "Recent Additions" - added in last 30 days
- "High Rated" - rating > 8.0

Setting up rules:
1. Select metadata field (year, rating, genre, etc.)
2. Choose operator (equals, greater than, contains)
3. Enter value
4. Combine multiple rules with AND/OR
```

**[02:30-04:00] - Tags and Labels**
```
TAGS provide flexible, multi-dimensional organization.

Unlike collections, items can have unlimited tags.

Creating tags:
- Use existing tags or create new ones
- Color-code for visual organization
- Nest tags hierarchically

Examples:
- Genre tags: #action, #comedy, #sci-fi
- Quality tags: #4K, #HDR, #remastered
- Personal tags: #favorite, #watch-again, #kids

Filtering:
- Click any tag to filter
- Combine multiple tags
- Exclude tags with NOT operator
```

**[04:00-05:30] - Favorites and Watchlists**
```
FAVORITES:
- Quick access from dashboard
- Heart icon on any item
- Separate favorites per media type

WATCHLIST:
- Items you want to watch/play/read
- Integration with external services
- Reminders for new releases

CUSTOM LISTS:
- Priority lists
- "Up next" queues
- Sharing lists with friends
```

**[05:30-06:00] - Search and Filters**
```
POWERFUL SEARCH:

Basic search:
- Type in search box
- Searches titles, descriptions, cast

Advanced filters:
- Year range
- Rating range
- File quality
- Added date
- Play count

Saved searches:
- Save filter combinations
- Quick access from sidebar
- Share with other users
```

---

## Video 1.4: Scanning and Auto-Detection (8 min)

### Script

**[00:00-01:30] - Starting a Scan**
```
The magic of Catalogizer is automatic detection.

To start scanning:
1. Click the "Scan" button
2. Select storage location
3. Choose media types to detect
4. Set options (deep scan, skip existing, etc.)

Real-time progress shows:
- Files analyzed
- Media found
- Detection confidence
- Current file being processed
```

**[01:30-03:00] - How Detection Works**
```
DETECTION PROCESS:

Step 1: Filename Analysis
- Pattern matching for titles
- Year extraction
- Quality indicators (1080p, BluRay, etc.)
- Episode/season detection

Step 2: File Analysis
- Media info extraction (codec, resolution)
- Duration calculation
- Checksum for duplicate detection
- Thumbnail generation

Step 3: Metadata Fetching
- Query online databases
- Match by title/year
- Download posters and artwork
- Get cast, crew, descriptions

Confidence score shows how sure Catalogizer is about each detection.
```

**[03:00-04:30] - Reviewing Results**
```
AFTER SCAN COMPLETES:

High confidence items (>80%):
- Automatically added to library
- Fully populated metadata
- Ready to use

Medium confidence (50-80%):
- Added to "Review" queue
- You confirm or correct
- One-click fixes

Low confidence (<50%):
- Flagged for manual entry
- All fields editable
- Search for correct metadata

The review interface shows:
- Suggested matches
- Confidence scores
- Alternative options
- Manual search capability
```

**[04:30-06:00] - Handling Duplicates**
```
DUPLICATE DETECTION:

Catalogizer automatically finds duplicates by:
- File checksum
- Title/year matching
- Similar filenames

Duplicate handling options:
- Keep best quality version
- Keep all versions
- Merge metadata
- Delete duplicates

Quality comparison shows:
- Resolution (4K vs 1080p)
- File size
- Codec (H.265 vs H.264)
- Audio tracks
```

**[06:00-07:30] - Scheduled and Automatic Scanning**
```
AUTOMATION OPTIONS:

Scheduled scans:
- Daily, weekly, or custom schedule
- Time-based (scan at 3 AM)
- Incremental (only new files)

Real-time monitoring:
- Watch folders for changes
- Auto-detect new files
- Immediate processing

Notifications:
- Email when scan completes
- Summary of new items found
- Alerts for issues

Settings:
- Exclude patterns (temp files, samples)
- Minimum file size
- File type whitelist
```

**[07:30-08:00] - Best Practices**
```
SCANNING BEST PRACTICES:

1. Organize files before scanning
   - Use consistent naming
   - Group related files
   - Remove samples and extras

2. Start with a subset
   - Test with one folder first
   - Verify detection accuracy
   - Adjust settings as needed

3. Regular maintenance
   - Weekly scans for new content
   - Monthly library cleanup
   - Review and merge duplicates
```

---

## Video 1.5: Playback and Streaming (6 min)

### Script

**[00:00-01:30] - Built-in Media Player**
```
BUILT-IN PLAYER features:

Video playback:
- Hardware acceleration
- Subtitle support (SRT, ASS, VTT)
- Audio track selection
- Playback speed (0.5x to 2x)
- Resume from last position

Music playback:
- Gapless playback
- Visualizations
- Playlist queue
- Shuffle and repeat

Reading:
- PDF viewer with search
- E-book reader with bookmarks
- Comic reader with zoom
```

**[01:30-03:00] - Streaming to Devices**
```
STREAMING OPTIONS:

Smart TVs:
- DLNA/UPnP support
- Samsung, LG, Sony, Roku
- Browse and play from TV interface

Mobile apps:
- iOS and Android apps
- Offline downloads
- Continue watching sync

Web browser:
- Access from any device
- No app installation needed
- Same interface as desktop

Casting:
- Chromecast support
- AirPlay for Apple devices
- Smart display integration
```

**[03:00-04:30] - Transcoding**
```
TRANSCODING features:

Automatic format conversion:
- Unsupported codecs → compatible formats
- High bitrate → bandwidth-optimized
- 4K → 1080p for mobile devices

Quality options:
- Original (direct stream, no conversion)
- High (1080p, high bitrate)
- Medium (720p, medium bitrate)
- Low (480p, low bitrate)

Hardware acceleration:
- Intel QuickSync
- NVIDIA NVENC
- AMD VCE
- Apple VideoToolbox

Settings:
- Per-device defaults
- Adaptive streaming (adjusts to bandwidth)
- Pre-transcode popular content
```

**[04:30-05:30] - Offline Access**
```
DOWNLOADS for offline:

Use cases:
- Travel without internet
- Commute on subway
- Reduce mobile data usage

Download options:
- Quality selection
- Single items or entire collections
- Auto-delete after watching
- Storage management

Sync features:
- Resume position syncs across devices
- Downloads on one device, continue on another
- Auto-download new episodes
```

**[05:30-06:00] - User Management**
```
MULTI-USER features:

User accounts:
- Separate libraries per user
- Individual watch histories
- Personalized recommendations
- Custom permissions

Sharing:
- Share specific collections
- Family groups
- Guest access with limits

Parental controls:
- Content ratings
- Time limits
- Allowed media types
- PIN protection

Congratulations! You now know the basics of Catalogizer.

Check out our advanced tutorials for:
- API integration
- Custom metadata providers
- Plugin development
- Server administration

Happy organizing!
```

---

# COURSE 2: Developer Training (8 videos - ~2 hours)

[Detailed developer training content would continue here...]

---

# COURSE 3: API Integration Guide (3 videos - ~45 min)

[Detailed API integration content would continue here...]

---

## Production Notes

### Recording Specifications
- **Resolution**: 1920x1080 (1080p minimum)
- **Frame Rate**: 30fps
- **Audio**: 48kHz, stereo, -16 LUFS
- **Format**: MP4 (H.264 codec)

### Visual Style
- Clean, modern UI recordings
- Highlight important elements with zoom/pan
- Consistent color scheme
- Professional but approachable tone

### Post-Production
- Captions for accessibility
- Chapter markers for navigation
- Code samples in description
- Links to related resources

### Distribution
- YouTube (primary)
- GitHub repository
- Documentation website
- Course platforms (Udemy, etc.)

---

**Last Updated**: 2026-03-22
**Version**: 1.0

---

## Module 1: Introduction (15 minutes)

### Learning Objectives
- Understand what autonomous QA testing is
- Learn the benefits over traditional QA
- Get an overview of HelixQA architecture

### Script

**[00:00-01:00] - Introduction & Hook**
```
"Welcome to the HelixQA Autonomous QA Session course. 

If you're tired of manually clicking through your application 
to find bugs, writing endless test scripts, or maintaining 
fragile test suites that break with every UI change - this 
course is for you.

Today I'm going to show you how to leverage Large Language 
Models to autonomously test your applications across multiple 
platforms - Android, Web, and Desktop - with minimal setup."
```

**[01:00-04:00] - What is Autonomous QA?**
```
"Traditional QA testing falls into three categories:

1. Manual Testing - Time-consuming, inconsistent, doesn't scale
2. Scripted Testing - Brittle, high maintenance, can't adapt to UI changes  
3. Record & Playback - Limited coverage, breaks easily

Autonomous QA is different. It uses AI to:
- Understand your application through documentation
- Intelligently navigate without brittle selectors
- Detect issues using visual and functional analysis
- Generate comprehensive bug reports automatically

Think of it as having a smart QA engineer that never gets tired, 
works 24/7, and can test on multiple platforms simultaneously."
```

**[04:00-08:00] - HelixQA Architecture Overview**
```
"HelixQA consists of four main phases:

Phase 1: Setup
- LLMsVerifier selects the best available LLM models
- DocProcessor extracts features from your documentation
- LLMOrchestrator initializes agent pools

Phase 2: Document-Driven Verification
- System navigates to documented features
- Verifies each feature works correctly
- Collects evidence along the way

Phase 3: Curiosity-Driven Exploration
- Explores unvisited parts of the app
- Tests edge cases
- Detects visual and functional bugs

Phase 4: Report Generation
- Creates detailed tickets for each issue
- Generates coverage reports
- Provides video evidence with timestamps

The system supports multiple LLM providers including 
Anthropic, OpenAI, Google, and more - automatically 
selecting the best model for each task."
```

**[08:00-12:00] - Live Demo Preview**
```
"Let me show you what a typical session looks like:

[Show screen recording]

In this example, HelixQA is testing a web application.
Notice how it:

1. Reads the README and docs to understand features
2. Navigates to the login page and tests authentication
3. Discovers the settings page wasn't documented
4. Finds a visual bug: truncated button text
5. Generates a detailed ticket with screenshot and fix suggestion

All of this happened automatically in about 5 minutes."
```

**[12:00-15:00] - Course Roadmap**
```
"In this course, we'll cover:

Module 1: Introduction (this video) - Overview and concepts
Module 2: Configuration - Setting up environment and agents
Module 3: Running Sessions - Executing tests and monitoring
Module 4: Advanced Features - Custom strategies and integrations
Module 5: Troubleshooting - Common issues and solutions

By the end, you'll be able to set up autonomous QA testing 
for your own projects and integrate it into your CI/CD pipeline.

Let's get started with configuration in Module 2."
```

---

## Module 2: Configuration (25 minutes)

### Learning Objectives
- Set up environment variables
- Configure LLM providers
- Install and configure CLI agents
- Understand strategy selection

### Script

**[00:00-03:00] - Prerequisites**
```
"Before we configure HelixQA, ensure you have:

1. Go 1.24 or later installed
   Check: go version

2. Git access to the Catalogizer repository

3. At least one LLM API key
   - Anthropic (recommended for vision)
   - OpenAI (GPT-4V)
   - Google (Gemini)

4. Optional but recommended:
   - ffmpeg for video recording
   - OpenCV for enhanced vision

For this tutorial, I'll use Anthropic's Claude as the 
primary LLM since it has excellent vision capabilities."
```

**[03:00-10:00] - Environment Configuration**
```
"Let's create the .env file:

[Show terminal]

cp .env.example .env
nano .env

The most important variables are:

# Master switches
HELIX_AUTONOMOUS_ENABLED=true
HELIX_AUTONOMOUS_PLATFORMS=desktop,web
HELIX_AUTONOMOUS_TIMEOUT=2h

# LLM Provider - minimum one required
ANTHROPIC_API_KEY=sk-ant-...

# Optional: Add more providers for redundancy
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=...

# CLI Agents
HELIX_AGENTS_ENABLED=opencode,claude-code
HELIX_AGENT_POOL_SIZE=3
HELIX_AGENT_TIMEOUT=60s

For testing desktop applications on Linux:
HELIX_DESKTOP_DISPLAY=:0
HELIX_DESKTOP_PROCESS=myapp

For web testing:
HELIX_WEB_URL=http://localhost:3000
HELIX_WEB_BROWSER=chromium

Save the file and we're ready for the next step."
```

**[10:00-16:00] - Installing CLI Agents**
```
"HelixQA supports multiple CLI agents. Let's install them:

1. OpenCode (recommended - multi-provider)
   go install github.com/opencode-ai/opencode@latest

2. Claude Code (Anthropic's official CLI)
   npm install -g @anthropic-ai/claude-code

3. Gemini CLI (Google)
   npm install -g @google/gemini-cli

Verify installations:
which opencode
which claude
which gemini

Each agent has different strengths:
- OpenCode: Flexible, supports multiple LLM backends
- Claude Code: Native Anthropic support, excellent vision
- Gemini: Largest context window, good for large codebases

You can mix and match agents. HelixQA will select the 
best one for each task based on your strategy configuration."
```

**[16:00-22:00] - Strategy Configuration**
```
"The strategy determines how LLMs are selected and scored.

In your .env, set:
LLMSVERIFIER_STRATEGY=helix-qa

Available strategies:

1. helix-qa (default) - Balanced for autonomous testing
   Vision: 25%, Speed: 25%, Quality: 30%, Cost: 10%, Reliability: 10%

2. speed - Fast responses, lower quality
   Useful for quick smoke tests

3. quality - Highest quality, slower
   Use for thorough regression testing

4. vision - Prioritizes vision capabilities
   Essential for UI-heavy applications

5. cost - Budget-conscious
   Minimizes API costs

For most QA testing, helix-qa is optimal. It balances:
- Vision capability for UI analysis
- Speed for responsive interactions  
- Quality for accurate bug detection
- Cost efficiency

You can customize weights if needed by implementing 
a custom strategy - we'll cover that in Module 4."
```

**[22:00-25:00] - Configuration Verification**
```
"Let's verify our configuration:

[Show terminal]

cd HelixQA
go build ./cmd/helixqa
./helixqa --help

Test configuration loading:
./helixqa validate --env ../.env

If everything is correct, you should see:
'Configuration valid: 3 agents configured, 2 platforms enabled'

Common issues:
- Missing API keys: Check .env file permissions (chmod 600)
- Agent not found: Add to PATH or specify full path
- Invalid strategy: Use one of the predefined strategy names

Once validation passes, you're ready to run your first 
autonomous session in Module 3."
```

---

## Module 3: Running Sessions (30 minutes)

### Learning Objectives
- Execute autonomous QA sessions
- Monitor progress in real-time
- Interpret session output
- Manage session lifecycle

### Script

**[00:00-05:00] - First Session Execution**
```
"Let's run your first autonomous QA session.

Prerequisites:
- Your application should be running
- Documentation exists (README, docs folder)
- Configuration from Module 2 is complete

Command structure:
./helixqa autonomous \\
  --project /path/to/project \\
  --platforms desktop \\
  --env ../.env \\
  --output ./qa-results \\
  --timeout 30m

[Show terminal executing command]

Watch the output:
- Phase indicators: [setup] [doc-driven] [curiosity] [report]
- Progress tracking: Feature X/Y
- Issue detection: Real-time alerts
- Timeline events: Actions being taken

The session runs through 4 phases automatically.
Let me explain each one..."
```

**[05:00-12:00] - Understanding Session Phases**
```
"Phase 1: Setup (30-60 seconds)

You'll see:
[setup] LLMsVerifier: Scoring models...
[setup] Selected: claude-3.5-sonnet (score: 0.87)
[setup] DocProcessor: Extracting features...
[setup] Found: 42 features from 12 documents
[setup] LLMOrchestrator: Spawning agents...
[setup] 3 agents ready
[setup] Recording started

This phase:
- Ranks available LLMs
- Parses documentation
- Initializes agent pool
- Starts video recording

Phase 2: Doc-Driven Verification

[doc-driven][desktop] Verifying: 'User Login' (1/42)
[doc-driven][desktop]   ✓ Navigate to /login
[doc-driven][desktop]   ✓ Enter test credentials  
[doc-driven][desktop]   ✓ Submit form
[doc-driven][desktop]   ✓ Verify dashboard loads
[doc-driven][desktop] Feature verified: user-login

The system:
- Reads feature descriptions from docs
- Uses LLM to generate test steps
- Executes navigation actions
- Verifies expected outcomes
- Captures before/after screenshots

Phase 3: Curiosity-Driven Exploration

[curiosity][desktop] Exploring: Settings page
[curiosity][desktop]   Testing: Form validation
[curiosity][desktop]   Issue: Empty input accepted (medium)
[curiosity][desktop]   Evidence: screenshot-034.png

This phase:
- Discovers unvisited screens
- Tests edge cases
- Detects undocumented bugs
- Continues until budget exhausted

Phase 4: Report Generation

[report] Stopping recordings...
[report] Aggregating results...
[report] Coverage: 95.2%
[report] Issues: 3 found
[report] Writing reports to ./qa-results/

Generates:
- Markdown report with summary
- Individual tickets per issue
- Timeline with video timestamps
- Coverage analysis"
```

**[12:00-20:00] - Monitoring & Control**
```
"While a session runs, you can:

1. View Progress
   ./helixqa status --session-id abc123
   
   Shows:
   - Current phase
   - Features verified/total
   - Issues found
   - Time elapsed/remaining

2. Pause/Resume
   ./helixqa pause --session-id abc123
   ./helixqa resume --session-id abc123
   
   Useful for:
   - Debugging application issues
   - System maintenance
   - API rate limit management

3. Cancel
   ./helixqa cancel --session-id abc123
   
   Gracefully stops the session and generates 
   partial reports with collected evidence.

4. Real-time Logs
   tail -f qa-results/session.log
   
   Shows detailed LLM interactions, 
   navigation steps, and error traces.

[Show each command in terminal]

Monitoring Best Practices:
- Check progress after first 5 minutes
- Review issues immediately if critical severity
- Monitor API usage for cost control
- Ensure video recording is working"
```

**[20:00-27:00] - Interpreting Results**
```
"After session completion, check:

1. Executive Summary (qa-report.md)
   
   Coverage: 95.2%
   Duration: 1h 23m
   Platforms: desktop, web
   
   Issues by Severity:
   - Critical: 0
   - High: 1
   - Medium: 2
   - Low: 0

2. Individual Tickets (tickets/HQA-*.md)

   Example ticket:
   # HQA-0001: Login button not responding
   
   Severity: High | Platform: Desktop
   
   Steps to Reproduce:
   1. Navigate to /login
   2. Enter credentials
   3. Click login button
   
   Expected: Dashboard loads
   Actual: Nothing happens
   
   Evidence:
   - Screenshot: screenshots/001.png
   - Video: videos/desktop.mp4 @ 05:23
   
   LLM Analysis:
   Button event handler not bound. 
   Check login.js line 45.

3. Timeline (timeline.json)
   
   Chronological event log:
   - 00:00 - Session start
   - 02:15 - Feature verification: login
   - 05:23 - Issue detected: button
   - 07:45 - Feature verification: signup
   
4. Coverage Report
   
   Features Verified: 40/42
   - login: ✓
   - signup: ✓
   - password-reset: ✓
   - profile: ⚠ skipped (timeout)
   - settings: ✓

Interpreting Coverage:
- >90%: Excellent, most features tested
- 70-90%: Good, some features missed
- <70%: Review documentation completeness"
```

**[27:00-30:00] - Session Management**
```
"Managing multiple sessions:

List active sessions:
./helixqa list --status running

Compare sessions:
./helixqa compare session1 session2

Archive old results:
./helixqa archive --before 2026-03-01

Integration with CI/CD:

GitLab CI example:
```yaml
qa-autonomous:
  script:
    - ./helixqa autonomous \\
        --project . \\
        --platforms web \\
        --output qa-results
    - ./scripts/parse-qa-results.sh
  artifacts:
    paths:
      - qa-results/
```

Parse results script:
#!/bin/bash
CRITICAL=$(jq '.issues.critical' qa-results/summary.json)
if [ "$CRITICAL" -gt 0 ]; then
  echo "Critical issues found!"
  exit 1
fi

This fails the pipeline if critical issues exist.

In Module 4, we'll cover advanced features like 
custom strategies and multi-platform testing."
```

---

## Module 4: Advanced Features (35 minutes)

### Learning Objectives
- Implement custom verification strategies
- Configure multi-platform testing
- Use LLM navigation effectively
- Integrate with existing workflows

### Script

**[00:00-08:00] - Custom Strategies**
```
"When predefined strategies aren't sufficient, 
create custom ones.

Example: Mobile-First Strategy

Create pkg/strategy/mobile.go:

package strategy

import (
    "context"
    "digital.vasic.llmsverifier/pkg/strategy"
)

type MobileFirstStrategy struct {
    base strategy.VerificationStrategy
}

func (s *MobileFirstStrategy) Score(ctx context.Context, 
    model strategy.ModelInfo) (strategy.StrategyScore, error) {
    
    baseScore, _ := s.base.Score(ctx, model)
    
    // Bonus for mobile-optimized models
    if model.AvgLatencyMs < 1000 {
        baseScore.Overall += 0.1
    }
    
    // Prefer models with good mobile support
    if model.Provider == "anthropic" {
        baseScore.Overall += 0.05
    }
    
    return baseScore, nil
}

func (s *MobileFirstStrategy) Name() string {
    return "mobile-first"
}

Register the strategy:

strategy.Register("mobile-first", func() strategy.VerificationStrategy {
    return &MobileFirstStrategy{
        base: strategy.NewDefaultStrategy(),
    }
})

Use in .env:
LLMSVERIFIER_STRATEGY=mobile-first

Other use cases:
- Compliance-focused (audit trail requirements)
- Performance-critical (sub-second responses)
- Cost-optimized (cheapest viable model)
- Region-specific (data residency)"
```

**[08:00-16:00] - Multi-Platform Testing**
```
"Testing across platforms simultaneously:

Configuration:
HELIX_AUTONOMOUS_PLATFORMS=android,web,desktop
HELIX_AGENT_POOL_SIZE=6  # 2 per platform

Parallel Execution:

The system spawns workers for each platform:

Worker 1: Android (emulator-5554)
  - ADB commands
  - Logcat monitoring
  - APK installation

Worker 2: Web (Playwright)
  - Chromium browser
  - Page navigation
  - Console log capture

Worker 3: Desktop (X11)
  - xdotool automation
  - Window management
  - Screenshot capture

Each worker:
- Has dedicated agent from pool
- Records separate video
- Generates platform-specific tickets
- Updates shared coverage map

Platform-Specific Configuration:

Android:
HELIX_ANDROID_DEVICE=emulator-5554
HELIX_ANDROID_PACKAGE=com.example.app

Web:
HELIX_WEB_URL=http://localhost:3000
HELIX_WEB_BROWSER=chromium

Desktop:
HELIX_DESKTOP_PROCESS=myapp
HELIX_DESKTOP_DISPLAY=:1

Cross-Platform Issues:

When an issue appears on multiple platforms:

HQA-0001: Login fails [Android, Web]
HQA-0002: Login fails [Desktop]

System detects similarity and links tickets.

[Show example multi-platform session]

Coverage aggregation:
Android: 42 features, 2 issues
Web: 40 features, 1 issue  
Desktop: 38 features, 3 issues
Overall: 95% coverage, 6 unique issues"
```

**[16:00-24:00] - Advanced Navigation**
```
"LLM Navigation Deep Dive:

The navigator uses a graph-based approach:

NavigationGraph {
  Screens: [Login, Dashboard, Settings]
  Transitions: {
    Login → Dashboard: {action: "login"}
    Dashboard → Settings: {action: "click-settings"}
  }
}

When targeting a screen:

1. Check if path exists in graph
   ShortestPath(Login, Settings) 
   → [Login → Dashboard → Settings]

2. If path unknown, infer with LLM:
   
   Prompt: "How do I reach Settings from Login?"
   LLM: "Click menu icon, then Settings"
   
3. Execute actions with verification
   - Take screenshot before
   - Perform action
   - Take screenshot after
   - Verify state changed

4. Update graph with new path

Edge Cases:

Dynamic UI (loading states):
- Wait for element with retry
- Use vision to detect ready state
- Timeout after configured limit

Unexpected dialogs:
- Detect popup/modal
- Handle or dismiss
- Log intervention

Navigation failures:
- Retry with alternative path
- Escalate to human if stuck
- Continue with other features

Optimization Tips:

1. Seed initial navigation graph
   - Manually define common paths
   - Reduces LLM calls
   - Faster initial navigation

2. Cache successful paths
   - Reuse across sessions
   - Updates with new discoveries

3. Prioritize high-value features
   - Critical user flows first
   - Peripheral features last

[Show navigation graph visualization]"
```

**[24:00-30:00] - Integration Patterns**
```
"Integrating HelixQA into your workflow:

1. Pre-Commit Hook

#!/bin/bash
# .git/hooks/pre-commit

./helixqa autonomous \\
  --platforms web \\
  --timeout 5m \\
  --quick-mode

if [ $? -ne 0 ]; then
  echo "QA issues found!"
  exit 1
fi

2. Nightly Regression

# CI Pipeline
schedule:
  - cron: "0 2 * * *"  # 2 AM daily

script:
  - ./helixqa autonomous \\
      --platforms android,web,desktop \\
      --timeout 4h \\
      --output nightly-results
  - ./scripts/upload-results.sh nightly-results

3. PR Validation

# Only test changed areas
./helixqa autonomous \\
  --diff-from main \\
  --smart-selection \\
  --timeout 15m

4. Production Monitoring

# Background crash detection
./helixqa monitor \\
  --production \\
  --alert-webhook $SLACK_WEBHOOK \\
  --continuous

Custom Integrations:

JIRA Ticket Creation:
```python
# scripts/jira-sync.py
import json
import requests

with open('qa-results/tickets/HQA-0001.md') as f:
    ticket = parse_ticket(f.read())

requests.post('https://jira.company.com/rest/api/2/issue',
    json={
        'fields': {
            'project': {'key': 'QA'},
            'summary': ticket.title,
            'description': ticket.description,
            'issuetype': {'name': 'Bug'},
            'priority': map_severity(ticket.severity)
        }
    })
```

Slack Notifications:
```bash
# scripts/notify-slack.sh
SUMMARY=$(jq -r '.summary' qa-results/report.json)
curl -X POST $SLACK_WEBHOOK \
  -H 'Content-type: application/json' \
  --data '{"text":"QA Complete: '$SUMMARY'"}'
```

Dashboard Integration:
```javascript
// Grafana panel
fetch('/api/qa-results/latest')
  .then(r => r.json())
  .then(data => {
    updateCoverageChart(data.coverage);
    updateIssueTable(data.issues);
  });
```

[Show integration architecture diagram]"
```

**[30:00-35:00] - Performance Optimization**
```
"Optimizing session performance:

1. Parallel Execution

GOMAXPROCS=8
HELIX_WORKER_POOL_SIZE=4

Distributes work across CPU cores
Reduces session duration by 60%

2. Smart Agent Selection

Strategy: Fastest viable model
- Quick checks: Groq (fast, cheap)
- Visual analysis: Claude (vision)
- Complex reasoning: GPT-4 (quality)

3. Caching

Enable response caching:
HELIX_CACHE_RESULTS=true
HELIX_CACHE_TTL=1h

Avoids redundant LLM calls
Reduces API costs by 40%

4. Incremental Testing

# Only test changed features
./helixqa autonomous \\
  --incremental \\
  --since-last-run

5. Resource Management

Memory limits:
HELIX_MAX_MEMORY_MB=2048

CPU limits:
HELIX_MAX_GOROUTINES=20

Video compression:
HELIX_VIDEO_QUALITY=medium  # vs high

Benchmarking:

Run performance test:
./helixqa benchmark \\
  --duration 1h \\
  --load 100-features

Results:
- Avg feature verification: 12s
- LLM calls per feature: 2.3
- API cost per session: $2.50
- Coverage rate: 94%

Optimization Checklist:

□ Use appropriate strategy
□ Enable caching
□ Limit platforms if not needed
□ Adjust video quality
□ Parallelize agents
□ Monitor API costs
□ Review timeout settings

In Module 5, we'll cover troubleshooting 
common issues that may arise."
```

---

## Module 5: Troubleshooting (20 minutes)

### Learning Objectives
- Diagnose common setup issues
- Debug session failures
- Optimize for specific scenarios
- Handle API rate limits

### Script

**[00:00-05:00] - Common Setup Issues**
```
"Let's troubleshoot the most common issues:

Issue 1: 'No agents available'

Symptoms:
- Session fails immediately
- Error: 'no suitable agent found'

Diagnosis:
which opencode  # Check if installed
which claude    # Check if installed

Solutions:
1. Install missing agents:
   go install github.com/opencode-ai/opencode@latest
   
2. Check .env paths:
   HELIX_AGENT_OPENCODE_PATH=/usr/local/bin/opencode
   
3. Verify API keys:
   echo $ANTHROPIC_API_KEY  # Should not be empty
   
4. Test agent manually:
   opencode --version

Issue 2: 'Vision provider failed'

Symptoms:
- Screenshots not analyzed
- Error: 'vision analysis failed'

Diagnosis:
- Check API key validity
- Verify model supports vision
- Check rate limits

Solutions:
1. Use vision-capable model:
   ANTHROPIC_MODEL=claude-3-opus-20240229
   
2. Fallback to non-vision:
   HELIX_VISION_PROVIDER=none
   
3. Check API quota:
   curl https://api.anthropic.com/v1/models \\
     -H "x-api-key: $KEY"

Issue 3: 'ADB device not found'

Symptoms:
- Android tests fail
- Error: 'device offline'

Solutions:
1. Check device connection:
   adb devices
   
2. Authorize device:
   adb kill-server
   adb start-server
   adb devices  # Accept RSA key
   
3. Update device ID in .env:
   HELIX_ANDROID_DEVICE=emulator-5554

[Show troubleshooting flowchart]"
```

**[05:00-12:00] - Session Debugging**
```
"When sessions fail or produce unexpected results:

Debug Mode:

./helixqa autonomous \\
  --verbose \\
  --debug \\
  --log-level debug

This shows:
- Every LLM prompt and response
- Navigation decisions
- Screenshot analysis details
- Timing information

Analyzing Logs:

# Filter for errors
grep "ERROR" qa-results/session.log

# Find slow operations
grep "duration" qa-results/session.log | sort -k2

# Track agent usage
grep "agent" qa-results/session.log | wc -l

Common Log Patterns:

Pattern: 'timeout waiting for response'
Cause: LLM API slow or rate limited
Fix: Increase HELIX_AGENT_TIMEOUT

Pattern: 'navigation failed after 3 retries'
Cause: UI changed or element not found
Fix: Update selectors or use vision

Pattern: 'screenshot capture failed'
Cause: Display not available or ffmpeg missing
Fix: Check display, install ffmpeg

Interactive Debugging:

Pause session at specific point:

./helixqa autonomous \\
  --breakpoint "after-login" \\
  --interactive

In interactive mode:
- Inspect current state
- Manually trigger actions
- Review screenshots
- Continue or abort

Session Replay:

./helixqa replay \\
  --session-id abc123 \\
  --from-timeline \\
  --speed 2x

Replays actions from timeline
Useful for reproducing issues"
```

**[12:00-18:00] - API Rate Limits & Costs**
```
"Managing API usage and costs:

Rate Limit Handling:

When you see:
'429 Too Many Requests'

HelixQA automatically:
1. Backs off exponentially
2. Retries with jitter
3. Switches to fallback provider
4. Logs rate limit events

Manual Rate Limit Configuration:

ANTHROPIC_RATE_LIMIT=40  # requests per minute
ANTHROPIC_RETRY_MAX=5
ANTHROPIC_RETRY_DELAY=2s

Cost Optimization:

Monitor costs:
./helixqa costs --session-id abc123

Output:
Provider      Tokens    Cost
Anthropic     45,230    $2.34
OpenAI        12,100    $0.89
Total                   $3.23

Cost Reduction Strategies:

1. Use cheaper models for simple tasks:
   - GPT-3.5 for text analysis
   - Claude Haiku for quick checks
   
2. Enable caching:
   HELIX_CACHE_RESULTS=true
   Cache hit rate: ~40%
   
3. Reduce screenshot frequency:
   HELIX_SCREENSHOT_INTERVAL=5s  # vs 1s
   
4. Shorter timeouts:
   HELIX_AGENT_TIMEOUT=30s  # vs 120s
   
5. Use vision selectively:
   HELIX_VISION_ENABLED=phases=curiosity-only

Budget Alerts:

Set spending limit:
HELIX_MAX_COST_PER_SESSION=5.00

Alert at 80%:
HELIX_COST_ALERT_THRESHOLD=0.8

Webhook notification:
HELIX_COST_ALERT_WEBHOOK=https://hooks.slack.com/...

Multi-Provider Load Balancing:

Distribute load across providers:

Provider      Weight    Use Case
Anthropic     50%       Vision tasks
OpenAI        30%       Text analysis
Google        20%       Fallback

Automatic failover when:
- Rate limit hit
- Service unavailable
- Cost threshold reached

[Show cost monitoring dashboard]"
```

**[18:00-20:00] - Getting Help & Resources**
```
"When you're stuck:

Documentation:
- README.md - Quick start
- docs/USER_GUIDE.md - Detailed usage
- docs/ARCHITECTURE_DIAGRAMS.md - System design
- docs/TROUBLESHOOTING.md - Common issues

Community Resources:
- GitHub Issues: github.com/vasic-digital/Catalogizer/issues
- Discussions: GitHub Discussions tab
- Wiki: Project wiki with examples

Debug Information to Collect:

When reporting issues, include:
1. Session logs: qa-results/session.log
2. Configuration: .env (redact API keys)
3. System info: OS, Go version
4. Command used: Full helixqa command
5. Expected vs actual behavior

Example Report:

Title: Session timeout on Android testing

Description:
Running autonomous session on Android emulator.
Session times out after 10 minutes with 20% coverage.

Environment:
- OS: Ubuntu 22.04
- Go: 1.24
- HelixQA: v1.0.0

Command:
./helixqa autonomous --platforms android --timeout 30m

Logs:
[ERROR] adb shell timeout
[ERROR] screenshot capture failed

Expected: Complete testing in 30m
Actual: Timeout at 10m

Next Steps:

Now that you've completed this course, you can:

1. Set up HelixQA for your project
2. Configure custom strategies
3. Integrate with CI/CD
4. Monitor and optimize performance
5. Contribute to the project

Thank you for watching!

For updates and advanced topics, 
subscribe to the channel and star the repo.

Happy autonomous testing!"
```

---

## Production Notes

### Recording Setup
- **Resolution:** 1920x1080 minimum
- **Frame Rate:** 30fps
- **Audio:** Clear narration, minimal background noise
- **Terminal:** Large font (16pt+), high contrast theme

### Visual Aids
- Use callouts for important commands
- Highlight configuration files
- Show split-screen for code + terminal
- Include zoom-in on critical details

### Interactive Elements
- Pause points for viewer practice
- Quiz questions at module ends
- Downloadable configuration templates
- Cheat sheet PDF

### Accessibility
- Closed captions for all videos
- Transcript available
- Keyboard shortcuts documented
- Screen reader compatible code
