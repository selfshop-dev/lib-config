# Contributing

Проект следует [`Conventional Commits`](https://www.conventionalcommits.org). Формат: `<type>: <summary>`, где `summary` — повелительное наклонение,английский язык, без точки в конце.

| Тип | Когда использовать |
|---|---|
| `feat` | Новая функциональность |
| `fix` | Исправление бага |
| `refactor` | Рефакторинг без изменения поведения |
| `perf` | Оптимизация производительности |
| `test` | Добавление или изменение тестов |
| `docs` | Документация |
| `ci` | Изменения CI/CD и GitHub Actions |
| `build` | Сборка, зависимости, инфраструктура проекта |
| `chore` | Рутинные задачи, не попадающие в другие типы |

Тип коммита автоматически определяет метку PR и тип следующего релиза:  

- `feat!` / `fix!` (breaking change) → major
- `feat` → minor
- `fix` / `perf` / `refactor` → patch

## Требования к коду

Подробно описаны в [`PULL_REQUEST_TEMPLATE`](https://github.com/selfshop-dev/.github/blob/main/PULL_REQUEST_TEMPLATE.md?plain=1) (раздел Checklist).

## Issues и уязвимости

- Общие вопросы и баги — через [`шаблоны issues`](../../issues/new/choose).
- **Уязвимости** — **только** через [`🔒 Private Vulnerability Reporting`](../../security/advisories/new). Не создавай публичный issue или discussion!