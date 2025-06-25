# k6s Configuration Examples

This directory contains example configuration files for different use cases.

## Available Examples

### Basic Configuration (`basic.yaml`)
Minimal configuration for getting started with k6s. Good for testing and learning.

```bash
k6s --config=examples/config/basic.yaml deployment list --watch
```

### Production Configuration (`production.yaml`)
Optimized configuration for production environments with:
- Enhanced monitoring settings
- Higher performance parameters
- Detailed change analysis
- Label filtering for specific applications

```bash
k6s --config=examples/config/production.yaml deployment list --watch
```

### Development Configuration (`development.yaml`)
Configuration optimized for local development with:
- Verbose logging for debugging
- Fast feedback loops
- All namespaces monitoring
- Development-specific label filtering

```bash
k6s --config=examples/config/development.yaml deployment list --watch
```

### Multi-namespace Configuration (`multi-namespace.yaml`)
Balanced configuration for monitoring multiple namespaces with:
- Cross-namespace deployment tracking
- Application-focused filtering
- Moderate performance settings

```bash
k6s --config=examples/config/multi-namespace.yaml deployment list --watch
```

### Complete Example (`config.yaml.example`)
Full configuration file with all available options and detailed comments. Use this as a reference for creating your own configuration.

```bash
# Copy to your preferred location and customize
cp examples/config/config.yaml.example ~/.k6s/k6s.yaml
```

## Usage

1. **Copy an example**: Choose the configuration that best matches your use case
2. **Customize**: Modify the settings according to your environment
3. **Use with k6s**: Reference the config file with the `--config` flag

```bash
# Create your own config directory
mkdir -p ~/.k6s

# Copy and customize an example
cp examples/config/production.yaml ~/.k6s/k6s.yaml

# Edit the configuration
vim ~/.k6s/k6s.yaml

# Use with k6s (will automatically find ~/.k6s/k6s.yaml)
k6s deployment list --watch
```

## Configuration Locations

k6s automatically looks for configuration files in these locations (in order):

1. File specified with `--config` flag
2. `./k6s.yaml` (current directory)
3. `~/.k6s/k6s.yaml` (user home directory)
4. `/etc/k6s/k6s.yaml` (system-wide)

## Environment Variables

All configuration options can be overridden using environment variables with the `K6S_` prefix:

```bash
# Override log level
export K6S_LOG_LEVEL=debug

# Override namespace
export K6S_INFORMER_NAMESPACE=production

# Override custom logic setting
export K6S_INFORMER_ENABLE_CUSTOM_LOGIC=true
```

## Configuration Validation

k6s validates configuration on startup. If there are validation errors, the application will exit with a descriptive error message.

Common validation rules:
- `resync_period` must be positive
- `worker_pool_size` must be positive
- `queue_size` must be positive
- `timeout` must be positive
- `max_retries` cannot be negative
