# Model Commands

Manage chat and embedding models.

## model list

List all models with their provider, type, and multimodal flag.

```bash
memoh model list
```

## model create

Create a new model. Prompts for provider, model ID, type, and (for embedding models) dimensions.

```bash
memoh model create [options]
```

| Option | Description |
|--------|-------------|
| `--model_id <id>` | Model ID (e.g. `gpt-4`, `text-embedding-3-small`) |
| `--name <name>` | Display name |
| `--provider <provider>` | Provider name |
| `--type <type>` | `chat` or `embedding` |
| `--dimensions <n>` | Embedding dimensions (required for embedding models) |
| `--multimodal` | Mark as multimodal |

Examples:

```bash
memoh model create --model_id gpt-4 --provider my-openai --type chat
memoh model create --model_id text-embedding-3-small --provider my-openai --type embedding --dimensions 1536
memoh model create
# Interactive prompts
```

## model delete

Delete a model by model ID.

```bash
memoh model delete --model <model_id>
```

Example:

```bash
memoh model delete --model gpt-4
```
