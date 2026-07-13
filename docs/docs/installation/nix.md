---
sidebar_position: 4
---

# Nix

Build the application:

```bash
nix build
```

Run it from the local checkout:

```bash
nix run . -- --help
```

Enter a development shell with the Go toolchain:

```bash
nix develop
```

Run it directly from GitHub:

```bash
nix run github:yungwood/ical-filter-proxy -- --help
```

Run a tagged release:

```bash
nix run github:yungwood/ical-filter-proxy/0.3.0 -- --help
```

## Build Metadata

Nix builds embed the flake source revision in `ical-filter-proxy -version`
rather than the tag name.

When you run a tagged release, such as `0.3.0`, the tag still controls the
source being built. The binary reports the exact commit revision for
traceability.
