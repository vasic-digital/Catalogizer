# Installation Wizard -- Tauri/Rust + React Course

**Component**: installer-wizard
**Language**: Rust (backend) / TypeScript + React (frontend)
**Total Duration**: 2 hours (4 modules)
**Level**: Intermediate

---

## Course Overview

This course covers the complete architecture of the Catalogizer Installation Wizard, a Tauri-based desktop application that guides users through initial setup and storage root configuration. You will learn the multi-step wizard flow with validation and navigation, protocol-specific configuration for SMB/FTP/NFS/WebDAV/Local sources, network scanning for automatic service discovery, and deployment of the final configuration to catalog-api.

---

### Module 1: Wizard Flow

**Duration**: 35 minutes
**Prerequisites**: Basic React/TypeScript, familiarity with Tauri IPC from the desktop course

#### Learning Objectives
- Understand the multi-step wizard architecture with forward/backward navigation and per-step validation
- Trace the state management through `WizardContext` and `ConfigurationContext`
- Explain how the wizard coordinates between React step components and the Rust backend
- Navigate the project structure: `src/components/wizard/` for steps, `src/contexts/` for state, `src-tauri/src/` for backend

#### Topics Covered
1. **Wizard context (`src/contexts/WizardContext.tsx`)**
   - Current step tracking with forward/backward navigation
   - Step completion state: each step reports whether it is valid and complete
   - Navigation guards: forward navigation blocked until current step passes validation
   - Step history for back-button support with state preservation
2. **Configuration context (`src/contexts/ConfigurationContext.tsx`)**
   - Accumulated configuration state built across all wizard steps
   - Protocol selection, credentials, paths, scan settings gathered progressively
   - Context persistence: partial configuration survives window close and resume
   - Validation state per field with error messages
3. **Step components (`src/components/wizard/`)**
   - `WelcomeStep.tsx`: introduction, system requirements check, license acceptance
   - `ProtocolSelectionStep.tsx`: choose storage protocol (SMB, FTP, NFS, WebDAV, Local)
   - Protocol-specific steps: `SMBConfigurationStep.tsx`, `FTPConfigurationStep.tsx`, `NFSConfigurationStep.tsx`, `WebDAVConfigurationStep.tsx`, `LocalConfigurationStep.tsx`
   - `NetworkScanStep.tsx`: automatic discovery of available services on the local network
   - `ConfigurationManagementStep.tsx`: review and manage all configured storage roots
   - `SummaryStep.tsx`: final review of all settings before deployment
4. **Layout and navigation (`src/components/layout/`)**
   - Wizard chrome: step indicator, progress bar, back/next/finish buttons
   - Responsive layout adapting to window size
   - Branded header with Vasic Digital identity
5. **Splash screen (`src/components/SplashScreen.tsx`)**
   - Animated loading screen during Tauri initialization
   - Backend readiness check before rendering the first wizard step

#### Hands-On Exercise
Run the wizard with `npm run tauri:dev` and navigate through all steps forward and backward. Observe how the WizardContext tracks step completion state by inspecting React DevTools. Partially fill a step, navigate backward, then forward again and verify the entered data is preserved. Close the wizard window, reopen it, and verify configuration context persistence.

#### Key Takeaways
- The wizard uses dual contexts: WizardContext for navigation flow, ConfigurationContext for accumulated settings
- Forward navigation is gated by step validation -- users cannot skip ahead without completing the current step
- Step state is preserved during backward navigation so users do not lose entered data
- Each protocol has a dedicated step component with protocol-specific fields and validation rules

---

### Module 2: Protocol Configuration

**Duration**: 35 minutes
**Prerequisites**: Module 1, understanding of network storage protocols

#### Learning Objectives
- Build protocol-specific configuration forms with field validation for each supported protocol
- Implement connection testing via Rust IPC commands that verify credentials and path accessibility
- Handle protocol-specific edge cases: SMB share permissions, FTP passive mode, NFS export restrictions
- Apply the factory pattern in the Rust backend for protocol-specific validation logic

