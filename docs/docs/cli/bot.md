# Bot Commands

Manage bots and chat with them.

## bot list

List all bots. Admins can filter by owner.

```bash
memoh bot list [options]
```

| Option | Description |
|--------|-------------|
| `--owner <user_id>` | Filter by owner user ID (admin only) |

## bot create

Create a new bot. Prompts for type and optionally name.

```bash
memoh bot create [options]
```

| Option | Description |
|--------|-------------|
| `--type <type>` | `personal` or `public` |
| `--name <name>` | Display name |
| `--avatar <url>` | Avatar URL |
| `--active` | Set bot active |
| `--inactive` | Set bot inactive |

## bot update

Update bot info. Bot ID can be passed as argument or selected interactively.

```bash
memoh bot update [id] [options]
```

| Option | Description |
|--------|-------------|
| `--name <name>` | Display name |
| `--avatar <url>` | Avatar URL |
| `--active` | Set bot active |
| `--inactive` | Set bot inactive |

## bot delete

Delete a bot. Asks for confirmation.

```bash
memoh bot delete [id]
```

## bot chat

Start an interactive streaming chat with a bot.

```bash
memoh bot chat [id]
```

Type messages and press Enter. Type `exit` to quit.

## bot set-model

Enable a model for a bot (chat, memory, or embedding).

```bash
memoh bot set-model [id] [options]
```

| Option | Description |
|--------|-------------|
| `--as <usage>` | `chat`, `memory`, or `embedding` |
| `--model <model_id>` | Model ID |

Example:

```bash
memoh bot set-model my-bot-id --as chat --model gpt-4
```
