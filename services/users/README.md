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

