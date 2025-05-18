# 🚀 Dispatcher Service

A service that reconstructs a valid flight itinerary from a list of airline tickets.

## 📋 Service Description

The Dispatcher service is a RESTful API that takes a list of flight tickets as input and reconstructs a valid itinerary that visits all destinations exactly once. It uses a modified version of Hierholzer's algorithm to find a valid path.

The service handles various edge cases and validates the input to ensure that the itinerary is valid:
- No duplicate tickets (same source and destination)
- No cycles in the itinerary
- Valid starting and ending points

## 📦 Prerequisites

To run this service locally, you need:

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

No Go installation is required as the service runs in a Docker container.

## 🏃 Running the Service Locally

### Using Docker Compose

1. Clone the repository:

```bash
git clone https://github.com/dsha256/dispatcher.git
cd dispatcher
```

2. Start the service using Docker Compose:

```bash
docker compose up --build
```

Alternatively, if you have [Task](https://taskfile.dev/) installed:

```bash
task compose-up
```

The service will be available at http://localhost:3000.

3. To stop the service:

```bash
docker compose down --remove-orphans --volumes
```

Or with Task:

```bash
task compose-down
```

## 🔌 API Endpoints

### Reconstruct Itinerary

Reconstructs a valid flight itinerary from a list of airline tickets.

- **URL**: `/api/v1/dispatcher/itinerary`
- **Method**: `POST`
- **Content-Type**: `application/json`

#### Request Body

```json
{
  "tickets": [
    ["LAX", "DXB"],
    ["JFK", "LAX"],
    ["SFO", "SJC"],
    ["DXB", "SFO"]
  ]
}
```

#### Success Response

- **Code**: 200 OK
- **Content**:

```json
{
  "status": "success",
  "message": "",
  "data": {
    "linear_path": ["JFK", "LAX", "DXB", "SFO", "SJC"]
  }
}
```

#### Error Response

- **Code**: 400 Bad Request
- **Content**:

```json
{
  "status": "error",
  "message": "multiple same destination",
  "err": "multiple same destination"
}
```

### Health Checks

The service provides two health check endpoints:

- **Liveness**: `/api/v1/liveness` - Checks if the service is running
- **Readiness**: `/api/v1/readiness` - Checks if the service is ready to process requests

## 🔍 Example Requests Using curl

### Reconstruct Itinerary

```bash
curl -X POST http://localhost:3000/api/v1/dispatcher/itinerary \
  -H "Content-Type: application/json" \
  -d '{
    "tickets": [
      ["LAX", "DXB"],
      ["JFK", "LAX"],
      ["SFO", "SJC"],
      ["DXB", "SFO"]
    ]
  }'
```

### Health Checks

```bash
# Liveness check
curl http://localhost:3000/api/v1/liveness

# Readiness check
curl http://localhost:3000/api/v1/readiness
```

## 🧪 Running Tests

### Using Docker

You can run tests inside a Docker container:

```bash
docker run --rm -v $(pwd):/app -w /app golang:1.24-alpine go test -v -race ./...
```

### Using Task

If you have [Task](https://taskfile.dev/) installed:

```bash
task test
```

This will run all tests in verbose mode with race detection enabled.

### Running Specific Tests

To run specific tests:

```bash
# Run unit tests for the dispatcher package
go test -v -race ./internal/dispatcher

# Run integration tests for the handler package
go test -v -race ./internal/handler
```

## 👨‍💻 Development

The service is built with Go 1.24 and uses the following components:

- Standard library HTTP server
- Custom middleware for logging and error recovery
- JSON for request/response serialization
