package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/url"
)

// ============ USER SERVICE ============

// UserService handles user-related operations
type UserService struct {
    baseURL    string
    httpClient *HTTPClient
}

// NewUserService creates a new user service client
func NewUserService(baseURL string, httpClient *HTTPClient) *UserService {
    return &UserService{
        baseURL:    baseURL,
        httpClient: httpClient,
    }
}

// RegisterRequest represents registration request
type RegisterRequest struct {
    Email    string `json:"email"`
    Username string `json:"username"`
    Password string `json:"password"`
}

// LoginRequest represents login request
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// AuthResponse represents auth response
type AuthResponse struct {
    User  map[string]interface{} `json:"user"`
    Token string                 `json:"token"`
}

// Register calls users service registration endpoint
func (us *UserService) Register(ctx context.Context, email, username, password string) (*AuthResponse, error) {
    reqBody := RegisterRequest{
        Email:    email,
        Username: username,
        Password: password,
    }

    respBody, err := us.httpClient.POST(ctx, fmt.Sprintf("%s/register", us.baseURL), nil, reqBody)
    if err != nil {
        return nil, err
    }

    var authResp AuthResponse
    if err := json.Unmarshal(respBody, &authResp); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return &authResp, nil
}

// Login calls users service login endpoint
func (us *UserService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
    reqBody := LoginRequest{
        Email:    email,
        Password: password,
    }

    respBody, err := us.httpClient.POST(ctx, fmt.Sprintf("%s/login", us.baseURL), nil, reqBody)
    if err != nil {
        return nil, err
    }

    var authResp AuthResponse
    if err := json.Unmarshal(respBody, &authResp); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return &authResp, nil
}

// GetProfile calls users service get profile endpoint
func (us *UserService) GetProfile(ctx context.Context, userID string) (map[string]interface{}, error) {
    respBody, err := us.httpClient.GET(ctx, fmt.Sprintf("%s/profile/%s", us.baseURL, url.PathEscape(userID)), nil)
    if err != nil {
        return nil, err
    }

    var profile map[string]interface{}
    if err := json.Unmarshal(respBody, &profile); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return profile, nil
}

// ============ PRODUCT SERVICE ============

// ProductService handles product-related operations
type ProductService struct {
    baseURL    string
    httpClient *HTTPClient
}

// NewProductService creates a new product service client
func NewProductService(baseURL string, httpClient *HTTPClient) *ProductService {
    return &ProductService{
        baseURL:    baseURL,
        httpClient: httpClient,
    }
}

// GetProducts calls products service list endpoint
func (ps *ProductService) GetProducts(ctx context.Context, categoryID *int64) ([]map[string]interface{}, error) {
    url := fmt.Sprintf("%s/products", ps.baseURL)
    if categoryID != nil {
        url = fmt.Sprintf("%s?category_id=%d", url, *categoryID)
    }

    respBody, err := ps.httpClient.GET(ctx, url, nil)
    if err != nil {
        return nil, err
    }

    var products []map[string]interface{}
    if err := json.Unmarshal(respBody, &products); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return products, nil
}

// GetProduct calls products service get endpoint
func (ps *ProductService) GetProduct(ctx context.Context, id int64) (map[string]interface{}, error) {
    respBody, err := ps.httpClient.GET(ctx, fmt.Sprintf("%s/products/%d", ps.baseURL, id), nil)
    if err != nil {
        return nil, err
    }

    var product map[string]interface{}
    if err := json.Unmarshal(respBody, &product); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return product, nil
}

// GetCategories calls products service categories endpoint
func (ps *ProductService) GetCategories(ctx context.Context) ([]map[string]interface{}, error) {
    respBody, err := ps.httpClient.GET(ctx, fmt.Sprintf("%s/categories", ps.baseURL), nil)
    if err != nil {
        return nil, err
    }

    var categories []map[string]interface{}
    if err := json.Unmarshal(respBody, &categories); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return categories, nil
}

// ============ CART SERVICE ============

// CartService handles cart-related operations
type CartService struct {
    baseURL    string
    httpClient *HTTPClient
}

// NewCartService creates a new cart service client
func NewCartService(baseURL string, httpClient *HTTPClient) *CartService {
    return &CartService{
        baseURL:    baseURL,
        httpClient: httpClient,
    }
}

