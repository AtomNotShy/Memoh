# Schedule Commands

Manage cron-based schedules for bots. All schedule commands require `--bot <id>` to specify the bot.

## schedule list

List all schedules for a bot.

```bash
memoh schedule list --bot <bot_id>
```

## schedule get

Get a schedule by ID.

```bash
memoh schedule get <id> --bot <bot_id>
```

## schedule create

Create a new schedule. Prompts for name, description, cron pattern, command, and optional max calls.

```bash
memoh schedule create [options] --bot <bot_id>
```

| Option | Description |
|--------|-------------|
| `--name <name>` | Schedule name |
| `--description <desc>` | Description |
| `--pattern <pattern>` | Cron pattern (e.g. `0 9 * * *` for daily at 9am) |
| `--command <cmd>` | Command to run in the bot container |
| `--max_calls <n>` | Max executions (optional, empty for unlimited) |
| `--enabled` | Create as enabled |
| `--disabled` | Create as disabled |

## schedule update

Update a schedule.

```bash
memoh schedule update <id> [options] --bot <bot_id>
```

| Option | Description |
|--------|-------------|
| `--name <name>` | Schedule name |
| `--description <desc>` | Description |
| `--pattern <pattern>` | Cron pattern |
| `--command <cmd>` | Command |
| `--max_calls <n>` | Max executions |
| `--enabled` | Enable |
| `--disabled` | Disable |

## schedule toggle

Enable or disable a schedule (flip current state).

```bash
memoh schedule toggle <id> --bot <bot_id>
```

## schedule delete

Delete a schedule.

```bash
memoh schedule delete <id> --bot <bot_id>
```