#### Topics Covered
1. **SMB configuration (`src/components/wizard/SMBConfigurationStep.tsx`)**
   - Fields: server hostname/IP, share name, username, password, domain (optional)
   - Connection test: Rust backend (`src-tauri/src/smb.rs`) attempts to list the share root
   - Common issues: guest access, NTLM vs Kerberos authentication, SMBv1 vs SMBv2/v3
   - Path format: `smb://server/share` URI construction from individual fields
2. **FTP configuration (`src/components/wizard/FTPConfigurationStep.tsx`)**
   - Fields: hostname, port (default 21), username, password, base path, TLS toggle
   - Passive mode configuration for NAT/firewall traversal
   - Connection test: Rust backend (`src-tauri/src/ftp.rs`) authenticates and lists the base directory
   - FTPS support: explicit TLS on port 21 vs implicit TLS on port 990
3. **NFS configuration (`src/components/wizard/NFSConfigurationStep.tsx`)**
   - Fields: server hostname/IP, export path, mount options
   - Export access verification via Rust backend (`src-tauri/src/nfs.rs`)
   - Platform considerations: native NFS on Linux, limited support on macOS, not available on Windows
   - Permission model: UID/GID mapping between client and server
4. **WebDAV configuration (`src/components/wizard/WebDAVConfigurationStep.tsx`)**
   - Fields: server URL, username, password, base path
   - HTTPS certificate validation with option to accept self-signed certificates
   - Connection test: Rust backend (`src-tauri/src/webdav.rs`) performs PROPFIND on the base path
   - Authentication methods: Basic, Digest, Bearer token
5. **Local configuration (`src/components/wizard/LocalConfigurationStep.tsx`)**
   - Native file dialog for directory selection via Tauri's dialog API
   - Path validation: directory exists, is readable, has sufficient permissions
   - Rust backend (`src-tauri/src/local.rs`) verifying path accessibility
   - Recursive scan depth configuration
6. **Rust backend validation (`src-tauri/src/`)**
   - Per-protocol Rust modules: `smb.rs`, `ftp.rs`, `nfs.rs`, `webdav.rs`, `local.rs`
   - IPC commands for connection testing invoked from each step component
   - Timeout handling: connection tests abort after configurable timeout
   - Error reporting: structured error types with user-friendly messages

#### Hands-On Exercise
Configure an SMB storage root by entering a NAS server address, share name, and credentials. Click "Test Connection" and observe the Rust backend log output as it attempts to connect. Intentionally enter wrong credentials and verify the error message is specific and helpful. Configure a local directory using the native file picker and verify path validation.

#### Key Takeaways
- Each protocol has a dedicated Rust backend module that handles connection testing independently
- Connection tests verify not just connectivity but actual read access to the configured path
- Protocol-specific edge cases (FTP passive mode, NFS UID mapping, WebDAV self-signed certs) are surfaced with clear guidance
- The native file dialog for local paths provides a familiar OS-native experience instead of a web-based file browser

---

### Module 3: Network Scanning

**Duration**: 25 minutes
**Prerequisites**: Module 1, Module 2, basic networking concepts

#### Learning Objectives
- Implement network service discovery using the Rust backend's scanning capabilities
- Display discovered services with protocol type, hostname, and accessibility status
- Enable one-click configuration from discovered services to pre-populated protocol steps
- Handle scan timeouts and partial results gracefully

#### Topics Covered
1. **Network scan step (`src/components/wizard/NetworkScanStep.tsx`)**
   - Scan trigger button with progress indicator
   - Results list showing discovered services: hostname, IP, protocol, port, share/export names
   - Status indicators: accessible (green), requires credentials (yellow), unreachable (red)
   - "Configure" action on each result pre-populating the appropriate protocol step
