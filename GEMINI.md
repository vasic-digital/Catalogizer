# GEMINI.md - Catalogizer

## Project Overview

Catalogizer is a comprehensive media collection management system that automatically detects, categorizes, and organizes your media files across multiple storage protocols including SMB, FTP, NFS, WebDAV, and local filesystem. It provides real-time monitoring, advanced analytics, and a modern web interface for managing your entire media library.

The project is composed of several components:

*   **`catalog-api`**: A Go-based REST API server that handles the core logic of the system, including media detection, metadata fetching, and real-time updates.
*   **`catalog-web`**: A React-based web application that provides a modern and responsive user interface for managing the media library.
*   **Client Applications**: The project also includes several client applications for different platforms, including Android, Android TV, and desktop.
*   **`installer-wizard`**: A graphical installation wizard for easy SMB configuration.

## Building and Running

### Backend (`catalog-api`)

1.  **Navigate to the `catalog-api` directory:**
    ```bash
    cd catalog-api
    ```
2.  **Install Go dependencies:**
    ```bash
    go mod tidy
    ```
3.  **Run the API server:**
    ```bash
    go run main.go
    ```

### Frontend (`catalog-web`)

1.  **Navigate to the `catalog-web` directory:**
    ```bash
    cd catalog-web
    ```
2.  **Install dependencies:**
    ```bash
    npm install
    ```
3.  **Start the development server:**
    ```bash
    npm run dev
    ```

### Docker

The project also includes a Docker-based deployment option.

1.  **Navigate to the `deployment` directory:**
    ```bash
    cd deployment
    ```
2.  **Start the services:**
    ```bash
    docker-compose up -d
    ```

## Development Conventions

*   **Go**: The Go code follows standard Go conventions and is formatted with `gofmt`.
*   **TypeScript/React**: The frontend code follows standard React and TypeScript conventions and is linted with ESLint and formatted with Prettier.
*   **Testing**: The project includes a comprehensive test suite for both the backend and the frontend.
    *   **Backend tests**: Run with `go test ./...` in the `catalog-api` directory.
    *   **Frontend tests**: Run with `npm test` in the `catalog-web` directory.
*   **Git**: The project uses Git for version control. Commit messages should follow the conventional commits specification.
