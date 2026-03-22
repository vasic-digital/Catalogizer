# HelixQA Autonomous QA Session - Architecture Diagrams

## System Architecture Overview

```mermaid
graph TB
    subgraph "HelixQA Core"
        CMD[CLI Command]
        COORD[SessionCoordinator]
        PM[PhaseManager]
    end
    
    subgraph "Phase 1: Setup"
        SETUP_LLMS[LLMsVerifier<br/>Strategy Selection]
        SETUP_DOC[DocProcessor<br/>Feature Extraction]
        SETUP_AGENT[LLMOrchestrator<br/>Agent Pool]
    end
    
    subgraph "Phase 2: Doc-Driven"
        NAV[LLMNavigator<br/>Path Inference]
        WORKER[PlatformWorker]
        EXEC[ActionExecutor]
    end
    
    subgraph "Phase 3: Curiosity"
        EXPLORE[NavigationGraph<br/>Screen Discovery]
        DETECT[IssueDetector<br/>LLM Analysis]
    end
    
    subgraph "Phase 4: Report"
        TICKET[TicketGenerator]
        REPORT[Reporter]
        EVIDENCE[EvidenceCollector]
    end
    
    subgraph "External Services"
        LLM1[Anthropic API]
        LLM2[OpenAI API]
        LLM3[Google API]
    end
    
    subgraph "Target Applications"
        ANDROID[Android App<br/>ADB]
        WEB[Web App<br/>Playwright]
        DESKTOP[Desktop App<br/>X11]
    end
    
    CMD --> COORD
    COORD --> PM
    
    PM --> SETUP_LLMS
    PM --> SETUP_DOC
    PM --> SETUP_AGENT
    
    SETUP_LLMS --> LLM1
    SETUP_DOC --> LLM1
    SETUP_AGENT --> LLM1
    
    PM --> WORKER
    WORKER --> NAV
    NAV --> EXEC
    EXEC --> ANDROID
    EXEC --> WEB
    EXEC --> DESKTOP
    
    PM --> EXPLORE
    EXPLORE --> DETECT
    DETECT --> LLM1
    
    PM --> TICKET
    PM --> REPORT
    PM --> EVIDENCE
    
    DETECT --> TICKET
    WORKER --> EVIDENCE
```

## Component Interaction Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI as HelixQA CLI
    participant Coord as SessionCoordinator
    participant Verifier as LLMsVerifier
    participant Pool as AgentPool
    participant Worker as PlatformWorker
    participant Navigator as LLMNavigator
    participant Detector as IssueDetector
    participant TicketGen as TicketGenerator
    
    User->>CLI: Run autonomous session
    CLI->>Coord: StartSession(config)
    
    rect rgb(200, 255, 200)
        Note over Coord,Verifier: Phase 1: Setup
        Coord->>Verifier: SelectBestModels(requirements)
        Verifier-->>Coord: Ranked Models
        Coord->>Pool: InitializeAgents(models)
        Pool-->>Coord: Agent Pool Ready
    end
    
    rect rgb(255, 255, 200)
        Note over Coord,Worker: Phase 2: Doc-Driven
        loop For Each Feature
            Coord->>Worker: VerifyFeature(feature)
            Worker->>Navigator: NavigateTo(target)
            Navigator->>Pool: GetAgent()
            Pool-->>Navigator: Agent
            Navigator->>Navigator: InferPath()
            Navigator->>Worker: ExecuteActions()
        end
    end
    
    rect rgb(255, 200, 200)
        Note over Coord,Detector: Phase 3: Curiosity
        loop Explore Unvisited
            Worker->>Navigator: ExploreUnknown()
            Worker->>Detector: AnalyzeScreen()
            Detector->>Pool: GetAgent()
            Pool-->>Detector: Agent
            Detector->>Detector: DetectIssues()
        end
    end
    
    rect rgb(200, 200, 255)
        Note over Coord,TicketGen: Phase 4: Report
        Coord->>Detector: GetIssues()
        Detector-->>Coord: Issues[]
        loop For Each Issue
            Coord->>TicketGen: GenerateTicket(issue)
            TicketGen->>Pool: GetAgent()
            Pool-->>TicketGen: Agent
            TicketGen->>TicketGen: SuggestFix()
        end
        Coord->>User: GenerateReport()
    end
