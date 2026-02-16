# Catalogizer Multi-Toolchain Builder Image
# Provides Go, Node.js, Rust, JDK 17, and Android SDK for building all components
# Usage: Built by docker-compose.build.yml

FROM ubuntu:22.04

LABEL maintainer="Catalogizer Team"
LABEL description="Multi-toolchain builder for Catalogizer project"

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=UTC

# ============================================================
# Layer 1: System dependencies
# ============================================================
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    git \
    curl \
    wget \
    pkg-config \
    libssl-dev \
    ca-certificates \
    gnupg \
    unzip \
    zip \
    webkit2gtk-4.1-dev \
    libayatana-appindicator3-dev \
    librsvg2-dev \
    libgtk-3-dev \
    patchelf \
    xvfb \
    postgresql-client \
    redis-tools \
    jq \
    file \
    bc \
    && rm -rf /var/lib/apt/lists/*

# ============================================================
# Layer 2: Go 1.24
# ============================================================
ENV GO_VERSION=1.24.1
RUN wget --retry-connrefused --waitretry=10 --tries=5 -q \
        "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"
ENV GOPATH="/root/go"
ENV CGO_ENABLED=1

RUN go version

# ============================================================
# Layer 3: Node.js 18 LTS
# ============================================================
RUN curl --retry 5 --retry-delay 10 -fsSL https://deb.nodesource.com/setup_18.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

RUN node --version && npm --version

# ============================================================
# Layer 4: Rust (latest stable)
# ============================================================
ENV RUSTUP_HOME="/root/.rustup"
ENV CARGO_HOME="/root/.cargo"
RUN curl --retry 5 --retry-delay 10 --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable
ENV PATH="/root/.cargo/bin:${PATH}"

RUN cargo install tauri-cli

RUN rustc --version && cargo --version

# ============================================================
# Layer 5: JDK 17 + Android SDK
# ============================================================
RUN apt-get update && apt-get install -y --no-install-recommends openjdk-17-jdk && rm -rf /var/lib/apt/lists/*
ENV JAVA_HOME="/usr/lib/jvm/java-17-openjdk-amd64"

ENV ANDROID_HOME="/opt/android-sdk"
ENV ANDROID_SDK_ROOT="${ANDROID_HOME}"
ENV PATH="${ANDROID_HOME}/cmdline-tools/latest/bin:${ANDROID_HOME}/platform-tools:${PATH}"

RUN mkdir -p "${ANDROID_HOME}/cmdline-tools" \
    && cd "${ANDROID_HOME}/cmdline-tools" \
    && wget --retry-connrefused --waitretry=10 --tries=5 -q \
        "https://dl.google.com/android/repository/commandlinetools-linux-11076708_latest.zip" -O cmdline-tools.zip \
    && unzip -q cmdline-tools.zip \
    && mv cmdline-tools latest \
    && rm cmdline-tools.zip

RUN yes | sdkmanager --licenses >/dev/null 2>&1 || true
RUN sdkmanager \
    "platform-tools" \
    "build-tools;34.0.0" \
    "platforms;android-34"

# ============================================================
# Layer 6: Playwright browsers
# ============================================================
RUN npx playwright install --with-deps chromium

# ============================================================
# Working directory
# ============================================================
WORKDIR /project

ENTRYPOINT ["/project/scripts/build-test-release.sh"]
