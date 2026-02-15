# Provider Commands

Manage LLM providers (OpenAI, Anthropic, Ollama, etc.).

## provider list

List all providers. Optionally filter by provider name.

```bash
memoh provider list [options]
```

| Option | Description |
|--------|-------------|
| `--provider <name>` | Filter by provider name |

Examples:

```bash
memoh provider list
memoh provider list --provider my-openai
```

## provider create

Create a new provider. Prompts for any missing fields.

```bash
memoh provider create [options]
```

| Option | Description |
|--------|-------------|
| `--name <name>` | Provider name |
| `--type <type>` | Client type |
| `--base_url <url>` | Base URL for the API |
| `--api_key <key>` | API key |

Supported client types: `openai`, `openai-compat`, `anthropic`, `google`, `azure`, `bedrock`, `mistral`, `xai`, `ollama`, `dashscope`

Examples:

```bash
memoh provider create --name my-ollama --type ollama --base_url http://localhost:11434
memoh provider create
# Interactive prompts
```

## provider delete

Delete a provider by name.

```bash
memoh provider delete --provider <name>
```

Example:

```bash
memoh provider delete --provider my-ollama
```
