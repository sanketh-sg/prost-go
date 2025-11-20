
# Building an API Gateway 


Frontend (Vue)  →  API Gateway (Node.js + GraphQL)  →  Microservices (Go)

Libraries used 

@apollo/server → Apollo GraphQL server

graphql → GraphQL core

express → HTTP server

node-fetch → To call microservices

The Gateway will:

Expose one single GraphQL endpoint /graphql

Make HTTP calls to your microservices

Aggregate responses

Do auth, validation,