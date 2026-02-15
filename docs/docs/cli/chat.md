# Chat Commands

## Default: Interactive Chat

Running `memoh` with no subcommand starts an interactive chat. Use `--bot <id>` to specify which bot to chat with; otherwise you'll be prompted to select one.

```bash
memoh [options]
memoh --bot <bot_id>
```

| Option | Description |
|--------|-------------|
| `--bot <id>` | Bot ID to chat with |

Type your message and press Enter. Type `exit` to quit. Responses stream in real time.

## tui

Terminal UI chat session. Same behavior as the default chat but explicitly invoked.

```bash
memoh tui [options]
memoh tui --bot <bot_id>
```

| Option | Description |
|--------|-------------|
| `--bot <id>` | Bot ID to chat with |

## version

Show the CLI version.

```bash
memoh version
```
