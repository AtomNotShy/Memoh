# Config Commands

The CLI stores its config in `~/.memoh/config.toml`. Use these commands to view or update it.

## config

Show the current config (host and port).

```bash
memoh config
```

Output example:
```
host = "127.0.0.1"
port = 8080
```

## config set

Update the config. Prompts for host and port if not provided via options.

```bash
memoh config set [options]
```

| Option | Description |
|--------|-------------|
| `--host <host>` | API host (e.g. `127.0.0.1` or `api.example.com`) |
| `--port <port>` | API port (default: 8080) |

Examples:

```bash
memoh config set --host 192.168.1.100 --port 8080
memoh config set
# Interactive prompts for host and port
```
