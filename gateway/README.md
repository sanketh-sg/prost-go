The gateway is a GraphQL API that aggregates all microservices. Instead of calling service directly at /endpoints they serve, you'll call the gateway's GraphQL endpoint.

In Go, a "package" can span MULTIPLE files.
When you run `go run .`, Go compiles ALL files in the package together.
This is different from importing external packages - it's compiling local files as one unit.

We can use web IDE to test graphql at `https://studio.apollographql.com/dev`
use URL `http://localhost/graphql`

POST /graphql - for mutations (register, login, etc.) data is present in request body
```
POST /graphql HTTP/1.1
Content-Type: application/json

{
  "query": "query { me { id } }",
  "variables": { ... },
  "operationName": "..."
}
```
GET /graphql - for introspection, query is present in URL

## Workflow

1️⃣  Client sends GraphQL mutation:
    POST /graphql
    Authorization: <not needed for register>
    
    mutation {
      register(email: "alice@example.com", username: "alice", password: "secret") {
        user { id, email, username }
        token
      }
    }

2️⃣  main.go receives request
    - CORS middleware allows it ✓
    - authMiddleware skips auth for register (no user yet)
    - router parses JSON into GraphQLQuery{}

3️⃣  schema.go executes query
    - ExecuteQuery() calls graphql-go library
    - graphql-go looks up "register" in Mutation fields
    - Calls registerField.Resolve() function

4️⃣  resolvers.go - registerField.Resolve executes:
    email := p.Args["email"].(string)           // "alice@example.com"
    username := p.Args["username"].(string)     // "alice"
    password := p.Args["password"].(string)     // "secret"
    
    authResp, err := ctx.UserService.Register(...) // Call layer 5

5️⃣  services.go - UserService.Register():
    reqBody := RegisterRequest{
        Email:    "alice@example.com",
        Username: "alice",
        Password: "secret",
    }
    
    respBody, err := us.httpClient.POST(        // Call layer 6
        "http://users:8083/register",           // Gateway knows service URL from config
        nil,
        reqBody,
    )
    
    json.Unmarshal(respBody, &authResp)

6️⃣  client.go - HTTPClient.POST():
    bodyBytes := json.Marshal(reqBody)          // Serialize to JSON
    req := http.NewRequestWithContext(ctx, "POST", "http://users:8083/register", ...)
    req.Header.Set("Content-Type", "application/json")
    resp := hc.client.Do(req)                   // Actually send HTTP request
    respBody := io.ReadAll(resp.Body)
    return respBody

7️⃣  Users Service (port 8083) processes:
    POST /register with { email, username, password }
    - Hash password
    - Insert into database
    - Generate JWT token
    - Return: { user: {id, email, username}, token: "eyJ..." }

8️⃣  Response bubbles back through layers:
    respBody = `{"user": {"id": "uuid1", ...}, "token": "eyJ..."}`
    
    services.go unmarshals → AuthResponse struct
    resolvers.go returns → authResp
    schema.go's FormatResult() wraps:
    {
      "data": {
        "register": {
          "user": {"id": "uuid1", "email": "alice@example.com", "username": "alice"},
          "token": "eyJhbGciOiJIUzI1NiIs..."
        }
      }
    }

9️⃣  Client (Apollo Studio) receives formatted response ✓
    Displays user info + token for subsequent authenticated queries