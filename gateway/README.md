The gateway is a GraphQL API that aggregates all microservices. Instead of calling service directly at /endpoints they serve, you'll call the gateway's GraphQL endpoint.

In Go, a "package" can span MULTIPLE files.
When you run `go run .`, Go compiles ALL files in the package together.
This is different from importing external packages - it's compiling local files as one unit.

We can use web IDE to test graphql at `https://studio.apollographql.com/dev`

POST /graphql - for mutations (register, login, etc.)
GET /graphql - for introspection