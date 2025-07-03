# Test plugin Plugin

A nixai plugin

## Installation

1. Build the plugin:
   ```bash
   go build -buildmode=plugin -o test-plugin.so .
   ```

2. Install the plugin:
   ```bash
   nixai plugin install test-plugin.so
   ```

## Usage

```bash
# Say hello
nixai plugin execute test-plugin hello

# Say hello to someone specific
nixai plugin execute test-plugin hello --params '{"name": "Alice"}'
```

## Operations

- **hello**: Say hello to someone

## Author

Plugin Developer

## License

MIT