```

## Strategy Pattern Architecture

```mermaid
classDiagram
    class VerificationStrategy {
        <<interface>>
        +Name() string
        +Score(ctx, model) StrategyScore
        +Validate(ctx, model) ValidationResult
        +Rank(ctx, models) []RankedModel
        +Select(ctx, ranked, req) ModelInfo
    }
    
    class DefaultStrategy {
        -weights map[string]float64
        -constraints []Constraint
        -fallbacks []FallbackRule
        +SetWeights(weights)
        +SetConstraints(constraints)
        +SetFallbacks(fallbacks)
    }
    
    class QAStrategy {
        -visionWeight float64
        -speedWeight float64
        -qualityWeight float64
        +SetTestContext(context, weight)
    }
    
    class Recipe {
        +ID string
        +Name string
        +Strategy VerificationStrategy
        +Constraints []Constraint
        +Weights map[string]float64
        +Apply(strategy)
    }
    
    class RecipeBuilder {
        +WithName(name)
        +WithStrategy(strategy)
        +WithWeight(dimension, weight)
        +WithConstraint(constraint)
        +Build() Recipe
    }
    
    VerificationStrategy <|.. DefaultStrategy
    VerificationStrategy <|.. QAStrategy
    Recipe o-- VerificationStrategy
    RecipeBuilder ..> Recipe : creates
```

## Agent Pool Architecture

```mermaid
classDiagram
    class AgentPool {
        <<interface>>
        +Register(agent) error
        +Acquire(ctx, requirements) Agent
        +Release(agent)
        +Available() []Agent
        +HealthCheck(ctx) []HealthStatus
    }
    
    class MultiProviderPool {
        -pools map[string]AgentPool
        -selector AgentSelector
        +Acquire(ctx, req) Agent
    }
    
    class AgentSelector {
        <<interface>>
        +Select(pools, req) string
    }
    
    class RoundRobinSelector {
        -counter int
        +Select(pools, req) string
    }
    
    class PreferenceSelector {
        -preferredOrder []string
        +Select(pools, req) string
    }
    
    class OpenCodeAdapter {
        -config OpenCodeConfig
        -cmd *exec.Cmd
        +Start(ctx) error
        +Send(ctx, prompt) Response
    }
    
    class ClaudeCodeAdapter {
        -config ClaudeConfig
        +Start(ctx) error
        +Send(ctx, prompt) Response
    }
    
    AgentPool <|.. MultiProviderPool
    AgentSelector <|.. RoundRobinSelector
    AgentSelector <|.. PreferenceSelector
    MultiProviderPool o-- AgentSelector
    MultiProviderPool o-- AgentPool
    AgentPool o-- OpenCodeAdapter
    AgentPool o-- ClaudeCodeAdapter
```

## Navigation Engine Flow

```mermaid
flowchart TD
    Start([Start Navigation]) --> Current{Current Screen?}
    Current -->|Known| Target{Target Screen?}
    Current -->|Unknown| Analyze[Analyze Screen<br/>with LLM Vision]
    Analyze --> AddGraph[Add to NavigationGraph]
    AddGraph --> Target
    
    Target -->|Known| Compute[Compute Shortest Path<br/>BFS Algorithm]
    Target -->|Unknown| Infer[Infer Path<br/>LLM Prompt]
    Infer --> AddGraph2[Add to NavigationGraph]
    AddGraph2 --> Compute
    
    Compute --> Execute[Execute Actions]
    Execute --> Verify{Verify Screen}
    Verify -->|Success| Update[Update Graph<br/>Record Transition]
    Verify -->|Failure| Retry{Retry?}
    
    Retry -->|Yes| Infer
    Retry -->|No| Error[Navigation Error]
    
    Update --> Arrived{At Target?}
    Arrived -->|No| Execute
    Arrived -->|Yes| Success([Navigation Complete])
    
    Error --> Fail([Navigation Failed])
