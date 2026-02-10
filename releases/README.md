# Catalogizer Releases

This directory contains production-ready builds organized by platform and application.

## Directory Structure

```
releases/
├── linux/
│   ├── catalog-api/           # Backend API binary
│   ├── catalog-web/            # Frontend static files
│   ├── catalogizer-desktop/   # Desktop application
│   └── installer-wizard/       # Installation wizard
├── windows/
│   ├── catalog-api/           # Backend API .exe
│   ├── catalog-web/           # Frontend static files
│   ├── catalogizer-desktop/  # Desktop application .exe
│   └── installer-wizard/      # Installation wizard .exe
├── macos/
│   ├── catalog-api/           # Backend API binary
│   ├── catalog-web/           # Frontend static files
│   ├── catalogizer-desktop/  # Desktop .app bundle
│   └── installer-wizard/      # Installation wizard .app
└── android/
    ├── catalogizer-android/   # Android APK/AAB
    └── catalogizer-androidtv/ # Android TV APK/AAB
```

## Build Instructions

### Backend API (Go)

```bash
# Linux
cd catalog-api
GOOS=linux GOARCH=amd64 go build -o ../releases/linux/catalog-api/catalog-api main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o ../releases/windows/catalog-api/catalog-api.exe main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o ../releases/macos/catalog-api/catalog-api main.go
```

### Frontend (React)

```bash
cd catalog-web
npm run build
cp -r dist/* ../releases/linux/catalog-web/
cp -r dist/* ../releases/windows/catalog-web/
cp -r dist/* ../releases/macos/catalog-web/
```

### Desktop Application (Tauri)

```bash
cd catalogizer-desktop

# Linux
npm run tauri:build -- --target x86_64-unknown-linux-gnu
cp src-tauri/target/release/bundle/appimage/*.AppImage ../releases/linux/catalogizer-desktop/

# Windows
npm run tauri:build -- --target x86_64-pc-windows-msvc
cp src-tauri/target/release/bundle/msi/*.msi ../releases/windows/catalogizer-desktop/

# macOS
npm run tauri:build -- --target x86_64-apple-darwin
cp -r src-tauri/target/release/bundle/dmg/*.dmg ../releases/macos/catalogizer-desktop/
```

### Installer Wizard (Tauri)

```bash
cd installer-wizard

# Linux
npm run tauri:build
cp src-tauri/target/release/bundle/appimage/*.AppImage ../releases/linux/installer-wizard/

# Windows
npm run tauri:build
cp src-tauri/target/release/bundle/msi/*.msi ../releases/windows/installer-wizard/

# macOS
npm run tauri:build
cp -r src-tauri/target/release/bundle/dmg/*.dmg ../releases/macos/installer-wizard/
```

### Android Applications

```bash
# catalogizer-android
cd catalogizer-android
./gradlew assembleRelease
./gradlew bundleRelease
cp app/build/outputs/apk/release/*.apk ../releases/android/catalogizer-android/
cp app/build/outputs/bundle/release/*.aab ../releases/android/catalogizer-android/

# catalogizer-androidtv
cd catalogizer-androidtv
./gradlew assembleRelease
cp app/build/outputs/apk/release/*.apk ../releases/android/catalogizer-androidtv/
```

## Automated Release Script

Use the automated script to build all platforms:

```bash
./scripts/build-all-releases.sh
```

## Version Naming Convention

All release files should follow this naming pattern:

```
{app-name}-v{version}-{platform}-{arch}.{ext}

Examples:
- catalog-api-v1.0.0-linux-amd64
- catalogizer-desktop-v1.0.0-windows-x64.exe
- catalogizer-android-v1.0.0-release.apk
```

## Release Checklist

Before creating a release:

- [ ] All tests pass (`./scripts/run-all-tests.sh`)
- [ ] Security scans clean (Snyk, SonarQube)
- [ ] Version numbers updated in:
  - [ ] catalog-api/main.go
  - [ ] catalog-web/package.json
  - [ ] catalogizer-desktop/package.json
  - [ ] catalogizer-android/app/build.gradle
- [ ] CHANGELOG.md updated
- [ ] Documentation updated
- [ ] Database migrations tested
- [ ] Configuration examples provided

## Distribution

### GitHub Releases

Upload all platform builds to GitHub Releases with:
- Release notes
- SHA256 checksums
- GPG signatures

### Direct Download

Host releases at:
- https://releases.catalogizer.com/{version}/{platform}/

### Package Managers

- **Linux:** AUR, Flatpak, Snap
- **Windows:** Winget, Chocolatey
- **macOS:** Homebrew
- **Android:** Google Play Store, F-Droid

## Security

All release binaries are:
- Code-signed with official certificates
- Verified with SHA256 checksums
- Scanned for vulnerabilities
- Built in clean CI/CD environment

## Support

For release issues, contact:
- GitHub Issues: https://github.com/your-org/catalogizer/issues
- Email: support@catalogizer.com
