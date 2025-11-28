package main

import (
	"context"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// BuildSchema builds the complete GraphQL schema
func BuildSchema(httpClient *HTTPClient, config *Config) *graphql.Schema {
    timestampType := graphql.NewScalar(graphql.ScalarConfig{
        Name:        "Timestamp",
        Description: "RFC3339 timestamp",
        ParseValue: func(value interface{}) interface{} {
            return value
        },
        ParseLiteral: func(valueAST ast.Value) interface{} {
            return valueAST
        },
        Serialize: func(value interface{}) interface{} {
            return value
        },
    })

    // User type
    userType := graphql.NewObject(graphql.ObjectConfig{
        Name: "User",
        Fields: graphql.Fields{
            "id": &graphql.Field{
                Type: graphql.NewNonNull(graphql.String),
            },
            "email": &graphql.Field{
                Type: graphql.NewNonNull(graphql.String),
            },
            "username": &graphql.Field{
                Type: graphql.NewNonNull(graphql.String),
            },
            "created_at": &graphql.Field{
                Type: timestampType,
            },
        },
    })

    // Category type
    categoryType := graphql.NewObject(graphql.ObjectConfig{
        Name: "Category",
        Fields: graphql.Fields{
            "id": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Int),
            },
            "name": &graphql.Field{
                Type: graphql.NewNonNull(graphql.String),
            },
            "description": &graphql.Field{
                Type: graphql.String,
            },
        },
    })

    // Product type
    productType := graphql.NewObject(graphql.ObjectConfig{
        Name: "Product",
        Fields: graphql.Fields{
            "id": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Int),
            },
            "name": &graphql.Field{
                Type: graphql.NewNonNull(graphql.String),
            },
            "description": &graphql.Field{
                Type: graphql.String,
            },
            "price": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Float),
            },
            "sku": &graphql.Field{
                Type: graphql.String,
            },
            "stock_quantity": &graphql.Field{
                Type: graphql.Int,
            },
            "category_id": &graphql.Field{
                Type: graphql.Int,
            },
            "image_url": &graphql.Field{
                Type: graphql.String,
            },
            "created_at": &graphql.Field{
                Type: timestampType,
            },
        },
    })

    // CartItem type
    cartItemType := graphql.NewObject(graphql.ObjectConfig{
        Name: "CartItem",
        Fields: graphql.Fields{
            "id": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Int),
            },
            "product_id": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Int),
            },
            "quantity": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Int),
            },
            "price": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Float),
            },
        },
    })

    // Cart type
    cartType := graphql.NewObject(graphql.ObjectConfig{
        Name: "Cart",
        Fields: graphql.Fields{
            "id": &graphql.Field{
                Type: graphql.NewNonNull(graphql.String),
            },
            "items": &graphql.Field{
                Type: graphql.NewList(cartItemType),
            },
            "total": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Float),
            },
            "status": &graphql.Field{
                Type: graphql.String,
            },
        },
    })

    // OrderItem type
    orderItemType := graphql.NewObject(graphql.ObjectConfig{
        Name: "OrderItem",
        Fields: graphql.Fields{
            "id": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Int),
            },
            "product_id": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Int),
            },
            "quantity": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Int),
            },
            "price": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Float),
            },
        },
    })

    // Order type
    orderType := graphql.NewObject(graphql.ObjectConfig{
        Name: "Order",
        Fields: graphql.Fields{
            "id": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Int),
            },
            "items": &graphql.Field{
                Type: graphql.NewList(orderItemType),
            },
            "total": &graphql.Field{
                Type: graphql.NewNonNull(graphql.Float),
            },
            "status": &graphql.Field{
                Type: graphql.NewNonNull(graphql.String),
            },
            "created_at": &graphql.Field{
                Type: timestampType,
            },
        },
    })

    // Auth response type
    authResponseType := graphql.NewObject(graphql.ObjectConfig{
        Name: "AuthResponse",
        Fields: graphql.Fields{
            "user": &graphql.Field{
                Type: graphql.NewNonNull(userType),
            },
            "token": &graphql.Field{
                Type: graphql.NewNonNull(graphql.String),
            },
        },
    })

    // Query root
    queryType := graphql.NewObject(graphql.ObjectConfig{
        Name: "Query",
        Fields: graphql.Fields{
            "me": &graphql.Field{
                Type: userType,
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return nil, nil
                },
            },
            "products": &graphql.Field{
                Type: graphql.NewList(productType),
                Args: graphql.FieldConfigArgument{
                    "category_id": &graphql.ArgumentConfig{
                        Type: graphql.Int,
                    },
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return nil, nil
                },
            },
            "product": &graphql.Field{
                Type: productType,
                Args: graphql.FieldConfigArgument{
                    "id": &graphql.ArgumentConfig{
                        Type: graphql.NewNonNull(graphql.Int),
                    },
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return nil, nil
                },
            },
            "categories": &graphql.Field{
                Type: graphql.NewList(categoryType),
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return nil, nil
                },
            },
            "cart": &graphql.Field{
                Type: cartType,
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return nil, nil
                },
            },
            "orders": &graphql.Field{
                Type: graphql.NewList(orderType),
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return nil, nil
                },
            },
            "order": &graphql.Field{
                Type: orderType,
                Args: graphql.FieldConfigArgument{
                    "id": &graphql.ArgumentConfig{
                        Type: graphql.NewNonNull(graphql.Int),
                    },
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return nil, nil
                },
            },
        },
    })

    // Mutation root
    mutationType := graphql.NewObject(graphql.ObjectConfig{
        Name: "Mutation",
        Fields: graphql.Fields{
            "register": &graphql.Field{
                Type: authResponseType,
                Args: graphql.FieldConfigArgument{
                    "email": &graphql.ArgumentConfig{
                        Type: graphql.NewNonNull(graphql.String),
                    },
                    "username": &graphql.ArgumentConfig{
                        Type: graphql.NewNonNull(graphql.String),
                    },
                    "password": &graphql.ArgumentConfig{
                        Type: graphql.NewNonNull(graphql.String),
                    },
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return nil, nil
                },
            },
            "login": &graphql.Field{
                Type: authResponseType,
                Args: graphql.FieldConfigArgument{
                    "email": &graphql.ArgumentConfig{
                        Type: graphql.NewNonNull(graphql.String),
                    },
                    "password": &graphql.ArgumentConfig{
                        Type: graphql.NewNonNull(graphql.String),
                    },
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return nil, nil
                },
            },
            "addToCart": &graphql.Field{
                Type: cartType,
                Args: graphql.FieldConfigArgument{
                    "product_id": &graphql.ArgumentConfig{
                        Type: graphql.NewNonNull(graphql.Int),
                    },
                    "quantity": &graphql.ArgumentConfig{
                        Type: graphql.NewNonNull(graphql.Int),
                    },
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return nil, nil
                },
            },
            "removeFromCart": &graphql.Field{
                Type: cartType,
                Args: graphql.FieldConfigArgument{
                    "product_id": &graphql.ArgumentConfig{
                        Type: graphql.NewNonNull(graphql.Int),
                    },
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return nil, nil
                },
            },
            "checkout": &graphql.Field{
                Type: orderType,
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return nil, nil
                },
            },
            "cancelOrder": &graphql.Field{
                Type: orderType,
                Args: graphql.FieldConfigArgument{
                    "id": &graphql.ArgumentConfig{
                        Type: graphql.NewNonNull(graphql.Int),
                    },
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return nil, nil
                },
            },
        },
    })

    // Create schema
    schema, err := graphql.NewSchema(graphql.SchemaConfig{
        Query:    queryType,
        Mutation: mutationType,
    })

    if err != nil {
        fmt.Printf("❌ Failed to create schema: %v\n", err)
        return nil
    }

    return &schema
}

// GraphQLQuery represents incoming GraphQL request
type GraphQLQuery struct {
    Query         string                 `json:"query"`
    Variables     map[string]interface{} `json:"variables,omitempty"`
    OperationName string                 `json:"operationName,omitempty"`
}

// ExecuteQuery executes GraphQL query
func ExecuteQuery(query string, variables map[string]interface{}, schema *graphql.Schema, ctx context.Context) *graphql.Result {
    result := graphql.Do(graphql.Params{
        Schema:         *schema,
        RequestString:  query,
        VariableValues: variables,
		Context: ctx,
    })

    return result
}

// FormatResult formats GraphQL result for HTTP response
func FormatResult(result *graphql.Result) map[string]interface{} {
    response := map[string]interface{}{}

    if len(result.Errors) > 0 {
        errors := make([]map[string]interface{}, len(result.Errors))
        for i, err := range result.Errors {
            errors[i] = map[string]interface{}{
                "message": err.Error(),
            }
        }
        response["errors"] = errors
    }

    if result.Data != nil {
        response["data"] = result.Data
    }

    return response
}