```

## Issue Detection Pipeline

```mermaid
flowchart LR
    subgraph "Input"
        A[Before Screenshot]
        B[After Screenshot]
        C[User Action]
    end
    
    subgraph "Analysis"
        D[Vision Engine<br/>SSIM Diff]
        E[LLM Analyzer<br/>Prompt: Detect Issues]
    end
    
    subgraph "Classification"
        F[Category:<br/>Visual/UX/Functional]
        G[Severity:<br/>Critical/High/Medium/Low]
        H[Confidence:<br/>0.0-1.0]
    end
    
    subgraph "Output"
        I[Issue Struct]
        J[Evidence Links]
        K[LLM Suggested Fix]
    end
    
    A --> D
    B --> D
    C --> E
    D --> E
    E --> F
    E --> G
    E --> H
    F --> I
    G --> I
    H --> I
    I --> J
    I --> K
```

## Data Flow Architecture

```mermaid
graph LR
    subgraph "Documentation"
        DOC[docs/*.md]
        FEAT[FeatureMap]
    end
    
    subgraph "Session"
        CFG[.env Config]
        SESS[SessionRecorder]
        TIMELINE[Timeline Events]
    end
    
    subgraph "Execution"
        AGENT[LLM Agent]
        SCREEN[Screenshots]
        VIDEO[Video Recording]
    end
    
    subgraph "Output"
        TICKETS[tickets/*.md]
        REPORT[qa-report.md]
        COVERAGE[Coverage Report]
    end
    
    DOC -->|Parse| FEAT
    CFG -->|Load| SESS
    
    FEAT -->|Verify| AGENT
    AGENT -->|Capture| SCREEN
    AGENT -->|Record| VIDEO
    
    AGENT -->|Detect Issues| TICKETS
    SESS -->|Export| TIMELINE
    
    SCREEN -->|Link| TICKETS
    VIDEO -->|Timestamp| TICKETS
    TIMELINE -->|Include| REPORT
    FEAT -->|Calculate| COVERAGE
    
    TICKETS --> REPORT
```

## Multi-Platform Support

```mermaid
graph TB
    subgraph "HelixQA Core"
        ORCH[Orchestrator]
    end
    
    subgraph "Platform Adapters"
        ANDROID_ADP[Android Adapter<br/>ADB Commands]
        WEB_ADP[Web Adapter<br/>Playwright]
        DESKTOP_ADP[Desktop Adapter<br/>X11/xdotool]
    end
    
    subgraph "Device/Simulator"
        ANDROID_DEV[Android Device<br/>or Emulator]
    end
    
    subgraph "Browser"
        WEB_BROWSER[Chromium<br/>Firefox<br/>WebKit]
    end
    
    subgraph "Desktop Environment"
        DESKTOP_ENV[X11 Display<br/>Window Manager]
    end
    
    ORCH -->|Platform: android| ANDROID_ADP
    ORCH -->|Platform: web| WEB_ADP
    ORCH -->|Platform: desktop| DESKTOP_ADP
    
    ANDROID_ADP -->|adb shell| ANDROID_DEV
    WEB_ADP -->|Playwright API| WEB_BROWSER
    DESKTOP_ADP -->|xdotool| DESKTOP_ENV
```

## State Management

```mermaid
stateDiagram-v2
    [*] --> Idle
    
    Idle --> Setup: Start Session
    
    Setup --> SetupComplete: Initialize Complete
    state Setup {
        [*] --> LoadingConfig
        LoadingConfig --> SelectingLLMs
        SelectingLLMs --> SpawningAgents
        SpawningAgents --> StartingRecording
        StartingRecording --> [*]
    }
    
    SetupComplete --> DocDriven: Phase 2
    
    state DocDriven {
        [*] --> FeatureVerification
        FeatureVerification --> NavigateToFeature
        NavigateToFeature --> ExecuteTestSteps
        ExecuteTestSteps --> VerifyOutcome
        VerifyOutcome --> FeatureVerification: Next Feature
        VerifyOutcome --> [*]: All Features Done
    }
    
    DocDriven --> Curiosity: Phase 3
    
    state Curiosity {
        [*] --> ExploreUnvisited
        ExploreUnvisited --> DetectIssues
        DetectIssues --> TestEdgeCases
        TestEdgeCases --> ExploreUnvisited: Continue
        TestEdgeCases --> [*]: Budget Exhausted
    }
    
    Curiosity --> Report: Phase 4
    
    state Report {
        [*] --> AggregateResults
        AggregateResults --> GenerateTickets
        GenerateTickets --> CreateTimeline
        CreateTimeline --> WriteReports
        WriteReports --> [*]
    }
    
    Report --> Complete: Session Complete
    
    Complete --> [*]
    
    Idle --> Error: Init Failed
    Setup --> Error: Setup Failed
    DocDriven --> Error: Critical Error
    Curiosity --> Error: Timeout
    
    Error --> Cleanup
    Cleanup --> [*]
```

## API Gateway Pattern

```mermaid
graph TB
    subgraph "Client"
        CLI[HelixQA CLI]
    end
    
    subgraph "API Layer"
        CMD[Command Handler]
        VALID[Validator]
        AUTH[Auth Middleware]
    end
    
    subgraph "Service Layer"
        SESSION[Session Service]
        STRATEGY[Strategy Service]
        AGENT_SVC[Agent Service]
        TICKET_SVC[Ticket Service]
    end
    
    subgraph "Data Layer"
        DB[(Session State)]
        CACHE[(Score Cache)]
        FS[File System<br/>Evidence Storage]
    end
    
    CLI -->|HTTP/gRPC| CMD
    CMD --> VALID
    VALID --> AUTH
    
    AUTH --> SESSION
    AUTH --> STRATEGY
    AUTH --> AGENT_SVC
    AUTH --> TICKET_SVC
    
    SESSION --> DB
    STRATEGY --> CACHE
    AGENT_SVC --> DB
    TICKET_SVC --> FS
    SESSION --> FS
```

## Deployment Architecture

```mermaid
graph TB
    subgraph "Development"
        DEV[Developer Workstation]
        ENV[.env File]
        SRC[Source Code]
    end
    
    subgraph "Build Pipeline"
        BUILD[Go Build]
        TEST[Go Test]
        VET[Go Vet]
    end
    
    subgraph "Runtime"
        HELIX[HelixQA Binary]
        CONFIG[Configuration]
        AGENTS[CLI Agents<br/>OpenCode, Claude, etc.]
    end
    
    subgraph "External APIs"
        ANTHROPIC[Anthropic API]
        OPENAI[OpenAI API]
        GOOGLE[Google API]
    end
    
    subgraph "Targets"
        ANDROID[Android Emulator]
        WEB[Web Browser]
        DESKTOP[Desktop App]
    end
    
    DEV --> ENV
    DEV --> SRC
    SRC --> BUILD
    BUILD --> TEST
    TEST --> VET
    VET --> HELIX
    
    ENV --> CONFIG
    HELIX --> CONFIG
    HELIX --> AGENTS
    
    HELIX -->|API Calls| ANTHROPIC
    HELIX -->|API Calls| OPENAI
    HELIX -->|API Calls| GOOGLE
    
    HELIX -->|ADB| ANDROID
    HELIX -->|Playwright| WEB
    HELIX -->|X11| DESKTOP
```

---

## Legend

- **Blue Boxes**: Core HelixQA Components
- **Green Boxes**: External Services/LLMs
- **Yellow Boxes**: Target Applications
- **Gray Boxes**: Infrastructure/Storage
- **Dashed Lines**: Optional/Conditional Flows
- **Solid Lines**: Primary Data Flow
