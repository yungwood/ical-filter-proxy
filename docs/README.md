# Documentation Site

This website is built using [Docusaurus](https://docusaurus.io/).

The root `README.md` should stay as the concise project overview and quickstart.
Use this Docusaurus site for expanded documentation, examples, and reference
material that would make the root README too long.

## Installation

```bash
npm ci
```

## Local Development

```bash
npm run start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

## Build

```bash
npm run build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

## Publishing

Docs changes are validated by CI on pushes and pull requests, but the public
GitHub Pages site is deployed only from release tags. The published docs should
represent the latest tagged release.

## Typecheck

```bash
npm run typecheck
```
