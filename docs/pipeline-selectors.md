# Pipeline Selectors

Pipeline selectors provide a flexible mechanism for matching contracts with pipelines without requiring explicit pipeline IDs. This allows users to define contracts based on what they want to happen to their data rather than knowing specific pipeline names.

## Overview

Instead of specifying a fixed pipeline ID, contracts can now use pipeline selectors to dynamically match with appropriate pipelines based on criteria such as:

- Pipeline type (trace, metric, log)
- Pipeline name patterns
- Environment tags
- Processing metadata
- Custom tags and attributes

## Basic Usage

### Traditional Approach (Explicit Pipeline ID)

```yaml
publisher: "auth-service"
pipeline: "traces"  # Fixed pipeline ID
version: "1.0.0"
# ... rest of contract
```

### New Approach (Pipeline Selectors)

```yaml
publisher: "auth-service"
version: "1.0.0"
pipeline_selectors:
  selectors:
    - field: "type"
      operator: "equals"
      value: "trace"
    - field: "tags.environment"
      operator: "equals"
      value: "production"
  priority: 1
# ... rest of contract
```

## Selector Operators

The following operators are supported for pipeline matching:

### `equals`
Exact string match.

```yaml
- field: "type"
  operator: "equals"
  value: "trace"
```

### `matches`
Regex pattern matching.

```yaml
- field: "name"
  operator: "matches"
  value: "(?i).*auth.*"  # Case-insensitive match for "auth"
```

### `contains`
Substring matching.

```yaml
- field: "name"
  operator: "contains"
  value: "User"
```

### `starts_with`
Prefix matching.

```yaml
- field: "tags.datacenter"
  operator: "starts_with"
  value: "us-"
```

### `ends_with`
Suffix matching.

```yaml
- field: "name"
  operator: "ends_with"
  value: "Pipeline"
```

## Supported Fields

Pipeline selectors can match against the following fields:

### Basic Fields
- `id` - Pipeline identifier
- `name` - Pipeline name
- `description` - Pipeline description
- `type` - Pipeline type (trace, metric, log)

### Tags
- `tags.<key>` - Any tag value (e.g., `tags.environment`, `tags.service`)

### Metadata
- `metadata.<key>` - Any metadata value (e.g., `metadata.processing_type`)

## Examples

### Match Production Trace Pipelines

```yaml
pipeline_selectors:
  selectors:
    - field: "type"
      operator: "equals"
      value: "trace"
    - field: "tags.environment"
      operator: "equals"
      value: "production"
```

### Match Auth Service Pipelines

```yaml
pipeline_selectors:
  selectors:
    - field: "name"
      operator: "contains"
      value: "auth"
    - field: "tags.service"
      operator: "equals"
      value: "auth"
```

### Match Aggregation Pipelines

```yaml
pipeline_selectors:
  selectors:
    - field: "type"
      operator: "equals"
      value: "metric"
    - field: "metadata.processing_type"
      operator: "equals"
      value: "aggregation"
```

### Complex Pattern Matching

```yaml
pipeline_selectors:
  selectors:
    - field: "name"
      operator: "matches"
      value: "(?i).*user.*service.*"
    - field: "tags.datacenter"
      operator: "starts_with"
      value: "us-"
    - field: "tags.environment"
      operator: "equals"
      value: "production"
```

## Priority

Selectors support a priority field to help resolve conflicts when multiple pipelines match:

```yaml
pipeline_selectors:
  selectors:
    - field: "type"
      operator: "equals"
      value: "trace"
  priority: 1  # Higher priority selectors are preferred
```

## Migration from Explicit Pipeline IDs

To migrate existing contracts:

1. **Keep existing contracts working**: Contracts with explicit `pipeline` fields continue to work
2. **Add selectors gradually**: You can add `pipeline_selectors` alongside existing `pipeline` fields
3. **Selectors take precedence**: When both are present, `pipeline_selectors` takes precedence

### Migration Example

**Before:**
```yaml
publisher: "auth-service"
pipeline: "traces"
version: "1.0.0"
```

**After:**
```yaml
publisher: "auth-service"
pipeline: "traces"  # Can be kept for backward compatibility
version: "1.0.0"
pipeline_selectors:
  selectors:
    - field: "type"
      operator: "equals"
      value: "trace"
    - field: "tags.environment"
      operator: "equals"
      value: "production"
```

## Best Practices

1. **Be specific**: Use multiple selectors to ensure precise matching
2. **Use meaningful tags**: Tag your pipelines with relevant metadata
3. **Test selectors**: Verify that your selectors match the intended pipelines
4. **Use priority**: Set appropriate priorities for complex scenarios
5. **Document patterns**: Document the expected pipeline naming and tagging conventions

## Pipeline Registration

For pipeline selectors to work, pipelines must be registered with the selector service. Pipelines should include:

- Descriptive names
- Relevant tags (environment, service, datacenter, etc.)
- Processing metadata
- Type information

Example pipeline registration:

```go
pipeline := &contract.PipelineInfo{
    ID:          "trace-auth-prod",
    Name:        "Auth Service Trace Pipeline",
    Description: "Processes trace data from auth service in production",
    Type:        "trace",
    Tags: map[string]string{
        "environment": "production",
        "service":     "auth",
        "datacenter":  "us-east-1",
    },
    Metadata: map[string]string{
        "processing_type": "validation",
    },
}
```
