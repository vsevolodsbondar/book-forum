# Auth Service Integration

## Overview

The backend communicates with the separate Auth Service through an HTTP client.

The integration follows these principles:

* The backend does not make HTTP requests directly from handlers.
* All communication with the Auth Service is encapsulated in `AuthHTTPClient`.
* `AuthInterface` defines the contract used by handlers and other application components.
* DTOs represent the request and response formats of the Auth Service API.
* A shared `doRequest` method handles common HTTP request logic.
* The Auth client is injected into handlers through dependency injection.
* `context.Context` is propagated from the incoming HTTP request to the Auth Service request.

The general flow is:

```text
HTTP Handler
     |
     v
AuthInterface
     |
     v
AuthHTTPClient
     |
     v
HTTP request
     |
     v
Auth Service
```

This separation makes the backend easier to test because handlers depend on an interface rather than directly on the HTTP implementation.

---

# 1. Auth Client Structure

The Auth client is responsible for communication with the Auth Service.

```go
type AuthHTTPClient struct {
    BaseURL string
    Client  *http.Client
}
```

### `BaseURL`

Contains the base URL of the Auth Service.

For Docker Compose, for example:

```text
http://auth:8081
```

The final endpoint is constructed by combining `BaseURL` with the endpoint path:

```go
c.BaseURL + "/v1/login"
```

### `Client`

An `*http.Client` is injected into the Auth client.

This allows configuration of properties such as request timeout:

```go
&http.Client{
    Timeout: 3 * time.Second,
}
```

The timeout is important for service-to-service communication so that the backend does not wait indefinitely if the Auth Service becomes unavailable.

---

# 2. Auth Interface

The rest of the application should depend on an interface rather than directly on `AuthHTTPClient`.

```go
type AuthInterface interface {
    RegisterUser(context.Context, RegisterUserRequestDTO) (RegisterUserResponseDTO, error)
    LoginUser(context.Context, LoginUserRequestDTO) (LoginUserResponseDTO, error)
    ValidateSession(context.Context, string) (ValidateSessionResponseDTO, error)
    LogoutUser(context.Context, string) error
}
```

The interface describes the operations supported by the Auth Service.

For example:

```go
response, err := h.auth.ValidateSession(ctx, sessionToken)
```

The handler does not need to know:

* how the HTTP request is created;
* what URL is used;
* how JSON is encoded;
* how the response is decoded;
* how headers are configured.

It only knows that an authentication operation is available.

---

# 3. DTOs

DTOs describe the API contract between the backend and the Auth Service.

For example, registration:

```go
type RegisterUserRequestDTO struct {
    Email    string `json:"email"`
    Username string `json:"username"`
    Password string `json:"password"`
}

type RegisterUserResponseDTO struct {
    User UserCreatedDTO `json:"user"`
}
```

Login:

```go
type LoginUserRequestDTO struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginUserResponseDTO struct {
    User    UserDTO    `json:"user"`
    Session SessionDTO `json:"session"`
}
```

Session information:

```go
type SessionDTO struct {
    ID        string `json:"id"`
    Token     string `json:"token"`
    ExpiresAt int64  `json:"expires_at"`
}
```

Session validation:

```go
type ValidateSessionResponseDTO struct {
    Session SessionDTO `json:"session"`
    User    UserDTO    `json:"user"`
}
```

Error responses from the Auth Service are represented separately:

```go
type ResponseError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

The JSON tags must match the Auth Service API contract.

---

# 4. Implementing an Auth Operation

Each public method of `AuthHTTPClient` should contain only the information specific to the particular endpoint.

For example:

```go
func (c *AuthHTTPClient) LoginUser(
    ctx context.Context,
    dto LoginUserRequestDTO,
) (LoginUserResponseDTO, error) {

    var result LoginUserResponseDTO

    params := AuthClientRequestParams{
        Method:         http.MethodPost,
        Path:           "/v1/login",
        RequestBody:    dto,
        ExpectedStatus: http.StatusOK,
        RequestResult:  &result,
    }

    err := c.doRequest(ctx, params)
    if err != nil {
        return LoginUserResponseDTO{}, err
    }

    return result, nil
}
```

The endpoint-specific method specifies:

1. HTTP method;
2. API path;
3. request body;
4. expected HTTP status;
5. destination for the response.

All common HTTP logic is handled by `doRequest`.

---

# 5. Shared `doRequest` Method

The purpose of `doRequest` is to avoid duplicating the same HTTP logic in every Auth client method.

```go
type AuthClientRequestParams struct {
    Method         string
    Path           string
    RequestBody    any
    ExpectedStatus int
    RequestResult  any
    RequestHeaders map[string]string
}
```

The method performs the following steps:

```text
Request DTO
    |
    v
json.Marshal
    |
    v
http.NewRequestWithContext
    |
    v
Set headers
    |
    v
http.Client.Do
    |
    v
Check HTTP status
    |
    +---- error status ---> decode Auth error
    |
    v
