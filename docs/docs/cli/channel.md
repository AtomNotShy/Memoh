# Channel Commands

Manage channels and bot/user channel configuration.

## channel list

List available channel types (e.g. telegram, feishu, local).

```bash
memoh channel list
```

## channel info

Show channel metadata and schema for a channel type.

```bash
memoh channel info [type]
```

If `type` is omitted, prompts to select from available channels.

## channel config get

Get a bot's channel configuration.

```bash
memoh channel config get [bot_id] [options]
```

| Option | Description |
|--------|-------------|
| `--type <type>` | Channel type |

## channel config set

Set a bot's channel configuration. Currently supports Feishu.

```bash
memoh channel config set [bot_id] [options]
```

| Option | Description |
|--------|-------------|
| `--type <type>` | Channel type (e.g. `feishu`) |
| `--app_id <id>` | Feishu App ID |
| `--app_secret <secret>` | Feishu App Secret |
| `--encrypt_key <key>` | Encrypt key (optional) |
| `--verification_token <token>` | Verification token (optional) |

## channel bind get

Get the current user's channel binding for a platform.

```bash
memoh channel bind get [options]
```

| Option | Description |
|--------|-------------|
| `--type <type>` | Channel type |

## channel bind set

Set the current user's channel binding. Currently supports Feishu (open_id or user_id).

```bash
memoh channel bind set [options]
```

| Option | Description |
|--------|-------------|
| `--type <type>` | Channel type (e.g. `feishu`) |
| `--open_id <id>` | Feishu Open ID |
| `--user_id <id>` | Feishu User ID |
