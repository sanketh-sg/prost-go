# 3-layer architecture
```
┌─────────────────────────────────────┐
│  HTTP Handlers (user_handler.go)    │  ← API Layer
│  - Receive requests                 │
│  - Validate input                   │
│  - Return responses                 │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│  Business Logic (repository)        │  ← Business Logic Layer
│  - Process data                     │
│  - Apply rules                      │
│  - Coordinate operations            │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│  Database Access (models)           │  ← Data Access Layer
│  - Query database                   │
│  - Persist data                     │
│  - Handle transactions              │
└─────────────────────────────────────┘
```

This project uses several key design patterns:

First, LAYERED ARCHITECTURE - separating handlers, repositories, 
and databases. This makes code maintainable and testable.

Second, DEPENDENCY INJECTION - we pass dependencies to constructors 
rather than creating them internally. This decouples components and 
enables easy testing.

Third, REPOSITORY PATTERN - we abstract database operations. 
The repository handles all SQL queries, making it easy to test 
business logic without touching the database.

Fourth, MIDDLEWARE PATTERN - for cross-cutting concerns like logging, 
authentication, and CORS. Middleware runs in a chain before reaching 
the handler.

Fifth, CONTEXT PATTERN - Go's context is passed through all layers 
to enable cancellation, timeouts, and prevent resource leaks. 
If a client disconnects, we stop database queries immediately.

Sixth, FACTORY PATTERN - in main.go, we create and wire all 
dependencies centrally, making it easy to change how objects 
are constructed.

Finally, GRACEFUL SHUTDOWN - we handle OS signals properly, 
waiting for in-flight requests to complete before closing connections.

For example, when a user registers: the HTTP handler receives the 
request, calls the repository to check if the email exists and 
insert the user, the repository executes SQL with the request context, 
and if anything fails at any layer, we return a consistent error response.

This architecture allows us to scale, test easily, and maintain 
clean separation of concerns.


Idempotency Concept and how it applies to my project

In distributed systems, messages can be delivered twice or more due to network failures. Idempotency prevents duplicate processing. So, doing the same operation multiple times produces the same result as doing it once.

UNIQUE(event_id, service_name) ensures the same event from the same service can only be recorded once.

Each service has its own idempotency table, as they have seperate schema we keep them seprate.

replaceSchema is used to prevent hard coding of schema names in queries. Any changes to schema name will result in modifying the queries everywhere which is bad standards.

repository holds a connection to DB, any transaction uses this to modify the DB. QueryRowContext executes the query on 1 row. QueryContext returns multiple rows.  Scan copies data from a database result row into Go variables. It handles type conversion (database bytes → Go types like string, int, time.Time), maps columns to variables in order, and detects errors like missing rows or type mismatches. You must pass pointers to Scan (using &) so it can modify your variables. Without Scan, you'd have raw bytes from the database that can't be used in your Go code. It's essential for every database read operation. 
ExecContext only executes queries and does not return any rows.
