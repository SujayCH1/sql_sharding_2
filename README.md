# SQL Sharding Project

## Project Description

The **SQL Sharding Project** is a system that automatically analyzes SQL schemas and queries to infer optimal shard keys and route queries efficiently across distributed database shards.

Instead of relying on manual shard-key selection, the system parses PostgreSQL-compatible SQL into an Abstract Syntax Tree (AST), builds relational and access-pattern graphs, and applies algorithmic analysis to determine suitable shard keys. A routing layer then ensures queries are executed on the correct shard(s) with minimal fan-out.

The project is designed to be **database-aware, scalable, and extensible**, making it suitable for distributed SQL systems and sharded database architectures.

---

## Tech Stack

- **Language:** Go (Golang)
- **Database:** PostgreSQL
- **SQL Parsing:** PostgreSQL AST parser (`pg_query_go`)
- **Routing:** Consistent Hashing
- **Architecture:**
  - Schema & AST Parser
  - Shard Key Inference Engine
  - Router Layer
  - Runtime Execution Plane
  - Migration Engine
  - Analytics & Observability Layer
- **Data Structures & Algorithms:**
  - FK graph analysis
  - Fan-out minimization
  - Join-frequency ranking
  - Static + dynamic shard-key selection
- **Tooling:**
  - Go Modules
  - Docker (optional)

---

## How It Works

### 1. Schema Parsing
- SQL schemas are parsed into a PostgreSQL AST.
- Tables, columns, primary keys, and foreign keys are extracted.
- Metadata is stored in internal system tables for further analysis.

---

### 2. Relationship Graph Construction
- A directed foreign-key graph is built across tables.
- Column fan-out and join depth are computed.
- High-connectivity columns become shard-key candidates.

---

### 3. Shard Key Inference
- **Static analysis:** PK–FK chains, cardinality hints.
- **Dynamic analysis:** Runtime query access patterns.
- Candidates are ranked based on locality preservation and fan-out reduction.

---

### 4. Routing Layer (Consistent Hashing)

- Shards are placed on a **consistent hash ring**.
- Shard keys are hashed to determine shard ownership.
- Virtual nodes are used for better load distribution.
- Supports:
  - Minimal reshuffling when adding/removing shards
  - Deterministic routing
  - Single-shard lookup for point queries

**Routing outcomes:**
- **Single-shard execution** (preferred)
- **Multi-shard fan-out** (fallback)

---

### 5. Runtime Query Execution

- Incoming SQL queries are parsed into AST form.
- The router inspects:
  - WHERE clauses
  - JOIN predicates
  - Shard key availability
- Queries are classified as:
  - Point queries
  - Scoped multi-shard queries
  - Full fan-out queries

**Execution flow:**
1. Resolve target shard(s)
2. Dispatch query in parallel (if fan-out)
3. Aggregate results
4. Return unified response

---

### 6. Schema & Data Migrations

- Schema changes are detected using AST diffs.
- Supports:
  - `ALTER TABLE` (add/remove columns)
  - Constraint changes (PK/FK)
- Migration engine ensures:
  - Forward-only migrations
  - Shard-safe execution
  - Compatibility checks before rollout

Planned support for:
- Online schema migrations
- Background data re-sharding

---

### 7. Logging

- Structured logging across all layers:
  - Router decisions
  - Shard selection
  - Query execution latency
  - Fan-out events
- Log levels:
  - DEBUG
  - INFO
  - WARN
  - ERROR
- Designed for easy integration with log aggregators.

---

### 8. Analytics & Observability

- Runtime metrics collected for:
  - Query frequency
  - Shard hit distribution
  - Fan-out ratio
  - Latency percentiles
- Used by:
  - Shard key re-evaluation
  - Hot-shard detection
  - Capacity planning
- Analytics feed back into dynamic shard-key ranking.

---

## Installation Guide

### Prerequisites

- Go **1.21+**
- PostgreSQL **14+**
- Git
- (Optional) Docker & Docker Compose
- Node **24.12.0**
- Typescript **~5.9.3**
- React **19.2.3**

---

### Clone the Repository

```bash
git clone https://github.com/SujayCH1/sql_sharding_2.git
