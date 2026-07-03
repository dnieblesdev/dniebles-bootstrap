# dniebles-bootstrap

This repository is reserved for personal machine bootstrap work outside `~/.dotfiles`.

> Migrated from `~/.dotfiles/docs/bootstrapper-handoff.md`.

## Boundary

| Area | Owner |
|------|-------|
| Dotfiles modules, declarative profiles, and symlink lifecycle | `~/.dotfiles` |
| Packages, runtimes, Homebrew, OS setup, and machine provisioning | This bootstrap project |

## Notes from the dotfiles handoff

- The dotfiles repository no longer implements machine bootstrap.
- Removed dotfiles bootstrap code can be recovered from dotfiles git history if this project needs to migrate previous behavior.
- Do not reintroduce bootstrap authority into `~/.dotfiles`.
