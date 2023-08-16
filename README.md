# Lib go

- A versatile and modular Go utility library that provides essential building blocks for building robust and maintainable applications.
This library offers features such as circuit breakers, health checks, HTTP clients, rate limiters, logging utilities, 
MongoDB and Redis clients, graceful shutdown management, startup helpers, and more.

## Project Structure

### Directory structure

```text
├── pkg                   # Reusable utility packages
│   ├── circuitbreaker    # Circuit breaker pattern implementation
│   ├── healthcheck       # Health check utilities
│   ├── httpclient        # HTTP client utility
│   ├── limiter           # Rate limiting functionality
│   ├── logger            # Logging utilities
│   ├── mongodb           # MongoDB client utilities
│   ├── redis             # Redis client utilities
│   ├── shutdown          # Graceful shutdown management
│   └── startup           # Startup helper utilities
├── README.md             # Project description and instructions
├── app.go                # Main application entry point
├── config.go             # Application configuration
```

