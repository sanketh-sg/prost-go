package main

import (
    "context"
    "fmt"
    "log"

    "github.com/graphql-go/graphql"
)

// ResolverContext holds resolver dependencies
type ResolverContext struct {
    UserService    *UserService
    ProductService *ProductService
    CartService    *CartService
    OrderService   *OrderService
    TokenValidator *TokenValidator
}

// GetUserFromContext extracts user from request context
func GetUserFromContext(ctx context.Context) (map[string]interface{}, error) {
    val := ctx.Value("user")
    if val == nil {
        return nil, fmt.Errorf("unauthenticated")
    }

    claims, ok := val.(*UserClaims)
    if !ok {
        return nil, fmt.Errorf("invalid user context")
    }

    return map[string]interface{}{
        "id":       claims.UserID,
        "email":    claims.Email,
        "username": claims.Username,
    }, nil
}

// AttachResolvers attaches resolver functions to schema
func AttachResolvers(schema *graphql.Schema, ctx *ResolverContext) {
    queryFields := schema.QueryType().Fields()

    // ========== QUERY RESOLVERS ==========

    // me - Get current user profile
    if meField, ok := queryFields["me"]; ok {
        meField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            user, err := GetUserFromContext(p.Context)
            if err != nil {
                return nil, fmt.Errorf("❌ %v", err)
            }

            userID := user["id"].(string)
            profile, err := ctx.UserService.GetProfile(p.Context, userID)
            if err != nil {
                log.Printf("❌ Error fetching profile: %v", err)
                return nil, err
            }

            return profile, nil
        }
    }

    // products - List all products or filter by category
    if productsField, ok := queryFields["products"]; ok {
        productsField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            var categoryID *int64
            if val, ok := p.Args["category_id"]; ok {
                if catID, ok := val.(int); ok {
                    id := int64(catID)
                    categoryID = &id
                }
            }

            products, err := ctx.ProductService.GetProducts(p.Context, categoryID)
            if err != nil {
                log.Printf("❌ Error fetching products: %v", err)
                return nil, err
            }

            return products, nil
        }
    }

    // product - Get single product by ID
    if productField, ok := queryFields["product"]; ok {
        productField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            id := p.Args["id"].(int)
            product, err := ctx.ProductService.GetProduct(p.Context, int64(id))
            if err != nil {
                log.Printf("❌ Error fetching product: %v", err)
                return nil, err
            }

            return product, nil
        }
    }

    // categories - List all categories
    if categoriesField, ok := queryFields["categories"]; ok {
        categoriesField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            categories, err := ctx.ProductService.GetCategories(p.Context)
            if err != nil {
                log.Printf("❌ Error fetching categories: %v", err)
                return nil, err
            }

            return categories, nil
        }
    }

    // cart - Get current user's cart
    if cartField, ok := queryFields["cart"]; ok {
        cartField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            user, err := GetUserFromContext(p.Context)
            if err != nil {
                return nil, fmt.Errorf("❌ %v", err)
            }

            userID := user["id"].(string)
            cartID := userID // Simplified: use user ID as cart ID
            
            cart, err := ctx.CartService.GetCart(p.Context, cartID)
            if err != nil {
                log.Printf("❌ Error fetching cart: %v", err)
                return nil, err
            }

            return cart, nil
        }
    }

    // orders - List all user's orders
    if ordersField, ok := queryFields["orders"]; ok {
        ordersField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            user, err := GetUserFromContext(p.Context)
            if err != nil {
                return nil, fmt.Errorf("❌ %v", err)
            }

            userID := user["id"].(string)
            orders, err := ctx.OrderService.GetOrders(p.Context, userID)
            if err != nil {
                log.Printf("❌ Error fetching orders: %v", err)
                return nil, err
            }

            return orders, nil
        }
    }

    // order - Get single order by ID
    if orderField, ok := queryFields["order"]; ok {
        orderField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            id := p.Args["id"].(int)
            order, err := ctx.OrderService.GetOrder(p.Context, int64(id))
            if err != nil {
                log.Printf("❌ Error fetching order: %v", err)
                return nil, err
            }

            return order, nil
        }
    }

    // ========== MUTATION RESOLVERS ==========

    mutationFields := schema.MutationType().Fields()

    // register - Create new user account
    if registerField, ok := mutationFields["register"]; ok {
        registerField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            email := p.Args["email"].(string)
            username := p.Args["username"].(string)
            password := p.Args["password"].(string)

            authResp, err := ctx.UserService.Register(p.Context, email, username, password)
            if err != nil {
                log.Printf("❌ Registration error: %v", err)
                return nil, err
            }

            return authResp, nil
        }
    }

    // login - Authenticate user and get token
    if loginField, ok := mutationFields["login"]; ok {
        loginField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            email := p.Args["email"].(string)
            password := p.Args["password"].(string)

            authResp, err := ctx.UserService.Login(p.Context, email, password)
            if err != nil {
                log.Printf("❌ Login error: %v", err)
                return nil, err
            }

            return authResp, nil
        }
    }

    // addToCart - Add product to user's cart
    if addToCartField, ok := mutationFields["addToCart"]; ok {
        addToCartField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            user, err := GetUserFromContext(p.Context)
            if err != nil {
                return nil, fmt.Errorf("❌ %v", err)
            }

            userID := user["id"].(string)
            cartID := userID // Simplified: use user ID as cart ID

            productID := p.Args["product_id"].(int)
            quantity := p.Args["quantity"].(int)

            cart, err := ctx.CartService.AddToCart(p.Context, cartID, int64(productID), quantity)
            if err != nil {
                log.Printf("❌ Error adding to cart: %v", err)
                return nil, err
            }

            return cart, nil
        }
    }

    // removeFromCart - Remove product from user's cart
    if removeFromCartField, ok := mutationFields["removeFromCart"]; ok {
        removeFromCartField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            user, err := GetUserFromContext(p.Context)
            if err != nil {
                return nil, fmt.Errorf("❌ %v", err)
            }

            userID := user["id"].(string)
            cartID := userID // Simplified: use user ID as cart ID
            productID := p.Args["product_id"].(int)

            cart, err := ctx.CartService.RemoveFromCart(p.Context, cartID, int64(productID))
            if err != nil {
                log.Printf("❌ Error removing from cart: %v", err)
                return nil, err
            }

            return cart, nil
        }
    }

    // checkout - Convert cart to order (triggers saga)
    if checkoutField, ok := mutationFields["checkout"]; ok {
        checkoutField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            user, err := GetUserFromContext(p.Context)
            if err != nil {
                return nil, fmt.Errorf("❌ %v", err)
            }

            userID := user["id"].(string)
            cartID := userID // Simplified: use user ID as cart ID

            // Call checkout which initiates saga and returns order
            result, err := ctx.CartService.Checkout(p.Context, cartID)
            if err != nil {
                log.Printf("❌ Checkout error: %v", err)
                return nil, err
            }

            return result, nil
        }
    }

    // cancelOrder - Cancel an existing order
    if cancelOrderField, ok := mutationFields["cancelOrder"]; ok {
        cancelOrderField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            id := p.Args["id"].(int)

            order, err := ctx.OrderService.CancelOrder(p.Context, int64(id))
            if err != nil {
                log.Printf("❌ Error cancelling order: %v", err)
                return nil, err
            }

            return order, nil
        }
    }

    log.Println("✓ Resolvers attached to schema")
}