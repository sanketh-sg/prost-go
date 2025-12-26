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

    // inventory - Get product inventory status
    if inventoryField, ok := queryFields["inventory"]; ok {
        inventoryField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            productID := p.Args["product_id"].(int)

            inventory, err := ctx.ProductService.GetInventory(p.Context, int64(productID))
            if err != nil {
                log.Printf("❌ Error fetching inventory: %v", err)
                return nil, err
            }

            return inventory, nil
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

    // createProduct - Create a new product (admin only)
    if createProductField, ok := mutationFields["createProduct"]; ok {
        createProductField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            // Verify authentication (admin operation)
            user, err := GetUserFromContext(p.Context)
            if err != nil {
                return nil, fmt.Errorf("❌ unauthenticated - admin operation")
            }
            log.Printf("✓ Admin user %s creating product", user["email"])

            // Extract arguments
            name := p.Args["name"].(string)
            price := p.Args["price"].(float64)
            
            var description, sku *string
            var stockQuantity, categoryID *int
            
            if desc, ok := p.Args["description"]; ok {
                if d, ok := desc.(string); ok && d != "" {
                    description = &d
                }
            }
            if s, ok := p.Args["sku"]; ok {
                if sk, ok := s.(string); ok && sk != "" {
                    sku = &sk
                }
            }
            if sq, ok := p.Args["stock_quantity"]; ok {
                if st, ok := sq.(int); ok {
                    stockQuantity = &st
                }
            }
            if cid, ok := p.Args["category_id"]; ok {
                if ci, ok := cid.(int); ok {
                    categoryID = &ci
                }
            }

            product, err := ctx.ProductService.CreateProduct(
                p.Context,
                name,
                *description,
                price,
                *sku,
                stockQuantity,
                categoryID,
            )
            if err != nil {
                log.Printf("❌ Error creating product: %v", err)
                return nil, err
            }

            log.Printf("✓ Product created: %s", name)
            return product, nil
        }
    }

    // updateProduct - Update an existing product (admin only)
    if updateProductField, ok := mutationFields["updateProduct"]; ok {
        updateProductField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            // Verify authentication (admin operation)
            user, err := GetUserFromContext(p.Context)
            if err != nil {
                return nil, fmt.Errorf("❌ unauthenticated - admin operation")
            }
            log.Printf("✓ Admin user %s updating product", user["email"])

            // Extract arguments
            id := p.Args["id"].(int)
            
            var name, description *string
            var price *float64
            var stockQuantity, categoryID *int
            
            if n, ok := p.Args["name"]; ok {
                if nm, ok := n.(string); ok && nm != "" {
                    name = &nm
                }
            }
            if d, ok := p.Args["description"]; ok {
                if desc, ok := d.(string); ok && desc != "" {
                    description = &desc
                }
            }
            if pr, ok := p.Args["price"]; ok {
                if prc, ok := pr.(float64); ok && prc > 0 {
                    price = &prc
                }
            }
            if sq, ok := p.Args["stock_quantity"]; ok {
                if st, ok := sq.(int); ok {
                    stockQuantity = &st
                }
            }
            if cid, ok := p.Args["category_id"]; ok {
                if ci, ok := cid.(int); ok {
                    categoryID = &ci
                }
            }

            product, err := ctx.ProductService.UpdateProduct(
                p.Context,
                int64(id),
                name,
                description,
                price,
                stockQuantity,
                categoryID,
            )
            if err != nil {
                log.Printf("❌ Error updating product: %v", err)
                return nil, err
            }

            log.Printf("✓ Product %d updated", id)
            return product, nil
        }
    }

    // deleteProduct - Delete a product (admin only)
    if deleteProductField, ok := mutationFields["deleteProduct"]; ok {
        deleteProductField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            // Verify authentication (admin operation)
            user, err := GetUserFromContext(p.Context)
            if err != nil {
                return nil, fmt.Errorf("❌ unauthenticated - admin operation")
            }
            log.Printf("✓ Admin user %s deleting product", user["email"])

            id := p.Args["id"].(int)

            message, err := ctx.ProductService.DeleteProduct(p.Context, int64(id))
            if err != nil {
                log.Printf("❌ Error deleting product: %v", err)
                return nil, err
            }

            log.Printf("✓ Product %d deleted", id)
            return message, nil
        }
    }

    // createCategory - Create a new category (admin only)
    if createCategoryField, ok := mutationFields["createCategory"]; ok {
        createCategoryField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            // Verify authentication (admin operation)
            user, err := GetUserFromContext(p.Context)
            if err != nil {
                return nil, fmt.Errorf("❌ unauthenticated - admin operation")
            }
            log.Printf("✓ Admin user %s creating category", user["email"])

            name := p.Args["name"].(string)
            var description string
            if desc, ok := p.Args["description"]; ok {
                if d, ok := desc.(string); ok {
                    description = d
                }
            }

            category, err := ctx.ProductService.CreateCategory(p.Context, name, description)
            if err != nil {
                log.Printf("❌ Error creating category: %v", err)
                return nil, err
            }

            log.Printf("✓ Category created: %s", name)
            return category, nil
        }
    }

    //reserveInventory - Reserve product inventory
    if reserveField, ok := mutationFields["reserveInventory"]; ok {
        reserveField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            productId := p.Args["product_id"].(int)
            quantity := p.Args["quantity"].(int)

            result, err := ctx.ProductService.ReserveInventory(p.Context,int64(productId),quantity)
            if err != nil {
                log.Printf("Error reserving inventory: %v", err)
            }
            log.Printf("Reserved %d units of product %d", quantity, productId)
            return result, nil
        }
    }

    // releaseInventory - Release reserved inventory
    if releaseField, ok := mutationFields["releaseInventory"]; ok {
        releaseField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
            productID := p.Args["product_id"].(int)
            quantity := p.Args["quantity"].(int)

            result, err := ctx.ProductService.ReleaseInventory(p.Context, int64(productID), quantity)
            if err != nil {
                log.Printf("❌ Error releasing inventory: %v", err)
                return nil, err
            }

            log.Printf("✓ Released %d units of product %d", quantity, productID)
            return result, nil
        }
    }
    
    log.Println("✓ Resolvers attached to schema")
}