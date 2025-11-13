# Go Hiring Challenge

This repository contains a Go application for managing products and their prices, including functionalities for CRUD operations and seeding the database with initial data.

## Project Structure

1. **cmd/**: Contains the main application and seed command entry points.

   - `server/main.go`: The main application entry point, serves the REST API.
   - `seed/main.go`: Command to seed the database with initial product data.

2. **app/**: Contains the application logic.
3. **sql/**: Contains a very simple database migration scripts setup.
4. **models/**: Contains the data models and repositories used in the application.
5. `.env`: Environment variables file for configuration.

## Setup Code Repository

1. Create a github/bitbucket/gitlab repository and push all this code as-is.
2. Create a new branch, and provide a pull-request against the main branch with your changes. Instructions to follow.

## Application Setup

- Ensure you have Go installed on your machine.
- Ensure you have Docker installed on your machine.
- Important makefile targets:
  - `make tidy`: will install all dependencies.
  - `make docker-up`: will start the required infrastructure services via docker containers.
  - `make seed`: ⚠️ Will destroy and re-create the database tables.
  - `make test`: Will run the tests.
  - `make run`: Will start the application.
  - `make docker-down`: Will stop the docker containers.

Note: The application listens on port 8484 by default. You can change it via the `HTTP_PORT` environment variable.

Follow up for the assignemnt here: [ASSIGNMENT.md](ASSIGNMENT.md)


## API Documentation

This project includes OpenAPI/Swagger documentation for all endpoints.

- Raw OpenAPI spec: `GET /openapi.yaml`
- Interactive Swagger UI: `GET /docs`

How to use:
1. Start the app (e.g. `make run`). By default the server binds to `http://localhost:${HTTP_PORT}`.
2. Open your browser at `http://localhost:${HTTP_PORT}/docs` to explore and test the API.
3. The OpenAPI document is also available at `http://localhost:${HTTP_PORT}/openapi.yaml`.

Documented endpoints:
- `GET /catalog` — query params: `offset`, `limit`, `category`, `price_lt`. Returns `total` and `products`.
- `GET /categories` — returns a list of categories.
- `POST /categories` — creates a category. Request/response body: `{ "code": string, "name": string }`.

Error schema:
Errors follow a consistent shape:
```json
{ "error": "message", "code": "invalid" }
```
