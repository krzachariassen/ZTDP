# Simple Application Deployment API

## Overview

We have successfully implemented a simple, user-friendly application deployment API that abstracts away the complexity of service-by-service deployment.

## New Deployment Endpoint

### POST `/v1/applications/{app_name}/deploy`

Deploy an entire application with all its services to a target environment in a single API call.

#### Request Body
```json
{
  "environment": "dev",      // Required: Target environment
  "version": "1.0.0"        // Optional: Future feature for app versioning
}
```

#### Success Response (200 OK)
```json
{
  "application": "my-app",
  "environment": "dev",
  "version": "1.0.0",
  "deployments": [
    "service-a:1.0.0",
    "service-b:1.0.0"
  ],
  "skipped": [],
  "failed": [],
  "summary": {
    "total_services": 2,
    "deployed": 2,
    "skipped": 0,
    "failed": 0,
    "success": true,
    "message": "Successfully deployed my-app to dev"
  }
}
```

#### Error Responses

**400 Bad Request** - Missing or invalid parameters
```json
{
  "error": "Environment is required"
}
```

**403 Forbidden** - Application not allowed to deploy to environment
```json
{
  "error": "Application 'my-app' is not allowed to deploy to environment 'prod'"
}
```

**404 Not Found** - Application or environment not found
```json
{
  "error": "Application not found"
}
```

## Before vs After

### Before: Complex Multi-Step Process
Users had to:
1. Get the application plan: `GET /v1/applications/{app}/plan`
2. Deploy each service individually: `POST /v1/applications/{app}/services/{service}/versions/{version}/deploy`
3. Handle policy enforcement manually
4. Track deployment status across multiple calls

### After: Simple Single API Call
Users now can:
1. Deploy entire application: `POST /v1/applications/{app}/deploy`
2. Get comprehensive deployment results
3. Automatic policy enforcement
4. Clear success/failure reporting

## Example Usage

```bash
# Deploy application to development
curl -X POST http://localhost:8080/v1/applications/checkout/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "dev"}'

# Deploy application to production (with version)
curl -X POST http://localhost:8080/v1/applications/checkout/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "prod", "version": "2.1.0"}'
```

## Internal Implementation

The new deployment handler:

1. **Validates Input**: Checks for required environment parameter
2. **Validates Application**: Ensures application exists in the graph
3. **Validates Environment**: Ensures target environment exists
4. **Enforces Policies**: Checks application is allowed to deploy to environment
5. **Generates Plan**: Uses internal planner to determine deployment order
6. **Executes Deployment**: Deploys all service versions according to plan
7. **Returns Results**: Provides comprehensive deployment summary

## Policy Integration

The deployment API fully integrates with the existing policy system:
- Respects environment access controls
- Enforces deployment restrictions
- Provides clear error messages for policy violations

## Backward Compatibility

The existing granular deployment APIs remain available for advanced users who need fine-grained control:
- `/v1/applications/{app}/services/{service}/versions/{version}/deploy`
- `/v1/applications/{app}/plan/apply/{env}`

## Testing

Comprehensive tests have been added to validate:
- Successful deployments
- Policy enforcement
- Error handling
- Response structure

Run tests with:
```bash
go test ./test/api/ -v
```