2. **Rust scanning backend (`src-tauri/src/network.rs`)**
   - Subnet scanning: probe local network range for common service ports
   - SMB discovery: port 445 probe, share enumeration on responsive hosts
   - FTP discovery: port 21 probe with banner detection
   - NFS discovery: showmount equivalent for export enumeration
   - WebDAV discovery: HTTP/HTTPS probing with OPTIONS method
   - mDNS/Bonjour: discovery of advertised services on the local network
3. **Scan configuration**
   - Subnet range: auto-detected from the host's network interface, editable by user
   - Timeout per host: configurable probe timeout (default 2 seconds)
   - Concurrent probes: parallel scanning with concurrency limit to avoid network flooding
   - Protocol filter: scan for specific protocols or all supported protocols
4. **Result handling**
   - Progressive results: services appear in the list as they are discovered, not after full scan completion
   - Deduplication: same host with multiple protocols shown as separate entries
   - Persistence: scan results cached for the wizard session so users can review without rescanning
   - Empty state: helpful message when no services are found with troubleshooting suggestions

#### Hands-On Exercise
Run a network scan from the wizard and observe services being discovered progressively. Click "Configure" on a discovered SMB share and verify the protocol step is pre-populated with the host and share name. Modify the subnet range to scan a different network segment. Examine the Rust backend logs to see the probe sequence and timing.

#### Key Takeaways
- Network scanning auto-discovers storage services, eliminating the need for users to know IP addresses and share names
- Progressive result display provides immediate feedback during what can be a 30+ second scan
- One-click configuration from scan results reduces manual entry errors and speeds up setup
- The Rust backend performs scanning efficiently with concurrent probes and per-host timeouts

---

### Module 4: Deployment

**Duration**: 25 minutes
**Prerequisites**: Modules 1-3

#### Learning Objectives
- Generate the final configuration from accumulated wizard state and write it to catalog-api's config format
- Register configured storage roots with the running catalog-api instance via REST API
- Display a deployment summary with status indicators for each configured source
- Handle deployment failures with retry and rollback options

#### Topics Covered
1. **Summary step (`src/components/wizard/SummaryStep.tsx`)**
   - Complete review of all configured storage roots with protocol, host, path details
   - Credential summary (masked passwords) for verification
   - Scan settings: initial scan trigger, scan schedule configuration
   - "Deploy" button initiating the configuration write and API registration
2. **Configuration writing**
   - Rust backend serializing the configuration to catalog-api's expected format
   - Environment variable generation for `.env` file (database type, JWT secret, API keys)
   - Storage root entries formatted for the catalog-api configuration file
   - File permissions: configuration written with restrictive permissions (0600)
3. **API registration**
   - REST API calls to catalog-api's `/api/v1/storage-roots` endpoint
   - Per-source registration with success/failure tracking
   - Authentication: wizard obtains an admin token before registration
   - Initial scan trigger: optional immediate scan of newly registered storage roots
4. **Deployment status and error handling**
   - Per-source status indicators: pending, deploying, success, failed
   - Error details: specific failure reason for each failed source (connection refused, auth failed, path not found)
   - Retry failed sources individually without redeploying successful ones
   - Rollback: remove successfully registered sources if the user cancels after partial failure
5. **Post-deployment**
   - Configuration management step (`ConfigurationManagementStep.tsx`) for ongoing modifications
   - Launch catalog-web in the default browser after successful deployment
   - Wizard state cleanup: clear sensitive data (credentials) from memory and temporary storage
   - Option to re-run the wizard to add additional storage roots later

#### Hands-On Exercise
Complete the full wizard flow: welcome, protocol selection, configuration with connection test, summary review, and deployment. Verify that the storage root appears in catalog-api by querying the REST API. Trigger an initial scan from the summary step and monitor progress. Intentionally misconfigure a source, deploy, observe the failure, correct the configuration, and retry.

#### Key Takeaways
- The deployment step translates accumulated wizard state into catalog-api configuration and API registrations
- Per-source status tracking provides granular feedback: users see exactly which sources succeeded and which failed
- Retry and rollback handle partial deployment failures without requiring a full restart of the wizard
- Post-deployment cleanup removes sensitive credential data from wizard memory and temporary storage
