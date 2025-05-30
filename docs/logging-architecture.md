# ZTDP Logging System Architecture

This document provides a comprehensive overview of the Zero Touch Developer Platform's enhanced logging and event system.

## Overview

The ZTDP logging system provides real-time, structured event streaming with interactive web-based exploration capabilities. It transforms traditional log viewing into an interactive experience where users can explore events, understand system behavior, and gain insights through rich visualizations.

## Architecture Components

### 1. Event Generation & Broadcasting

#### Backend Event Creation (`api/handlers/logs.go`)

The logging system creates structured events through the `createEventHandler()` function:

```go
func createEventHandler() func(eventType string, data []byte) error {
    // Creates rich, structured events with:
    // - Event categorization (Application, Deployment, Policy, Resource, Connection)
    // - Descriptive messages with emojis
    // - Full payload information
    // - Timestamp and metadata
}
```

**Key Features:**
- **Smart Categorization**: Events are automatically categorized based on payload content
- **Rich Messaging**: Events include descriptive text with contextual emojis (🎯 🚀 📦 ✅ ❌ 🔒 🔍 🔧 🗑️)
- **Structured Data**: Every event includes full context, metadata, and structured payload

#### WebSocket Broadcasting (`internal/logging/realtime.go`)

The `BroadcastEvent()` method enables real-time event distribution:

```go
func (rt *RealTimeLogger) BroadcastEvent(eventData map[string]interface{}) {
    // Direct broadcasting of structured event data to WebSocket clients
    // Supports both regular logs and structured events
}
```

### 2. Real-time WebSocket Streaming

#### Endpoint: `/v1/logs/stream`

**Protocol Support:**
- **Regular Logs**: `{"type": "log.entry", "data": {...}}`
- **Structured Events**: `{"type": "event.structured", "data": {...}}`

**Event Structure:**
```json
{
  "type": "event.structured",
  "data": {
    "timestamp": "2024-12-09T10:30:00Z",
    "level": "info",
    "message": "🎯 Application checkout created successfully",
    "event_category": "Application Created",
    "event": {
      "type": "application.created",
      "source": "application",
      "subject": "checkout",
      "payload": { /* full event details */ }
    }
  }
}
```

### 3. Frontend Event Processing

#### WebSocket Message Handling (`static/graph-modern.html`)

The frontend distinguishes between regular logs and structured events:

```javascript
// Regular log handling
if (data.type === 'log.entry') {
    logsManager.addLogEntry(data.data);
}

// Structured event handling
if (data.type === 'event.structured') {
    logsManager.addStructuredEvent(data.data);
}
```

#### Interactive Log Display

**Features:**
- **Clickable Events**: Events with structured data are clickable and expandable
- **Rich Details**: Expanded view shows event type, source, subject, and full JSON payload
- **Visual Indicators**: Chevron icons indicate expandable content
- **Color Coding**: Different event types have distinct colors

#### Smart Filtering System

**Filter Categories:**
- **Log Levels**: ERROR, WARN, INFO, DEBUG
- **Event Types**: Application Created/Updated/Deleted, Deployment Events, Policy Events, Resource Events, Connection Events
- **Components**: application, deployment, policy, resource, websocket, event-subscriber, logs-websocket, system
- **Free Text**: Search across all event content

```javascript
function matchesEventType(logData, filterValue) {
    // Smart categorization logic that matches events to filter categories
    // Based on event payload content and patterns
}
```

### 4. Visual Styling System (`static/graph-modern.css`)

#### Event Display Classes

```css
.log-expandable {
    /* Base styling for clickable log entries */
    cursor: pointer;
    transition: background-color 0.2s ease;
}

.log-event-type {
    /* Event type tags with color coding */
    padding: 2px 6px;
    border-radius: 3px;
    font-size: 0.8em;
}

.log-details {
    /* Expandable details container */
    max-height: 0;
    overflow: hidden;
    transition: max-height 0.3s ease;
}

.log-expanded .log-details {
    /* Expanded state */
    max-height: 500px;
}
```

#### Color Coding System

