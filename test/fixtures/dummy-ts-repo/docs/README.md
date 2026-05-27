# Dummy TS Fixture

This directory holds a synthetic TypeScript project used only by GopherCram's
integration tests. It exists so we can pack a realistic-looking project and
compare GopherCram's output against npx-launched `repomix` runs.

## Layout

- `src/` — application source (TypeScript + React-flavoured TSX)
- `tests/` — Vitest unit tests
- `docs/` — markdown documentation (this file)
- `.env` — fake credentials; should be excluded by `.gitignore`
- `.env.example` — template version; safe to include in output
- `node_modules/` — fake dependency; should be excluded by default patterns

## Fake secrets

`src/config.ts` and `.env` contain clearly-fake placeholder secrets. They
exist so the security scanner has something to flag during integration tests
and grant no access to anything.
