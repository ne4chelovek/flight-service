# Flight Service

Flight Service is a microservice that handles flight-related operations and communicates via Kafka for distributed messaging.

## Configuration

The service is configured using environment variables:

- `DB_DSN`: PostgreSQL connection string (default: `postgres://postgres:password@localhost:5432/flight_service_db?sslmode=disable`)
- `KAFKA_BROKERS`: Comma-separated list of Kafka brokers (default: `localhost:9092`)
- `KAFKA_TOPIC`: Kafka topic name (default: `flight_requests`)
- `KAFKA_GROUP_ID`: Kafka consumer group ID (default: `flight_consumers`)
- `HTTP_PORT`: HTTP server port (default: `8080`)

## Running the Service

### Prerequisites

- Go 1.21+
- Docker and Docker Compose (for running dependencies)

### Local Development

1. Copy `.env.example` to `.env` and adjust values as needed:
   ```bash
   cp .env.example .env
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the service:
   ```bash
   go run cmd/server/cmd.go
   ```

### Using Docker Compose

The service can be run with its dependencies using the provided Docker Compose configuration:

```bash
docker-compose up
```

## Endpoints

- `GET /health`: Health check endpoint
- `GET /metrics`: Metrics endpoint
- `GET /flights`: Get flights
- `POST /bookings`: Create a booking