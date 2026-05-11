# Contributing

This project follows [`Conventional Commits`](https://www.conventionalcommits.org). Format: `<type>: <summary>`, where `summary` is imperative mood, in English, with no trailing period.

| Type | When to use |
|---|---|
| `feat` | New functionality |
| `fix` | Bug fix |
| `refactor` | Refactoring without behavior change |
| `perf` | Performance optimization |
| `test` | Adding or updating tests |
| `docs` | Documentation |
| `ci` | CI/CD and GitHub Actions changes |
| `build` | Build system, dependencies, project infrastructure |
| `chore` | Routine tasks that don't fit other types |

The commit type automatically determines the PR label and the next release type:
- `feat!` / `fix!` (breaking change) → major
- `feat` → minor
- `fix` / `perf` / `refactor` → patch

## Code Requirements
Described in detail in the [`PULL_REQUEST_TEMPLATE`](https://github.com/selfshop-dev/.github/blob/main/PULL_REQUEST_TEMPLATE.md?plain=1) (Checklist section).

## Issues and Vulnerabilities
- General questions and bugs — via [`issue templates`](../../issues/new/choose).
- **Vulnerabilities** — **only** via [`🔒 Private Vulnerability Reporting`](../../security/advisories/new). Do not create a public issue or discussion!