Decode JSON response
```

### Request body

If a request body exists, it is encoded as JSON:

```go
jsonBody, err := json.Marshal(params.RequestBody)
```

The resulting bytes are used as the HTTP request body:

```go
reader = bytes.NewReader(jsonBody)
```

The content type is then set:

```go
req.Header.Set("Content-Type", "application/json")
```

---

# 6. Context Propagation

The incoming request context should be passed to the Auth Service:

```go
req, err := http.NewRequestWithContext(
    ctx,
    params.Method,
    c.BaseURL+params.Path,
    reader,
)
```

The handler obtains the context from the original HTTP request:

```go
ctx := r.Context()
```

and passes it to the Auth client:

```go
response, err := h.auth.ValidateSession(ctx, sessionCookie)
```

This means that if the original request is cancelled or times out, the request to the Auth Service can also be cancelled.

---

# 7. Authentication Headers

Operations requiring authentication can provide headers through `RequestHeaders`.

For example:

```go
headers := map[string]string{
    "Authorization": "Bearer " + token,
}
```

The headers are added by `doRequest`:

```go
for key, value := range params.RequestHeaders {
    req.Header.Set(key, value)
}
```

This allows the same `doRequest` method to support both authenticated and unauthenticated endpoints.

For example:

```text
POST /v1/login

(no Authorization header)
```

while:

```text
POST /v1/session/validate

Authorization: Bearer <session-token>
```

---

# 8. Expected HTTP Status

Each operation defines the status code expected from the Auth Service.

Examples:

```go
http.StatusCreated // registration
http.StatusOK      // login/session validation
http.StatusNoContent // logout
```

The client verifies the returned status:

```go
if resp.StatusCode != params.ExpectedStatus {
    ...
}
```

This is important because a successful HTTP request does not necessarily mean that the operation itself was successful.

For example, `http.Client.Do()` can return:

```go
err == nil
```

while the Auth Service responds with:

```text
401 Unauthorized
```

Therefore, the HTTP status must always be checked explicitly.

---

# 9. Handling Auth Service Errors

When the Auth Service returns an unexpected status, the response body is expected to contain a JSON error:

```json
{
    "code": "INVALID_CREDENTIALS",
    "message": "invalid credentials"
}
```

The client decodes it:

```go
var responseError ResponseError

if err := json.NewDecoder(resp.Body).Decode(&responseError); err != nil {
    ...
}
```

The error is then converted into an application error:

```go
return fmt.Errorf(
    "%w: %s",
    e.ErrAuthService,
    responseError.Message,
)
```

Wrapping the error with `%w` allows the rest of the application to identify it using:

```go
errors.Is(err, e.ErrAuthService)
```

without depending on the exact error message.

---

# 10. Decoding Successful Responses

If the Auth Service returns an expected status and the operation has a response body, `doRequest` decodes the JSON response into the provided DTO:

```go
if params.RequestResult != nil {
    if err := json.NewDecoder(resp.Body).Decode(params.RequestResult); err != nil {
        return fmt.Errorf(
            "%w: %w",
            e.ErrJSONDecodeFailed,
            err,
        )
    }
}
```

The caller provides the destination:

```go
var result LoginUserResponseDTO

params := AuthClientRequestParams{
    ...
    RequestResult: &result,
}
```

The pointer is important because `json.Decoder.Decode` must be able to modify the destination.

---

# 11. Adding a New Auth Endpoint

To add another Auth Service endpoint, follow these steps.

### Step 1 — Create request/response DTOs

For example:

```go
type SomeRequestDTO struct {
    Value string `json:"value"`
}

type SomeResponseDTO struct {
    Result string `json:"result"`
}
```

Make sure the JSON tags match the Auth Service API.

### Step 2 — Add the method to `AuthInterface`

```go
type AuthInterface interface {
    ...
    SomeOperation(context.Context, SomeRequestDTO) (SomeResponseDTO, error)
}
```

### Step 3 — Implement it in `AuthHTTPClient`

```go
func (c *AuthHTTPClient) SomeOperation(
    ctx context.Context,
    dto SomeRequestDTO,
) (SomeResponseDTO, error) {

    var result SomeResponseDTO

    params := AuthClientRequestParams{
        Method:         http.MethodPost,
        Path:           "/v1/some-endpoint",
        RequestBody:    dto,
        ExpectedStatus: http.StatusOK,
        RequestResult:  &result,
    }

    if err := c.doRequest(ctx, params); err != nil {
        return SomeResponseDTO{}, err
    }

    return result, nil
}
```

### Step 4 — Use the interface from the handler/service

The handler should use:

```go
h.auth.SomeOperation(ctx, dto)
```

rather than constructing an HTTP request itself.

---

# 12. Dependency Injection

The Auth client is injected into the handler through its constructor.

The handler contains the interface:

```go
type CommentHandler struct {
    service *service.CommentService
    auth    client.AuthInterface
}
```

Constructor:

```go
func NewCommentHandler(
    service *service.CommentService,
    authService client.AuthInterface,
) *CommentHandler {
    return &CommentHandler{
        service: service,
        auth:    authService,
    }
}
```

This keeps the handler independent from the concrete HTTP client.

---

# 13. Using the Auth Client in a Handler

For endpoints that require authentication, the handler extracts the session token and validates it through the Auth interface.

Example:

```go
sessionCookie, err := helper.ExtractSessionCookie(r)
if err != nil {
    return err
}

