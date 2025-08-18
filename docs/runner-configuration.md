# Waveform Runner Configuration

The Waveform runner supports a flexible configuration system that honors the XDG Base Directory Specification and supports both YAML and TOML formats.

## Configuration File Locations

Waveform looks for configuration files in the following order (highest priority first):

1. **Current Directory**:
   - `.waveform.yaml`
   - `waveform.yaml`
   - `.waveform.toml`
   - `waveform.toml`

2. **XDG Config Directory**:
   - `$XDG_CONFIG_HOME/waveform/config.yaml`
   - `$XDG_CONFIG_HOME/waveform/config.toml`
   - `~/.config/waveform/config.yaml` (default XDG location)
   - `~/.config/waveform/config.toml` (default XDG location)

3. **Legacy Home Directory**:
   - `~/.waveform.yaml`
   - `~/waveform.yaml`
   - `~/.waveform.toml`
   - `~/waveform.toml`

## Configuration Format

The runner configuration supports both YAML and TOML formats. Choose the format that best fits your preferences and tooling.

### YAML Format Example

```yaml
# Runner settings
runner:
  # Logging configuration
  log_level: info                    # debug, info, warn, error
  log_format: json                   # json, console
  
  # Test execution settings
  timeout: 30s                       # Default timeout for test execution
  parallel: 1                        # Number of parallel test executions
  
  # Output settings
  output:
    formats: ["summary", "junit", "lcov"]  # Report formats to generate
    directory: "./waveform-reports"        # Output directory for reports
    overwrite: false                       # Whether to overwrite existing files
    verbose: false                         # Verbose output
  
  # Cache settings
  cache:
    enabled: true                    # Whether to enable caching
    directory: "~/.cache/waveform"   # Cache directory
    ttl: 1h                         # Cache TTL
    max_size: 100                   # Maximum cache size in MB

# Collector definitions for matching system
collectors:
  production:
    name: "production-collector"
    description: "Production environment collector"
    config_path: "/etc/collector/production.yaml"
    tags: ["production", "main", "telemetry"]
    
    # Environment-specific settings
    environment:
      datacenter: "us-west-1"
      tier: "production"
      retention_days: 30
    
    # Pipeline configurations
    pipelines:
      traces:
        name: "traces-pipeline"
        description: "Trace processing pipeline for production"
        signals: ["traces"]
        selectors:
          - field: "type"
            operator: "equals"
            value: "trace"
            priority: 1
          - field: "environment"
            operator: "equals"
            value: "production"
            priority: 2

# Global pipeline selectors for dynamic pipeline matching
pipeline_selectors:
  - field: "service.name"
    operator: "matches"
    value: "auth|payment|user"
    priority: 10
  
  - field: "environment"
    operator: "equals"
    value: "production"
    priority: 5

# Global settings
global:
  environment: development
  default_timeout: 30s
  fail_fast: false
  retry:
    max_attempts: 3
    initial_backoff: 1s
    max_backoff: 30s
    backoff_multiplier: 2.0
```

### TOML Format Example

```toml
[runner]
log_level = "info"
log_format = "json"
timeout = "30s"
parallel = 1

[runner.output]
formats = ["summary", "junit", "lcov"]
directory = "./waveform-reports"
overwrite = false
verbose = false

[runner.cache]
enabled = true
directory = "~/.cache/waveform"
ttl = "1h"
max_size = 100

[collectors.production]
name = "production-collector"
description = "Production environment collector"
config_path = "/etc/collector/production.yaml"
tags = ["production", "main", "telemetry"]

[collectors.production.environment]
datacenter = "us-west-1"
tier = "production"
retention_days = 30

[collectors.production.pipelines.traces]
name = "traces-pipeline"
description = "Trace processing pipeline for production"
signals = ["traces"]

[[collectors.production.pipelines.traces.selectors]]
field = "type"
operator = "equals"
value = "trace"
priority = 1

[[collectors.production.pipelines.traces.selectors]]
field = "environment"
operator = "equals"
value = "production"
priority = 2

[[pipeline_selectors]]
field = "service.name"
operator = "matches"
value = "auth|payment|user"
priority = 10

[[pipeline_selectors]]
field = "environment"
operator = "equals"
value = "production"
priority = 5

[global]
environment = "development"
default_timeout = "30s"
fail_fast = false

[global.retry]
max_attempts = 3
initial_backoff = "1s"
max_backoff = "30s"
backoff_multiplier = 2.0
```

## Configuration Sections

### Runner Settings

The `runner` section configures the Waveform runner behavior:

- **`log_level`**: Logging level (`debug`, `info`, `warn`, `error`)
- **`log_format`**: Log format (`json`, `console`)
- **`timeout`**: Default timeout for test execution
- **`parallel`**: Number of parallel test executions
- **`output`**: Output configuration (see below)
- **`cache`**: Cache configuration (see below)

### Output Settings

The `output` section configures report generation:

- **`formats`**: List of report formats to generate (`summary`, `junit`, `lcov`)
- **`directory`**: Output directory for reports
- **`overwrite`**: Whether to overwrite existing files
- **`verbose`**: Enable verbose output

