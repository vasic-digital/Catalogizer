# Device Connect (.devconnect) Guide

## Overview

The `.devconnect` feature provides automatic device connection management for HelixQA testing. It ensures Android TV devices are always connected before QA sessions begin, eliminating manual `adb connect` steps.

## What is .devconnect?

`.devconnect` is a configuration file (opposite of `.devignore`) that lists IP addresses of Android devices you **want** to keep connected for testing.

### Key Features

- **Auto-connect**: Devices are automatically connected via `adb connect`
- **Reachability validation**: Pings devices before attempting connection
- **Idempotent**: Safe to run multiple times - skips already connected devices
- **Git-ignored**: Local IP addresses are never committed to the repository
- **Validation**: Ensures devices are responsive before marking as connected

## Quick Start

### 1. Create .devconnect File

```bash
# Create .devconnect with your Android TV device IP
echo "192.168.0.214" > .devconnect
```

### 2. Run devconnect Script

```bash
./scripts/devconnect.sh
```

### 3. Verify Connection

```bash
adb devices
```

Output:
```
List of devices attached
192.168.0.214:5555    device
```

## File Format

The `.devconnect` file uses a simple line-based format:

```
# Device Connect List for HelixQA
# ==============================
# IP addresses listed here will be auto-connected via 'adb connect'

# Android TV - Mi Box 4
192.168.0.214

# Another device with explicit port
192.168.0.215:5556

# Tablet
192.168.0.100
```

### Format Rules

| Format | Description | Example |
|--------|-------------|---------|
| `IP` | IP address with default port 5555 | `192.168.0.214` |
| `IP:PORT` | IP address with explicit port | `192.168.0.214:5555` |
| `# comment` | Comments (ignored) | `# Mi Box 4` |
| Empty lines | Ignored | (blank line) |

## The devconnect.sh Script

### Usage

```bash
./scripts/devconnect.sh [OPTIONS]
```

### Options

| Option | Description |
|--------|-------------|
| (none) | Use default `.devconnect` file |
| `DEVCONNECT_FILE=path` | Use custom file path |

### Environment Variables

```bash
# Use custom device list file
DEVCONNECT_FILE=/path/to/my-devices.txt ./scripts/devconnect.sh
```

### Output

The script provides color-coded output:

- **Green [INFO]**: Success messages
- **Yellow [WARN]**: Warnings (non-fatal)
- **Red [ERROR]**: Errors (fatal)

Example output:
```
[INFO] Device Connect Script for HelixQA
[INFO] ==================================
[INFO] Reading device list from: .devconnect

[INFO] Processing device: 192.168.0.214:5555
[INFO] ✓ Device 192.168.0.214 is reachable
[INFO] ✓ Successfully connected to 192.168.0.214:5555
[INFO] ✓ Device 192.168.0.214:5555 is responsive

[INFO] ==================================
[INFO] Device Connect Summary:
[INFO]   Success: 1
[INFO]   Failed:  0

[INFO] Currently connected ADB devices:
192.168.0.214:5555    device
```

## Integration with HelixQA

### Pre-Flight Checklist

Before running HelixQA, always:

```bash
# 1. Check .devignore - ensure devices are NOT excluded
grep -i "mibox\|android" .devignore || echo "✓ No exclusions"

# 2. Check .devconnect - ensure devices ARE listed
cat .devconnect | grep -v "^#" | grep -v "^$"

# 3. Connect devices
./scripts/devconnect.sh

# 4. Verify connection
adb devices

# 5. Run HelixQA
./HelixQA/bin/helixqa autonomous -platforms android ...
```

### Automated Integration

The `helixqa-orchestrator.sh` script automatically runs `devconnect.sh`:

```bash
# One command handles everything
./scripts/helixqa-orchestrator.sh android
```

This executes:
1. Environment validation
2. `devconnect.sh` (auto-connect devices)
3. APK installation
4. HelixQA execution

## Troubleshooting

### Device Not Reachable

```
[ERROR] ✗ Device 192.168.0.214 is not reachable (ping failed)
```

**Solutions:**
- Verify device is powered on
- Check device is on same network
- Verify ADB over network is enabled on device
- Check firewall rules

### ADB Connection Refused

```
[ERROR] ✗ ADB connect failed for 192.168.0.214:5555
```

**Solutions:**
- Enable "ADB over network" in Android TV developer options
- Restart ADB daemon: `adb kill-server && adb start-server`
- Check device IP hasn't changed

### Device in .devignore

```
[WARN]  Skipping 192.168.0.214 - device ATMOSphere is in .devignore
```

**This is expected behavior.** Devices in `.devignore` are intentionally excluded from testing.

## Best Practices

### 1. Keep .devconnect Updated

```bash
# Add new device
echo "192.168.0.215" >> .devconnect

# Remove device
sed -i '/192.168.0.215/d' .devconnect
```

### 2. Use Static IPs

Configure your router to assign static IPs to Android TV devices:
- Mi Box 4: `192.168.0.214`
- Shield TV: `192.168.0.215`

### 3. Validate Before QA

Add to your CI/CD pipeline:

```yaml
# Example GitLab CI
qa:android:
  script:
    - ./scripts/devconnect.sh
    - adb devices
    - ./HelixQA/bin/helixqa autonomous -platforms android
```

### 4. Multiple Devices

For testing multiple devices:

```bash
# .devconnect
192.168.0.214  # Mi Box 4 - Living room
192.168.0.215  # Shield TV - Bedroom
192.168.0.216  # Chromecast - Kitchen
```

All devices will be connected automatically.

## Architecture

```
┌─────────────────┐
│   .devconnect   │
│   (IP list)     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  devconnect.sh  │
│                 │
│  1. Parse IPs   │
│  2. Ping test   │
│  3. ADB connect │
│  4. Verify      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   ADB Devices   │
│   (connected)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    HelixQA      │
│   (autonomous   │
│     testing)    │
└─────────────────┘
```

## Comparison: .devconnect vs .devignore

| Feature | .devconnect | .devignore |
|---------|-------------|------------|
| Purpose | Include devices | Exclude devices |
| Action | `adb connect` | Skip testing |
| Content | IP addresses | Device name patterns |
| Example | `192.168.0.214` | `ATMOSphere` |
| Git tracked | No (ignored) | Yes |
| When used | Before HelixQA | During HelixQA |

## Reference

### Files

| File | Purpose |
|------|---------|
| `.devconnect` | Device IP list (git-ignored) |
| `scripts/devconnect.sh` | Auto-connect script |
| `.devignore` | Device exclusion patterns |

### Commands

```bash
# Quick reference
./scripts/devconnect.sh                    # Connect all devices
adb devices                                # List connected devices
adb connect 192.168.0.214:5555            # Manual connect
adb disconnect 192.168.0.214:5555         # Disconnect
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All devices connected successfully |
| 1 | One or more devices failed to connect |

## See Also

- [HelixQA Orchestrator Guide](./HELIXQA_ORCHESTRATOR_GUIDE.md)
- [Testing Guide](./TESTING_GUIDE.md)
- [AGENTS.md](../../AGENTS.md) - Agent constraints