response, err := h.auth.ValidateSession(
    r.Context(),
    sessionCookie,
)
if err != nil {
    return err
}

commentDTO := model.CreateCommentDTO{
    Comment: commentRaw,
    UserID:  int(response.User.ID),
}
```

The handler receives the authenticated user's ID from the Auth Service and uses it when creating the application object.

The important responsibility separation is:

```text
Handler
    |
    | extract session token
    v
Auth Client
    |
    | validate token
    v
Auth Service
    |
    | return authenticated user
    v
Handler
    |
    | pass UserID to service
    v
Comment Service
```

The handler should not implement session validation itself.

---

# 14. Dependency Wiring

During application startup, the concrete Auth client is created and injected.

For local testing, a mock implementation can be used:

```go
mockAuthClient := &client.MockAuthClient{}

dependencyWiring(mux, db, mockAuthClient)
```

The dependency wiring function receives the interface:

```go
func dependencyWiring(
    mux *http.ServeMux,
    db *sql.DB,
    auth client.AuthInterface,
) *http.ServeMux {

    commentsRepo := repository.NewSQLiteCommentRepository(db)
    commentService := service.NewCommentService(commentsRepo)
    commentHandler := handler.NewCommentHandler(
        commentService,
        auth,
    )

    RegisterRoutes(mux, commentHandler)

    return mux
}
```

This makes it possible to replace the real Auth Service with a mock without changing the handler.

---

# 15. Production Auth Client

When connecting to the real Auth Service, replace the mock with `AuthHTTPClient`:

```go
authClient := &client.AuthHTTPClient{
    BaseURL: "http://auth:8081",
    Client: &http.Client{
        Timeout: 3 * time.Second,
    },
}

dependencyWiring(mux, db, authClient)
```

The Docker Compose service name `auth` is used as the hostname because containers communicate through the Docker network.

The backend therefore calls:

```text
http://auth:8081
```

rather than:

```text
http://localhost:8081
```

inside the Docker network.

---

# 16. Recommended Request Flow

A typical authenticated endpoint should follow this structure:

```text
1. Receive HTTP request
        |
        v
2. Parse and validate request data
        |
        v
3. Extract session token
        |
        v
4. Call AuthInterface.ValidateSession()
        |
        v
5. AuthHTTPClient sends request to Auth Service
        |
        v
6. Auth Service validates session
        |
        v
7. Return authenticated User
        |
        v
8. Pass UserID to application service
        |
        v
9. Execute business operation
        |
        v
10. Return HTTP response
```

The handler should not contain low-level HTTP client logic.

---

# 17. Error Responsibility

Error handling should remain separated by responsibility:

### Auth client

Responsible for:

* HTTP communication errors;
* JSON encoding/decoding errors;
* unexpected responses from the Auth Service;
* converting Auth Service errors into application errors.

### Service

Responsible for:

* business rules;
* application-level validation;
* business errors.

### Repository

Responsible for:

* database operations;
* database errors;
* translating expected database conditions into application errors where appropriate.

### Global error handler

Responsible for:

* mapping application errors to HTTP status codes;
* logging unexpected internal errors;
* returning the final JSON error response.

This prevents handlers from becoming responsible for every layer of error handling.

---

# 18. Testing

Because handlers depend on `AuthInterface`, they can be tested with a mock:

```go
mockAuthClient := &client.MockAuthClient{}
```

The mock can simulate cases such as:

```text
Session is valid
Session is expired
Auth Service returns an error
User does not exist
```

This avoids requiring the real Auth Service to be running during handler unit tests.

The production implementation:

```go
*client.AuthHTTPClient
```

and the test implementation:

```go
*client.MockAuthClient
```

both satisfy:

```go
client.AuthInterface
```

which is the main benefit of using the interface.

---

# Summary

The integration should follow this architecture:

```text
                 ┌─────────────────────┐
                 │      Handler        │
                 └──────────┬──────────┘
                            │
                            │ AuthInterface
                            v
                 ┌─────────────────────┐
                 │   AuthHTTPClient    │
                 └──────────┬──────────┘
                            │
                       HTTP/JSON
                            │
                            v
                 ┌─────────────────────┐
                 │    Auth Service     │
                 └─────────────────────┘
```

The main implementation rules are:

1. Keep Auth Service communication inside `AuthHTTPClient`.
2. Expose operations through `AuthInterface`.
3. Keep Auth Service request/response formats in DTOs.
4. Reuse `doRequest` for common HTTP logic.
5. Always pass the request `context.Context`.
6. Check the expected HTTP status explicitly.
7. Decode Auth Service error responses when possible.
8. Inject `AuthInterface` into handlers/services instead of creating clients there.
9. Use a mock implementation for tests.
10. Create the real `AuthHTTPClient` during application dependency wiring.
