# Freelance Platform Database

This repository contains the database design, SQL scripts, and a Golang backend implementation for a Freelance Platform. It demonstrates advanced database concepts, including triggers, stored procedures, views, and transaction management via the Unit of Work and Repository patterns.

## Features & Requirements Fulfilled

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
│   └── app/
│       └── main.go           # Application entry point and business scenarios
├── database/
│   ├── 01_schema.sql         # DDL: Tables, Triggers, and Indexes
│   └── 02_views_sps.sql      # Views and Stored Procedures
├── internal/
│   └── store/
│       ├── models.go         # Domain entities mapped to DB views
│       ├── repository.go     # Repository pattern implementation
│       └── uow.go            # Unit of Work pattern implementation
├── go.mod
└── README.md
```

## Getting Started

### Prerequisites

- [PostgreSQL](https://www.postgresql.org/) (v12 or higher)
- [Go](https://golang.org/) (v1.20 or higher)

### 1. Database Setup

1. Create a new PostgreSQL database named `freelance`.
2. Execute the schema scripts in order:

   ```bash
   psql -U postgres -d freelance -f database/01_schema.sql
   psql -U postgres -d freelance -f database/02_views_sps.sql
   ```

### 2. Application Setup

1. Navigate to the project directory and download dependencies:

   ```bash
   go mod tidy
   ```

2. Define .env by 
   ```
   copy .env.example .env
   ```
   and indicate correct password for your db

3. Run the application to test the database transaction scenario:

   ```bash
   go run cmd/app/main.go
   ```

## Architecture Notes

The Go application exclusively interacts with the database through Stored Procedures and Views. **No raw `INSERT`, `UPDATE`, or `DELETE` statements exist in the Go code.**

Transaction management is handled by the `UnitOfWork` interface. If a multi-step business process (e.g., accepting a proposal and updating a project status) fails at any step, `defer uow.Rollback()` ensures complete data consistency.

## Author

**Mykola Shevchenko**

*Taras Shevchenko National University of Kyiv*
