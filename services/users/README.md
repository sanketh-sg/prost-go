User service is a synchronous service it does not publish any events to other services or consume any. This is because user service needs instant results when login or register.

How it handles multiple requests?
✅ Gin uses goroutines - each HTTP request runs in its own goroutine
✅ Non-blocking - can handle thousands of concurrent requests
✅ Timeouts - ReadTimeout (15s), WriteTimeout (15s), IdleTimeout (60s)

Repository runs queries and manipulate DB, it acts as layer of abstraction by providing methods that can be used to talk to DB.

Handlers are the functions that handle the request to a specific URL, These functions internally use methods provided by repository  

We need to optimize the queries to execute and return results in short amount time as they consume connections from pool.
When Do You Need More Connections?

Scenario	            Connections Needed
Development (1 user)	        5-10
Testing (10 concurrent users)	50-100
Production (100+ users)	        200-500+
High traffic	                1000+


## Unit testing

Test files must end with `_test.go`, go test runner always loos for these files. All test functions must start with Test.

Go philosophy: "Tests live with the code they test"

It is best practice to keep them in the same package as the target code for testing, because some function are not private and access them from a different package would not be possible.

Why?
1. ✅ Tests can access private functions
2. ✅ Tests are close to implementation
3. ✅ Easy to refactor together
4. ✅ Clear ownership
5. ✅ Standard across Go community

Arrange, Act, Assert pattern
 * Arrange: Setup test data and mocks
 * Act: Call the function being tested
 * Assert: Verify the results

setup_test.go is the foundation contains reusable test utilities and mocks.

A mock is a fake object that pretends to be a real object but under your control. With mocks test results are instant and no need to fetch data from DB each time.
No dependency on DB conn at all tests are independent.

```
Without mock:
Test → Handler → Real Repository → Real Database
                 ↑ If test fails, is it handler bug or repository bug?
                 ↑ Hard to know where error came from

With mock:
Test → Handler → Fake Repository (mock)
       ↑ Test only checks handler logic
       ↑ Repository logic is tested separately
```
### Anatomy of a mock
```go
// 1. Create mock
mockRepo := new(MockUserRepository)
//          ↑ This is a fake object

// 2. Set up expectations
mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
    return u.Email == "alice@example.com"
})).Return(nil)
// ↑ Mock.On() = "When this method is called..."
// ↑ mock.Anything = "...with any first argument..."
// ↑ mock.MatchedBy() = "...and second argument matches this condition..."
// ↑ .Return(nil) = "...then return nil"

// 3. Pass to handler
handler := NewUserHandler(mockRepo, "secret")
//                        ↑ Handler will use this mock

// 4. Call handler
handler.Register(request)
// Inside handler, when it calls mockRepo.CreateUser(...)
// The mock intercepts it and returns the pre-set value

// 5. Verify mock was called
mockRepo.AssertExpectations(t)
// "Was CreateUser called as expected?"
```

Should I test this?? If it has business logic then always yes.
```
Is this code...?
│
├─ A trivial getter/setter?
│  └─ NO → Don't test
│
├─ Just forwarding to a library?
│  └─ NO → Don't test (library is tested)
│
├─ Business logic?
│  └─ YES → TEST IT
│
├─ An edge case that could break?
│  └─ YES → TEST IT
│
├─ Error handling?
│  └─ YES → TEST IT
│
├─ Integration with external system?
│  └─ YES → TEST IT
│
└─ Something users will notice if broken?
   └─ YES → TEST IT
```

## Testing handlers

While the testing of other functions were pretty staright forward, handler testing is bit of a challenge why??

* here we test http endpoints it depends in repository which is DB. We dont want to test on DB slow and requires additional setup to conn etc.
* we have do this in isolation using mocks and interfaces.

First establish an interface for the repository that it depends, so when we are testing it accepts the mock repo but it prod it accepts the real repo. An interface defines a contract: "anything with these methods is a valid repository." 

The handler now doesn't care if it gets a real database repository or a mock. It just needs something that implements the interface.
Using mock you can create fake data, database.

ARRANGE: We create a fake user and mock repo that returns it
ACT: We call handler.functions with test data
ASSERT: We check the response

For repository tests, I setup a test database connection once, run migrations to create schema, then after each test I clean up by deleting test data or rolling back transactions. This ensures test isolation.