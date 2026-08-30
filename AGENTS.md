# Agent Guidelines

## GitHub CLI Authentication

This repo is owned by the `masculinecache` GitHub account. The `gh` CLI must use the masculinecache config directory, not the default.

- **Interactive shells**: `mise.local.toml` sets `GH_CONFIG_DIR=~/.config/gh-masculinecache` automatically when mise is activated (see shell hook in `.zshrc`).
- **Non-interactive shells** (agents, scripts, CI): mise is not auto-activated. Prefix `gh` commands with the env var explicitly:

```sh
GH_CONFIG_DIR=~/.config/gh-masculinecache gh pr create ...
GH_CONFIG_DIR=~/.config/gh-masculinecache gh pr merge ...
```

Without this, `gh` defaults to `~/.config/gh` and operates under the wrong account.
