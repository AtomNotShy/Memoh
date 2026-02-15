# Auth Commands

## login

Log in to the Memoh server. Prompts for username and password, then stores the JWT token in `~/.memoh/token.json`.

```bash
memoh login
```

Interactive prompts:
- Username
- Password

## logout

Clear the stored token and log out.

```bash
memoh logout
```

## whoami

Show the current logged-in user (username, display name, user ID, role). Falls back to token info if the API call fails.

```bash
memoh whoami
```
