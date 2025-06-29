# nixai config

Manage nixai configuration settings.

---

## Command Help Output

```sh
./nixai config --help
Manage nixai configuration settings.

Usage:
  nixai config [get|set|edit] [key] [value]

Available Commands:
  get     View current configuration
  set     Set a configuration value
  edit    Edit the configuration file in your editor

Flags:
  -h, --help   help for config

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai config get
  nixai config set ai.provider ollama
  nixai config edit
```

---

## Real Life Examples

- **Switch AI provider to Gemini:**
  ```sh
  nixai config set ai.provider gemini
  # Changes the default AI provider to Gemini
  ```
- **Edit the configuration file in your editor:**
  ```sh
  nixai config edit
  # Opens the YAML config in your default editor
  ```
- **View all current configuration values:**

  ```bash
  nixai config get
  # Prints all current config settings
  ```

---

## Configuration Troubleshooting

### AI Provider Configuration Issues

If you encounter errors like:

```text
❌ Failed to initialize AI provider: provider 'ollama' is not configured
```

This indicates your configuration file has empty or missing AI provider definitions. This can happen after:

- Initial installation
- Configuration reset
- Manual editing
- Version upgrades

**Quick Fix:**

```bash
nixai config reset
```

**Manual Check:**

```bash
cat ~/.config/nixai/config.yaml | grep -A 5 "providers:"
```

If you see `providers: {}` (empty), the configuration needs to be reset.

**For detailed troubleshooting:** See [AI Provider Configuration Troubleshooting Guide](TROUBLESHOOTING_AI_PROVIDER_CONFIGURATION.md)

### Configuration File Location

- **User Config**: `~/.config/nixai/config.yaml`
- **System Config**: `/etc/nixai/config.yaml` (NixOS module)
- **Embedded Default**: Built into the binary

The system automatically falls back through these locations to ensure configuration is always available.
