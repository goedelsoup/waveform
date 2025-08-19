# Advanced Contract Features

This directory contains examples demonstrating the sophisticated validation capabilities of Waveform's advanced contract system. These features enable complex, production-ready contract testing scenarios.

## Overview of Advanced Features

### 🎯 Enhanced Validation Rules

Advanced validation rules provide sophisticated matching and validation capabilities beyond simple equality checks.

#### Supported Operators

- **Basic Comparison**: `equals`, `not_equals`, `greater_than`, `less_than`, `greater_or_equal`, `less_or_equal`
- **Pattern Matching**: `matches`, `not_matches` (regex support)
- **Existence Checks**: `exists`, `not_exists`
- **String Operations**: `contains`, `not_contains`, `starts_with`, `ends_with`
- **Range Operations**: `in_range`, `not_in_range`
- **Set Operations**: `one_of`, `not_one_of`

#### Range Validation

```yaml
validation_rules:
  - field: "span.attributes.payment.amount"
    operator: "in_range"
    range:
      min: 0.01
      max: 10000.00
      inclusive: true
      min_inclusive: true  # Optional override
      max_inclusive: false # Optional override
    description: "Payment amount validation"
    severity: "error"
```

#### Pattern Matching

```yaml
validation_rules:
  - field: "span.attributes.transaction.id"
    operator: "matches"
    pattern: "^txn_[0-9a-f]{32}$"
    description: "Transaction ID format validation"
    severity: "warning"
```

### 🔀 Conditional Validation Logic

Implement complex business rules with conditional validation.

#### If-Then-Else Logic

```yaml
validation_rules:
  - field: "span.attributes.payment.cvv"
    operator: "exists"
    condition:
      if:
        field: "span.attributes.payment.method"
        operator: "equals"
        value: "credit_card"
      then:
        field: "span.attributes.payment.cvv"
        operator: "matches"
        pattern: "^[0-9]{3,4}$"
    description: "CVV required for credit card payments"
```

#### Boolean Logic (AND/OR/NOT)

```yaml
validation_rules:
  - field: "span.attributes.fraud.score"
    operator: "less_or_equal"
    value: 0.8
    condition:
      or:
        - field: "span.attributes.payment.amount"
          operator: "greater_than"
          value: 1000
        - field: "span.attributes.customer.risk_level"
          operator: "equals"
          value: "high"
    description: "High-value or high-risk payments need low fraud scores"
```

### ⏱️ Temporal Validation

Time-based validation rules for performance monitoring and SLA compliance.

```yaml
validation_rules:
  - field: "span.duration"
    operator: "less_than"
    value: "500ms"
    temporal:
      window_size: "5m"
      aggregation: "p95"
      threshold: "300ms"
      comparison: "less_than"
      baseline: "previous_week"
      tolerance: 0.15
    description: "P95 latency SLA validation"
```

#### Temporal Aggregations

- `count`, `sum`, `avg`, `min`, `max`
- `p50`, `p90`, `p95`, `p99` (percentiles)
- `stddev`, `variance`

#### Baseline Comparisons

- `previous_hour`, `previous_day`, `previous_week`
- `same_hour_yesterday`, `same_day_last_week`
- Custom time ranges

### 🔄 Transformation Validation

Validate that data transformations occurred as expected.

```yaml
validation_rules:
  - field: "span.attributes.order.total_cents"
    operator: "exists"
    transform:
      type: "add"
      source: "span.attributes.order.total"
      target: "span.attributes.order.total_cents"
      function: "multiply"
      parameters:
        factor: 100
    description: "Currency conversion validation"
```

#### Transformation Types

- `add`: Field was added
- `remove`: Field was removed
- `modify`: Field value was changed
- `rename`: Field was renamed

### 📊 Advanced Matchers

Enhanced matchers for specific telemetry signal types.

#### Trace Matchers

```yaml
matchers:
  traces:
    - span_name: "process_payment"
      count:
        expected: 1
        min: 1
        max: 5
      duration:
        min: "50ms"
        max: "5s"
        expected: "200ms"
        tolerance: "100ms"
      status_code:
        class: "2xx"
        not_allowed: [400, 401, 403, 500]
```

#### Metric Matchers

```yaml
matchers:
  metrics:
    - name: "http_request_duration_seconds"
      value:
        range:
          min: 0.001
          max: 1.000
        tolerance: 0.05
      histogram:
        buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0]
        count: 100
        sum: 15.5
```

#### Log Matchers

```yaml
matchers:
  logs:
    - body: "Order completed successfully"
      timestamp:
        format: "RFC3339"
        relative: "within_last_minute"
        precision: "millisecond"
      count:
        expected: 1
```

