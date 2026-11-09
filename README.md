# Freelance Platform Database

This repository contains the database design, SQL scripts, and a Golang backend implementation for a Freelance Platform. It demonstrates advanced relational database concepts (Lab 1) and Polyglot Persistence architecture utilizing NoSQL databases for performance optimization (Lab 2).

## Lab 2: NoSQL Integration & Benchmarking (Polyglot Persistence)

In the second phase of this project, we introduced NoSQL databases to handle specific domain requirements more efficiently, proving the effectiveness of using the right tool for the job.

1. **Document Database (MongoDB):**
   - **Why:** Freelancer profiles are read-heavy and require joining multiple SQL tables (`users`, `freelancer_profiles`, `skills`, `reviews`).
   - **Implementation:** Denormalized the profile data into a single MongoDB document (`FreelancerProfileDoc`), allowing O(1) read complexity.
   - **Benchmark Results:** Tested 5,000 concurrent reads against 10,000 seeded records. MongoDB outperformed PostgreSQL by **~6x** (average latency of ~175µs vs ~1ms) due to the elimination of expensive SQL `JOIN` operations.

2. **Key-Value Store (Redis):**
   - **Why:** Caching frequently accessed, computationally expensive data like "Top Active Projects".
   - **Implementation:** Implemented JSON serialization and TTL (Time-To-Live) caching using Redis.

## Lab 1: Features & Relational Requirements Fulfilled

1. **Logical Data Model:** Complete ER diagram outlining relationships, foreign keys, and constraints.
2. **Normalized Schema:** 16 distinct entities covering user profiles, project management, financial transactions, and messaging.
3. **Audit & Data Integrity:**
   - `Soft delete` mechanism implemented for critical entities (Users, Projects, Messages) to maintain historical data.
   - Audit trails (`updated_at`, `updated_by`) managed automatically via PL/pgSQL triggers.
4. **Relational DBMS:** Fully implemented and optimized for **PostgreSQL**.
5. **Database Objects (10+):**
   - **Views:** Aggregated data slices (e.g., active projects, freelancer statistics).
   - **Stored Procedures:** Encapsulated business logic (e.g., accepting proposals, sending messages) preventing direct data manipulation from the application layer.
   - **Triggers:** Automated timestamp updates.
6. **Indexing Strategies:** Utilizes both **B-Tree** (for foreign keys) and **GIN** indexes (for full-text search on project descriptions).
7. **Application Architecture:** Data access layer written in **Go** using the **Repository** and **Unit of Work** patterns to ensure atomic transactions and clean separation of concerns.

## Database Schema (ER Diagram)

![ER Diagram](docs/er_diagram.png)

*(Note: Ensure the diagram image is placed in the `docs` folder or update this path.)*

## Project Structure

```text
.
├── cmd/
│   ├── app/
│   │   └── main.go           # Application entry point and business scenarios (Lab 1)
│   └── benchmark/
│       └── main.go           # 10k seeder and concurrency benchmark for PG vs Mongo (Lab 2)
├── database/
│   ├── 01_schema.sql         # DDL: Tables, Triggers, and Indexes
│   └── 02_views_sps.sql      # Views and Stored Procedures
├── internal/
│   ├── nosql/
│   │   ├── mongodb.go        # MongoDB Document storage implementation
│   │   └── redis.go          # Redis Key-Value caching implementation
│   └── store/
│       ├── models.go         # Domain entities mapped to DB views
│       ├── repository.go     # Repository pattern implementation
│       └── uow.go            # Unit of Work pattern implementation
├── docker-compose.yml        # Docker infrastructure for MongoDB and Redis
├── go.mod
└── README.md
```

## Getting Started

### Prerequisites

- [PostgreSQL](https://www.postgresql.org/) (v12 or higher)
- [Go](https://golang.org/) (v1.20 or higher)
- [Docker & Docker Compose](https://www.docker.com/) (For NoSQL containers)

### 1. Database Setup (PostgreSQL)

1. Create a new PostgreSQL database named `freelance`.
2. Execute the schema scripts in order:

   ```bash
   psql -U postgres -d freelance -f database/01_schema.sql
   psql -U postgres -d freelance -f database/02_views_sps.sql
   ```

### 2. Infrastructure Setup (NoSQL)

Start the MongoDB and Redis containers in the background:

```bash
docker-compose up -d
```

### 3. Application Setup

1. Navigate to the project directory and download dependencies:

   ```bash
   go mod tidy
   ```

2. Define environment variables by copying the example file:

   ```bash
   copy .env.example .env
   ```

   *(Update the `.env` file with the correct PostgreSQL password for your local DB).*

3. **Run Lab 1 (Transactions):** Test the PostgreSQL Repository and Unit of Work patterns:

   ```bash
   go run cmd/app/main.go
   ```

4. **Run Lab 2 (Benchmark):** Seed 10,000 records and run the concurrency benchmark (PostgreSQL vs MongoDB) and test Redis caching:

   ```bash
   go run cmd/benchmark/main.go
   ```

## Architecture Notes

- **SQL Operations:** The Go application exclusively interacts with the PostgreSQL database through Stored Procedures and Views. **No raw `INSERT`, `UPDATE`, or `DELETE` statements exist in the Go code.**
- **Transactions:** Transaction management is handled by the `UnitOfWork` interface. If a multi-step business process (e.g., accepting a proposal and updating a project status) fails at any step, `defer uow.Rollback()` ensures complete data consistency.
- **Polyglot Persistence:** To bypass heavy SQL joins on read-heavy routes, documents are synchronized into a NoSQL store, reducing latency drastically at the cost of slight data duplication.

## Author

**Mykola Shevchenko**

*Taras Shevchenko National University of Kyiv*