// GetCart calls cart service get endpoint
func (cs *CartService) GetCart(ctx context.Context, cartID string) (map[string]interface{}, error) {
    respBody, err := cs.httpClient.GET(ctx, fmt.Sprintf("%s/carts/%s", cs.baseURL, url.PathEscape(cartID)), nil)
    if err != nil {
        return nil, err
    }

    var cart map[string]interface{}
    if err := json.Unmarshal(respBody, &cart); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return cart, nil
}

// AddToCart calls cart service add item endpoint
func (cs *CartService) AddToCart(ctx context.Context, cartID string, productID int64, quantity int) (map[string]interface{}, error) {
    reqBody := map[string]interface{}{
        "product_id": productID,
        "quantity":   quantity,
    }

    respBody, err := cs.httpClient.POST(ctx, fmt.Sprintf("%s/carts/%s/items", cs.baseURL, url.PathEscape(cartID)), nil, reqBody)
    if err != nil {
        return nil, err
    }

    var cart map[string]interface{}
    if err := json.Unmarshal(respBody, &cart); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return cart, nil
}

// RemoveFromCart calls cart service remove item endpoint
func (cs *CartService) RemoveFromCart(ctx context.Context, cartID string, productID int64) (map[string]interface{}, error) {
    respBody, err := cs.httpClient.DELETE(ctx, fmt.Sprintf("%s/carts/%s/items/%d", cs.baseURL, url.PathEscape(cartID), productID), nil)
    if err != nil {
        return nil, err
    }

    var cart map[string]interface{}
    if err := json.Unmarshal(respBody, &cart); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return cart, nil
}

// Checkout calls cart service checkout endpoint
func (cs *CartService) Checkout(ctx context.Context, cartID string) (map[string]interface{}, error) {
    respBody, err := cs.httpClient.POST(ctx, fmt.Sprintf("%s/carts/%s/checkout", cs.baseURL, url.PathEscape(cartID)), nil, nil)
    if err != nil {
        return nil, err
    }

    var result map[string]interface{}
    if err := json.Unmarshal(respBody, &result); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return result, nil
}

// ============ ORDER SERVICE ============

// OrderService handles order-related operations
type OrderService struct {
    baseURL    string
    httpClient *HTTPClient
}

// NewOrderService creates a new order service client
func NewOrderService(baseURL string, httpClient *HTTPClient) *OrderService {
    return &OrderService{
        baseURL:    baseURL,
        httpClient: httpClient,
    }
}

// GetOrder calls orders service get endpoint
func (os *OrderService) GetOrder(ctx context.Context, orderID int64) (map[string]interface{}, error) {
    respBody, err := os.httpClient.GET(ctx, fmt.Sprintf("%s/orders/%d", os.baseURL, orderID), nil)
    if err != nil {
        return nil, err
    }

    var order map[string]interface{}
    if err := json.Unmarshal(respBody, &order); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return order, nil
}

// GetOrders calls orders service list endpoint
func (os *OrderService) GetOrders(ctx context.Context, userID string) ([]map[string]interface{}, error) {
    respBody, err := os.httpClient.GET(ctx, fmt.Sprintf("%s/users/%s/orders", os.baseURL, url.PathEscape(userID)), nil)
    if err != nil {
        return nil, err
    }

    var orders []map[string]interface{}
    if err := json.Unmarshal(respBody, &orders); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return orders, nil
}

// CancelOrder calls orders service cancel endpoint
func (os *OrderService) CancelOrder(ctx context.Context, orderID int64) (map[string]interface{}, error) {
    respBody, err := os.httpClient.POST(ctx, fmt.Sprintf("%s/orders/%d/cancel", os.baseURL, orderID), nil, nil)
    if err != nil {
        return nil, err
    }

    var order map[string]interface{}
    if err := json.Unmarshal(respBody, &order); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return order, nil
}

// GetSagaState calls orders service get saga state endpoint
func (os *OrderService) GetSagaState(ctx context.Context, correlationID string) (map[string]interface{}, error) {
    respBody, err := os.httpClient.GET(ctx, fmt.Sprintf("%s/saga/%s", os.baseURL, url.PathEscape(correlationID)), nil)
    if err != nil {
        return nil, err
    }

    var sagaState map[string]interface{}
    if err := json.Unmarshal(respBody, &sagaState); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return sagaState, nil
}