- **Application Events**: Blue tones (#007acc)
- **Deployment Events**: Green tones (#28a745)
- **Policy Events**: Orange tones (#fd7e14)
- **Resource Events**: Purple tones (#6f42c1)
- **Connection Events**: Teal tones (#20c997)

## Event Categories & Classification

### Automatic Event Categorization

Events are automatically classified based on payload analysis:

1. **Application Created/Updated/Deleted**: Events involving application lifecycle
2. **Deployment Events**: Service deployments, version updates, environment changes
3. **Policy Events**: Policy enforcement, validation, compliance checks
4. **Resource Events**: Infrastructure provisioning, configuration changes
5. **Connection Events**: WebSocket connections, client interactions

### Classification Logic

```javascript
// Example categorization patterns
if (payload.application || payload.app) {
    if (eventType.includes('created')) return 'Application Created';
    if (eventType.includes('updated')) return 'Application Updated';
    if (eventType.includes('deleted')) return 'Application Deleted';
}

if (payload.environment || payload.deployment) {
    return 'Deployment Events';
}

// Additional patterns for policy, resource, and connection events
```

## Usage Examples

### Accessing the Logging System

1. **Start the ZTDP API server**:
   ```bash
   go run ./cmd/api/main.go
   ```

2. **Open the web interface**:
   ```
   http://localhost:8080/static/graph-modern.html
   ```

3. **Generate events by interacting with the API**:
   ```bash
   # Create an application (generates Application Created event)
   curl -X POST http://localhost:8080/v1/applications \
     -H "Content-Type: application/json" \
     -d '{"metadata": {"name": "test-app", "owner": "team-x"}}'
   ```

### Event Exploration Workflow

1. **Real-time Monitoring**: Events appear in real-time as they occur
2. **Filtering**: Use the filter controls to focus on specific event types or components
3. **Event Details**: Click on structured events to view full details
4. **Payload Analysis**: Examine JSON payloads to understand event context
5. **System Insights**: Use event patterns to understand system behavior

## Development Patterns

### Adding New Event Types

1. **Define Event Structure** in your handler:
   ```go
   eventData := map[string]interface{}{
       "type":    "your.event.type",
       "source":  "your-component",
       "subject": "your-subject",
       "payload": yourPayload,
   }
   ```

2. **Broadcast Event**:
   ```go
   rt.BroadcastEvent(eventData)
   ```

3. **Update Frontend Categorization** if needed in `matchesEventType()`

### Styling New Event Types

Add CSS rules for new event categories:

```css
.log-event-type.your-event-category {
    background-color: #your-color;
    color: white;
}
```

## Performance Considerations

### WebSocket Scaling

- **Connection Management**: Monitor WebSocket connection count and implement connection limits if needed
- **Event Volume**: Consider event batching for high-volume scenarios
- **Memory Usage**: Implement client-side log rotation for long-running sessions

### Frontend Performance

- **Virtual Scrolling**: Consider implementing virtual scrolling for large log volumes
- **Event Filtering**: Client-side filtering reduces server load
- **State Management**: Proper cleanup of event listeners and DOM elements

## Security Considerations

### Event Data Sensitivity

- **Payload Sanitization**: Ensure sensitive data is not exposed in event payloads
- **Access Control**: Consider implementing authentication for WebSocket endpoints
- **Data Retention**: Implement appropriate log retention policies

### WebSocket Security

- **Origin Validation**: Validate WebSocket origin headers
- **Rate Limiting**: Implement rate limiting for WebSocket connections
- **Error Handling**: Graceful handling of connection failures and reconnections

## Future Enhancements

### Planned Improvements

1. **Event Schema Validation**: Implement JSON schema validation for events
2. **Historical Event Storage**: Persist events for historical analysis
3. **Advanced Analytics**: Add event aggregation and trend analysis
4. **Export Capabilities**: Enable event export in various formats
5. **Internationalization**: Add multi-language support for the UI

### Integration Opportunities

1. **External Monitoring**: Integration with external monitoring systems
2. **Alerting**: Event-based alerting and notification systems
3. **Compliance**: Audit trail generation for compliance requirements
4. **Machine Learning**: Event pattern analysis and anomaly detection

## Troubleshooting

### Common Issues

1. **WebSocket Connection Failures**:
   - Check server logs for connection errors
   - Verify firewall and proxy configurations
   - Ensure proper CORS settings

2. **Missing Events**:
   - Verify event broadcasting is properly configured
   - Check event categorization logic
   - Confirm WebSocket connection is active

3. **Performance Issues**:
   - Monitor client-side memory usage
   - Implement log rotation if needed
   - Consider reducing event detail level

### Debugging Tools

1. **Browser Developer Tools**: Monitor WebSocket traffic and console logs
2. **Server Logs**: Check API server logs for WebSocket errors
3. **Event Inspection**: Use browser console to inspect event data structures

## Conclusion

The ZTDP logging system represents a significant advancement in platform observability, transforming traditional log viewing into an interactive, insightful experience. By combining real-time streaming, smart categorization, and rich visualization, it provides developers and operators with the tools they need to understand and monitor complex distributed systems effectively.

The architecture is designed for extensibility, performance, and ease of use, making it a solid foundation for future enhancements and integrations.
