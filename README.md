# Profile API - HNG Internship Stage 0

A simple RESTful API built with Go that returns user profile information along with dynamic cat facts from an external API.

## Features

- **GET /me** - Returns user profile with a random cat fact
- **GET /** - Welcome endpoint
- Dynamic UTC timestamps on every request
- External API integration with timeout handling
- CORS support
- Environment variable configuration
- Proper error handling and fallback messages

## Tech Stack

- **Language**: Go
- **External API**: [Cat Facts API](https://catfact.ninja/fact)
- **Dependencies**: 
  - `github.com/joho/godotenv` - Environment variable management

## Prerequisites

- Go 1.16 or higher installed
- Internet connection (for fetching cat facts)

## Installation

1. **Clone the repository**
```bash
git clone https://github.com/kingsonseang/hng-internship-backend-stage-0-profile-api
cd hng-internship-backend-stage-0-profile-api
```

2. **Install dependencies**
```bash
go mod download
```

3. **Create `.env` file** in the project root
```bash
touch .env
```

4. **Add your configuration** to `.env`
```env
USER_EMAIL=your.email@example.com
USER_NAME=Your Full Name
USER_STACK=Go
PORT=8080
```

## Running Locally

**Start the server:**
```bash
go run main.go
```

The server will start on `http://localhost:8080` (or the PORT specified in your `.env` file).

**Test the endpoints:**
```bash
# Root endpoint
curl http://localhost:8080/

# Profile endpoint
curl http://localhost:8080/me
```

## API Documentation

### GET /

Returns a welcome message.

**Response:**
```json
{
  "status": "success",
  "message": "HNG Internship Backend Track Stage 0 Task",
  "timestamp": "2025-10-18T17:30:45.123Z"
}
```

### GET /me

Returns user profile information with a random cat fact.

**Response:**
```json
{
  "status": "success",
  "user": {
    "email": "your.email@example.com",
    "name": "Your Full Name",
    "stack": "Go"
  },
  "timestamp": "2025-10-18T17:30:45.123Z",
  "fact": "Cats sleep 70% of their lives."
}
```

**Error Response (if cat API fails):**
```json
{
  "status": "success",
  "user": {
    "email": "your.email@example.com",
    "name": "Your Full Name",
    "stack": "Go"
  },
  "timestamp": "2025-10-18T17:30:45.123Z",
  "fact": "Cat fact unavailable"
}
```

### 404 Response

For any unknown routes:
```json
{
  "status": "error",
  "message": "resource not found",
  "timestamp": "2025-10-18T17:30:45.123Z"
}
```

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| USER_EMAIL | Your email address | No | default@example.com |
| USER_NAME | Your full name | No | Unknown |
| USER_STACK | Your backend stack | No | Go |
| PORT | Server port | No | 8080 |

## Deployment

The API can be deployed to various platforms:

- **Railway**: Connect GitHub repo and set environment variables
- **Fly.io**: Use `fly launch` and configure secrets
- **DigitalOcean**: Deploy to App Platform or Droplet
- **AWS/GCP**: Deploy to EC2, App Engine, or container services

**Important**: Set all environment variables in your deployment platform's dashboard.

## Project Structure

```
.
├── main.go          # Main application file
├── go.mod           # Go module dependencies
├── go.sum           # Dependency checksums
├── .env             # Environment variables (not committed)
├── .gitignore       # Git ignore file
└── README.md        # This file
```

## Error Handling

- **External API timeout**: 5 second timeout on cat facts API
- **API failure**: Returns fallback message "Cat fact unavailable"
- **Invalid routes**: Returns 404 with error JSON
- **Invalid methods**: Returns 405 Method Not Allowed

## Development

**Run with hot reload** (optional):
```bash
# Install air for hot reload
go install github.com/air-verse/air@latest

# Run with air
air
```

**Build for production:**

**Using Docker (recommended):**

Build and run:
```bash
# Build image
docker build -t profile-api .

# Run container
docker run -p 8080:8080 \
  -e USER_EMAIL="your@email.com" \
  -e USER_NAME="Your Name" \
  -e USER_STACK="Go" \
  -e PORT=8080 \
  profile-api
```

**Or use docker-compose:**

Run:
```bash
docker-compose up -d
```

**Without Docker (direct binary):**
```bash
# Windows
go build -o api.exe main.go
./api.exe

# Linux/Mac
go build -o api main.go
./api
```

## License

MIT

## Author

kingsonseang - HNG Internship Backend Track