### Cache Settings

The `cache` section configures caching behavior:

- **`enabled`**: Whether to enable caching
- **`directory`**: Cache directory
- **`ttl`**: Cache TTL (time-to-live)
- **`max_size`**: Maximum cache size in MB

### Collector Definitions

The `collectors` section defines collector configurations for the matching system:

Each collector can have:
- **`name`**: Collector name/identifier
- **`description`**: Collector description
- **`config_path`**: Path to the collector configuration file
- **`environment`**: Environment-specific settings
- **`tags`**: Tags for categorization
- **`pipelines`**: Pipeline configurations (see below)

### Pipeline Configurations

Each pipeline within a collector can have:
- **`name`**: Pipeline name
- **`description`**: Pipeline description
- **`signals`**: Signal types this pipeline handles (`traces`, `metrics`, `logs`)
- **`selectors`**: Pipeline selectors for matching (see below)
- **`environment`**: Environment-specific overrides

### Pipeline Selectors

Pipeline selectors define criteria for dynamic pipeline matching:

- **`field`**: Field to match against
- **`operator`**: Operator for comparison (`equals`, `not_equals`, `matches`, `exists`, `not_exists`, `greater_than`, `less_than`)
- **`value`**: Value to match against
- **`priority`**: Priority for this selector (higher = more specific)

### Global Settings

The `global` section contains global configuration:

- **`environment`**: Default environment
- **`default_timeout`**: Default timeout for operations
- **`fail_fast`**: Whether to fail fast on errors
- **`retry`**: Retry configuration (see below)

### Retry Settings

The `retry` section configures retry behavior:

- **`max_attempts`**: Maximum number of retries
- **`initial_backoff`**: Initial backoff duration
- **`max_backoff`**: Maximum backoff duration
- **`backoff_multiplier`**: Backoff multiplier

## Usage Examples

### Basic Configuration

Create a simple configuration file in your project:

```yaml
# .waveform.yaml
runner:
  log_level: info
  timeout: 30s
  output:
    formats: ["summary"]
    directory: "./reports"

global:
  environment: development
```

### Environment-Specific Configuration

Use different configurations for different environments:

```yaml
# Production configuration
collectors:
  production:
    name: "production-collector"
    config_path: "/etc/collector/production.yaml"
    pipelines:
      traces:
        name: "production-traces"
        signals: ["traces"]
        selectors:
          - field: "environment"
            operator: "equals"
            value: "production"
            priority: 10

global:
  environment: production
  fail_fast: true
```

### Pipeline Matching

Configure dynamic pipeline matching based on telemetry characteristics:

```yaml
pipeline_selectors:
  # High-priority selectors for critical services
  - field: "service.name"
    operator: "matches"
    value: "auth|payment|user"
    priority: 10
  
  # Environment-based selectors
  - field: "environment"
    operator: "equals"
    value: "production"
    priority: 5
  
  # Signal type selectors
  - field: "type"
    operator: "equals"
    value: "trace"
    priority: 2
```

## Integration with Contracts

The runner configuration works with contract files to enable dynamic pipeline matching:

1. **Contract Definition**: Contracts can use `pipeline_selectors` instead of explicit pipeline IDs
2. **Dynamic Matching**: The runner uses selectors to match contracts with appropriate pipelines
3. **Environment Support**: Different environments can have different collector configurations
4. **Priority System**: Higher priority selectors take precedence over lower priority ones

### Example Contract with Selectors

```yaml
# contract.yaml
publisher: "auth-service"
version: "1.0.0"

# Use pipeline selectors instead of explicit pipeline ID
pipeline_selectors:
  selectors:
    - field: "type"
      operator: "equals"
      value: "trace"
    - field: "service.name"
      operator: "equals"
      value: "auth-service"
    - field: "environment"
      operator: "equals"
      value: "production"
  priority: 5

inputs:
  traces:
    - span_name: "http_request"
      service_name: "auth-service"
      attributes:
        http.method: "POST"
        http.url: "/auth/login"

matchers:
  traces:
    - span_name: "http_request"
      attributes:
        http.method: "POST"
        normalized.method: "post"
```

## Best Practices

1. **Use XDG Directories**: Store global configurations in `~/.config/waveform/`
2. **Project-Specific Configs**: Use `.waveform.yaml` in project directories for project-specific settings
3. **Environment Separation**: Use different collector configurations for different environments
4. **Priority Management**: Use priority values to control selector precedence
5. **Tagging**: Use tags to categorize and organize collectors
6. **Documentation**: Include descriptions for all collectors and pipelines

## Migration from Command Line Flags

The runner configuration can replace many command line flags:

| Command Line Flag | Configuration Option |
|------------------|---------------------|
| `--verbose` | `runner.output.verbose: true` |
| `--junit-output` | `runner.output.formats: ["junit"]` |
| `--lcov-output` | `runner.output.formats: ["lcov"]` |
| `--summary-output` | `runner.output.formats: ["summary"]` |

Command line flags still take precedence over configuration file settings for backward compatibility.
