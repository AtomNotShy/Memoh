# Memoh CLI

The Memoh CLI (`memoh`) is a command-line tool for managing bots, channels, providers, models, schedules, and chatting with bots. It talks to a running Memoh server via its API.

## Installation

The CLI is part of the Memoh monorepo. Install from source:

```bash
git clone https://github.com/memohai/Memoh.git
cd Memoh
pnpm install
```

Run the CLI:

```bash
cd packages/cli
pnpm start -- --help
```

To use `memoh` as a global command:

```bash
cd packages/cli
pnpm build
pnpm link --global
memoh --help
```

Ensure your Memoh server is running (see [Docker installation](/installation/docker)) and the API is reachable at the configured host/port (default: `127.0.0.1:8080`).

## Configuration

The CLI stores config in `~/.memoh/config.toml` and auth token in `~/.memoh/token.json`. Use `memoh config` to view and `memoh config set` to change host/port.

## Commands

| Command | Description |
|---------|-------------|
| [login](./auth#login) | Log in to the Memoh server |
| [logout](./auth#logout) | Log out and clear token |
| [whoami](./auth#whoami) | Show current user |
| [config](./config) | Show or update CLI config (host, port) |
| [provider](./provider) | List, create, delete LLM providers |
| [model](./model) | List, create, delete models |
| [bot](./bot) | List, create, update, delete bots; chat; set model |
| [channel](./channel) | List channels; get/set bot channel config; get/set user binding |
| [schedule](./schedule) | List, create, update, toggle, delete bot schedules |
| [chat](./chat) | Interactive chat with a bot (default command) |
| [tui](./chat#tui) | Terminal UI chat session |
| [version](./chat#version) | Show CLI version |

Most commands require authentication. Run `memoh login` first.