### 🎚️ Validation Severity Levels

Control validation impact with severity levels.

```yaml
validation_rules:
  - field: "span.attributes.required_field"
    operator: "exists"
    severity: "error"    # Fails the contract

  - field: "span.attributes.recommended_field"
    operator: "exists"
    severity: "warning"  # Warning only

  - field: "span.attributes.optional_field"
    operator: "exists"
    severity: "info"     # Informational
```

### 📋 Schema Validation

Validate contract structure and syntax.

```yaml
schema:
  version: "2.0"
  required_fields: ["publisher", "version", "inputs", "matchers"]
  field_types:
    publisher: "string"
    version: "string"
  validation_rules:
    - field: "publisher"
      type: "string"
      required: true
      pattern: "^[a-z][a-z0-9-]*[a-z0-9]$"
      min_length: 3
      max_length: 50
```

### 🧬 Contract Inheritance

Reuse and compose contracts for maintainability.

```yaml
inheritance:
  extends: ["base-contract"]          # Inherit from parent contracts
  includes: ["common-validations"]    # Include reusable components
  mixins: ["performance-validations"] # Apply mixin behaviors
  overrides:
    validation_rules:
      severity: "error"               # Override inherited settings
```

## Example Contracts

### 1. [conditional-validation.yaml](./conditional-validation.yaml)
Demonstrates conditional validation logic, range validation, and pattern matching for a payment service.

**Key Features:**
- Conditional CVV validation for credit cards
- Payment amount range validation
- Transaction ID pattern matching
- Complex AND/OR conditional logic
- Temporal validation for processing times

### 2. [multi-signal-validation.yaml](./multi-signal-validation.yaml)
Shows validation across traces, metrics, and logs with cross-signal correlation.

**Key Features:**
- Cross-signal consistency validation
- Temporal correlation between signals
- Metric increment validation
- Log enrichment validation
- Multi-signal transformation validation

### 3. [temporal-validation.yaml](./temporal-validation.yaml)
Focuses on time-based validation for performance monitoring and SLA compliance.

**Key Features:**
- P95 latency SLA validation
- Anomaly detection based on historical patterns
- Peak hour performance validation
- Histogram bucket validation
- Performance tier classification

## Running the Examples

### Prerequisites

```bash
# Ensure you have the Waveform framework set up
go mod tidy
```

### Run Advanced Contract Example

```bash
cd examples/advanced-contracts
go run main.go
```

This will demonstrate:
- Loading advanced contracts
- Processing data through sophisticated pipelines
- Validating against complex rules
- Reporting detailed validation results

### Test Individual Contracts

```bash
# Test conditional validation
waveform --contracts "examples/contracts/advanced/conditional-validation.yaml" --mode processor

# Test multi-signal validation
waveform --contracts "examples/contracts/advanced/multi-signal-validation.yaml" --mode pipeline

# Test temporal validation
waveform --contracts "examples/contracts/advanced/temporal-validation.yaml" --mode pipeline
```

## Best Practices

### 1. **Start Simple, Add Complexity Gradually**
Begin with basic validation rules and progressively add advanced features as needed.

### 2. **Use Appropriate Severity Levels**
- `error`: Critical validations that must pass
- `warning`: Important but non-blocking validations
- `info`: Informational validations for monitoring

### 3. **Leverage Conditional Logic**
Use conditional validation to implement business rules and context-aware validation.

### 4. **Monitor Performance with Temporal Rules**
Implement SLA monitoring and performance regression detection with temporal validation.

### 5. **Design for Reusability**
Use contract inheritance and mixins to share common validation patterns across services.

### 6. **Document Complex Rules**
Always include descriptive text for complex validation rules to aid debugging and maintenance.

## Advanced Use Cases

### Performance Monitoring
- SLA compliance checking
- Performance regression detection
- Anomaly detection
- Capacity planning validation

### Security Validation
- Input validation and sanitization
- Authentication flow validation
- Authorization checks
- Audit trail validation

### Business Rule Enforcement
- Payment processing rules
- Order fulfillment validation
- Customer tier handling
- Promotional logic validation

### Data Quality Assurance
- Field format validation
- Data completeness checks
- Consistency validation across signals
- Transformation verification

## Migration from Basic Contracts

To migrate from basic contracts to advanced contracts:

1. **Add validation_rules section** with sophisticated rules
2. **Enhance matchers** with count, duration, and other advanced features
3. **Add schema validation** to ensure contract correctness
4. **Implement temporal rules** for performance monitoring
5. **Use conditional logic** for business rule enforcement

The advanced contract system is fully backward compatible with existing basic contracts.
