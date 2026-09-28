[English](../../README.md) · [简体中文](README.zh-CN.md) · **Русский** · [हिन्दी](README.hi.md) · [العربية](README.ar.md)

<p align="center">
  <img src="assets/social-preview.png" alt="mcp-assert" width="600">
</p>

<p align="center">
  <a href="https://github.com/blackwell-systems"><img src="https://raw.githubusercontent.com/blackwell-systems/blackwell-docs-theme/main/badge-trademark.svg" alt="Blackwell Systems"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/go-1.23+-blue.svg" alt="Go"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-green.svg" alt="License: MIT"></a>
  <a href="https://github.com/blackwell-systems/mcp-assert"><img src="https://raw.githubusercontent.com/blackwell-systems/mcp-assert/main/assets/badge-passing.svg?v=3" alt="mcp-assert: passing" height="20"></a>
  <a href="https://github.com/blackwell-systems/mcp-assert"><img src="https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/blackwell-systems/mcp-assert/main/assets/downloads-badge.json" alt="Downloads"></a>
</p>

**Тестируйте ваш MCP-сервер на реальном протоколе. Без моков. Без импортов. Без привязки к языку.**

mcp-assert подключается к вашему серверу точно так же, как это делают Claude, Cursor или любой другой MCP-клиент: реальный транспорт stdio/SSE/HTTP, полное рукопожатие initialize, настоящие вызовы инструментов. Он сверяет ответы с ожиданиями, которые вы задаёте в YAML. Если сервер проходит mcp-assert, он работает с любым MCP-клиентом.

> [!WARNING]
> Мы просканировали 102 MCP-сервера и обнаружили **4 794 проблемы со схемами** (2 239 ошибок) в 55 серверах, включая AWS, Serena и Grafana. Самая частая ошибка: у параметров отсутствуют определения типов, из-за чего агенты отправляют значения неверных типов. См. [оценочную таблицу](https://blackwell-systems.github.io/mcp-assert/scorecard/).

```
Your YAML        ──→  mcp-assert  ──→  MCP Server
(inputs + assertions)    (client)        (any language)
                            │
                        Pass / Fail
```

### Ваш сервер не заметит разницы

mcp-assert говорит на полном протоколе MCP: рукопожатие initialize, обнаружение через `tools/list`, `tools/call` с реальными аргументами. Он находит баги, которые пропускают юнит-тесты, потому что тестирует по проводу, а не внутри процесса.

### Применяется в продакшене

- **[Wyre Technology](https://github.com/wyre-technology)**: 25 MCP-серверов протестированы через общий baseline-процесс с использованием `mcp-assert-action`
- **[Ant Group (AntV)](https://github.com/antvis/mcp-server-chart)**: интегрирован в CI в течение 3 дней после запуска
- **[Vera](https://github.com/aallan/vera)**: рекомендованный тестовый харнесс в дорожной карте проекта ([#529](https://github.com/aallan/vera/issues/529))
- **Смерженные фикс-PR**: Google, Grafana, LangChain, официальные MCP SDK

Стандарт тестирования для MCP, как pytest для Python или Jest для JavaScript.

Добавьте его в любой проект MCP-сервера одной строкой:

```yaml
- uses: blackwell-systems/mcp-assert-action@v1
  with:
    suite: evals/
```

<p align="center">
  <img src="assets/demo.gif" alt="mcp-assert demo" width="720">
</p>

> [!NOTE]
> LLM — для субъективных результатов. Утверждения — для детерминированных. Большинство MCP-инструментов детерминированы. mcp-assert их покрывает.

## Установка

```bash
# npm (no Go required)
npx @blackwell-systems/mcp-assert

# pip (no Go required)
pip install mcp-assert

# Go
go install github.com/blackwell-systems/mcp-assert/cmd/mcp-assert@latest

# Homebrew
brew install blackwell-systems/tap/mcp-assert

# Docker
docker run blackwellsystems/mcp-assert audit --server "npx my-server"

# Snap (Linux)
sudo snap install mcp-assert --classic

# Scoop (Windows)
scoop bucket add blackwell-systems https://github.com/blackwell-systems/scoop-bucket
scoop install mcp-assert

# Winget (Windows)
winget install BlackwellSystems.mcp-assert

# curl | sh (macOS / Linux)
curl -fsSL https://raw.githubusercontent.com/blackwell-systems/mcp-assert/main/install.sh | sh
```

## Быстрый старт

### Проверьте любой MCP-сервер за секунды. Без настройки.

Направьте его на любой сервер:

```bash
mcp-assert audit --server "npx my-mcp-server"
```

```
  Server: my-server
  Transport: stdio
  Score: 83%

  ✓ read_query      1ms  [E000] responds, returns content
  ✗ create_table    0ms  [E201] internal error: panic: nil pointer...
  ✓ list_tables     1ms  [E000] responds, returns content

  3 tools tested, 2 healthy, 1 crashed
```

Структурированные коды ошибок мгновенно классифицируют проблемы. Все 24 кода см. в [справочнике по ошибкам](docs/ERROR_REFERENCE.md).

> [!TIP]
> Аудит подключается, обнаруживает каждый инструмент через `tools/list`, вызывает каждый из них со входными данными, сгенерированными по схеме, и сообщает, какие инструменты падают, а какие корректно обрабатывают ошибки. YAML не требуется. Чтобы копнуть глубже, сгенерируйте файлы утверждений и настройте их:


```bash
# Audit + generate starter YAML for CI
mcp-assert audit --server "npx my-mcp-server" --output evals/

# Edit the generated YAMLs: add expected content, setup steps, multi-step flows

# Run in CI with regression detection
mcp-assert ci --suite evals/ --threshold 95
```

### Пишите утверждения с нуля

```bash
# Scaffold your first assertion
mcp-assert init evals                   # Or: init evals --server "my-server" for auto-generation

# Run it
mcp-assert run --suite evals/ --fixture evals/fixtures
```

Полное пошаговое руководство см. в [руководстве по началу работы](https://blackwell-systems.github.io/mcp-assert/getting-started/).

### Уже используете Vitest, Jest, Bun, PHPUnit или pytest?

```bash
# Vitest
npm install -D @blackwell-systems/vitest-mcp-assert
```

```ts
// mcp.test.ts
import { describeMcpSuite } from '@blackwell-systems/vitest-mcp-assert'
describeMcpSuite('mcp server', 'evals/')
```

```bash
# pytest
pip install pytest-mcp-assert
pytest --mcp-suite evals/
```

> [!IMPORTANT]
> Одни и те же YAML-файлы работают в CLI, Vitest, Jest, Bun, PHPUnit, pytest и Go test. Миграция не нужна. Напишите один раз, запускайте где угодно.

## Всё, что вы можете сделать

| Команда | Что делает | Требуется настройка |
|---------|-------------|----------------|
| `audit --server "..."` | Сканирует любой сервер, относит каждый инструмент к категории здоров/упал/истёк тайм-аут | Нет |
| `fuzz --server "..."` | Бросает состязательные входные данные в каждый инструмент, находит падения и зависания | Нет |
| `init --server "..."` | Генерирует полный набор тестов из tools/list + захватывает снимки | Нет |
| `run --suite evals/` | Запускает YAML-утверждения, сообщает pass/fail | YAML-файлы |
| `ci --suite evals/` | Запуск с порогами, baseline, JUnit XML, GitHub Step Summary | YAML-файлы |
| `coverage --suite evals/ --server "..."` | Сообщает, у каких инструментов есть утверждения, а у каких нет | YAML-файлы |
| `snapshot --suite evals/ --update` | Захватывает ответы как golden-файлы для обнаружения регрессий | YAML-файлы |
| `watch --suite evals/` | Перезапускает при изменениях YAML, показывает диффы при смене статуса | YAML-файлы |
| `matrix --languages go:gopls,ts:tsserver` | Один и тот же набор на нескольких языковых серверах | YAML-файлы |
| `intercept --server "..." --trajectory t.yaml` | Проксирует между агентом и сервером, захватывает живую трассу вызовов инструментов | Trajectory YAML |
| `lint --server "..."` | 24 правила статического анализа для удобства работы агента; `--fix` автоматически генерирует улучшения схемы | Нет |

Начните с `audit` (нулевая настройка), затем `fuzz` (состязательное тестирование), затем `init` (генерирует всё), затем настройте YAML под ваши конкретные утверждения.

## Покрытие без усилий

```bash
# Generate stub assertions for every tool the server exposes
mcp-assert generate --server "my-mcp-server" --output evals/ --fixture ./fixtures

# Capture actual outputs as snapshots
mcp-assert snapshot --suite evals/ --server "my-mcp-server" --update

# Assert nothing changed
mcp-assert run --suite evals/ --server "my-mcp-server"
```

## Lint + автоисправление

Статический анализ выявляет проблемы со схемами, не выполняя инструменты. 24 правила обнаруживают проблемы, из-за которых агенты дают сбой:

```bash
mcp-assert lint --server "npx my-mcp-server"
```

```
  E  E103   create_entities       Required parameter "entities" has no description
  W  W114   generate_chart        Input schema is 5 levels deep. LLMs struggle with nesting
  W  W112   (server)              Server exposes 27 tools. LLM accuracy degrades beyond 20

5 error(s), 11 warning(s)
```

Автоматически сгенерировать исправления:

```bash
mcp-assert lint --server "npx my-mcp-server" --fix
```

```
memory-server: 9 tools, 25 findings, 23 auto-fixable

  E103   create_entities   Add description: "The entities value (array)"
  W109   search_nodes      Add examples to "query": [search term]
  W116   read_graph        Append: "Returns the graph data as JSON."

23 fixes generated.
```

Используйте `--strict` в CI, чтобы падать на предупреждениях:

```bash
mcp-assert lint --server "..." --strict --threshold 0
```

## Чем отличается от фреймворков «LLM-как-судья»

Для детерминированных инструментов mcp-assert подходит лучше. Для субъективных результатов фреймворки «LLM-как-судья» остаются правильным выбором. Используйте оба, если ваш сервер сочетает разные типы инструментов.

| Измерение | Eval-фреймворки «LLM-как-судья» | mcp-assert |
|---|---|---|
| Лучше всего для | Субъективные результаты (проза, творческий контент) | Детерминированные результаты (данные, состояние, валидация) |
| Оценивание | Оценка языковой моделью (гибко, дорого) | На основе утверждений (точно, бесплатно) |
| Скорость | Секунды на тест (round-trip к LLM) | Миллисекунды на тест (без LLM) |
| Стоимость CI | Вызовы API при каждом запуске | Ноль внешних зависимостей |
| Надёжность | Не измеряется | pass@k / pass^k на каждое утверждение |
| Регрессии | Не поддерживается | Сравнение с baseline, падение при откате |
| Многоязычность | Не поддерживается | Одно утверждение на N языковых серверов |

## Почему бы просто не написать тесты?

Вам понадобятся бутстрап протокола MCP, независимый от сервера раннер (ваши Go-тесты не смогут протестировать ваш TypeScript-сервер) и eval-функции (обнаружение регрессий, изоляция в Docker, вывод JUnit). mcp-assert берёт всё это на себя. Один YAML-файл, любой сервер, любой язык.

## Интеграция с CI

Используйте [GitHub Action mcp-assert](https://github.com/blackwell-systems/mcp-assert-action) для CI без настройки:

```yaml
- uses: blackwell-systems/mcp-assert-action@v1
  with:
    suite: evals/
    threshold: 95
```

Загружает бинарник, запускает утверждения, выгружает результаты JUnit XML, пишет GitHub Step Summary. Тулчейн Go на ваших раннерах не требуется.

Или запустите напрямую:

```bash
mcp-assert ci --suite evals/ --threshold 95 --junit results.xml
```

Про JUnit XML, markdown-сводки, бейджи и обнаружение регрессий см. в [руководстве по интеграции с CI](https://blackwell-systems.github.io/mcp-assert/ci-integration/).

## Интеграция с pytest

Запускайте утверждения mcp-assert как тестовые элементы pytest:

```bash
pip install pytest-mcp-assert
pytest --mcp-suite evals/
```

Каждый YAML-файл становится элементом pytest с семантикой pass/fail/skip. Настройка через `pyproject.toml`:

```toml
[tool.pytest.ini_options]
mcp_suite = "evals/"
mcp_fixture = "fixtures/"
```

Затем просто запустите `pytest`. Все опции см. в `pytest-plugin/README.md`.

## Интеграция с Vitest

Запускайте утверждения mcp-assert как тесты Vitest:

```bash
npm install -D @blackwell-systems/vitest-mcp-assert
```

Автоматически обнаруживать все YAML-файлы в каталоге:

```ts
// mcp.test.ts
import { describeMcpSuite } from '@blackwell-systems/vitest-mcp-assert'
describeMcpSuite('mcp server', 'evals/')
```

Или запускать отдельные утверждения:

```ts
import { test } from 'vitest'
import { runMcpAssert } from '@blackwell-systems/vitest-mcp-assert'
test('echo tool', () => runMcpAssert('evals/echo.yaml'))
```

Одни и те же YAML-файлы работают в Vitest, pytest и CLI. Все опции см. в `vitest-plugin/README.md`.

## Документация

Полная документация доступна на [blackwell-systems.github.io/mcp-assert](https://blackwell-systems.github.io/mcp-assert):

- [Начало работы](https://blackwell-systems.github.io/mcp-assert/getting-started/): установка, генерация каркаса, первый запуск
- [Написание утверждений](https://blackwell-systems.github.io/mcp-assert/writing-assertions/): формат YAML, все 18 типов утверждений + 4 типа траекторий, 8 типов блоков, 6 плагинов для тестовых фреймворков (pytest, Vitest, Jest, Bun, PHPUnit, Go test), setup-шаги, capture, фикстуры
- [Справочник по CLI](https://blackwell-systems.github.io/mcp-assert/cli/): полный справочник команд с флагами и примерами
- [Примеры](https://blackwell-systems.github.io/mcp-assert/examples/): 65 примеров наборов на 8 языках (606 утверждений)
- [Интеграция с CI](https://blackwell-systems.github.io/mcp-assert/ci-integration/): GitHub Action, JUnit XML, обнаружение регрессий
- [Бейдж](https://blackwell-systems.github.io/mcp-assert/badge/): добавьте бейдж «Works with mcp-assert» в README вашего сервера
- [Архитектура](https://blackwell-systems.github.io/mcp-assert/architecture/): внутреннее устройство и проектные решения
- [Дорожная карта](https://blackwell-systems.github.io/mcp-assert/roadmap/): что уже выпущено и что дальше
- [Оценочная таблица](https://blackwell-systems.github.io/mcp-assert/scorecard/): 32 бага найдены в 13 серверах, 9 фикс-PR отправлены, 58 серверов просканированы

<p align="center">
  <img src="assets/download-stats.svg?v=2" alt="Download stats" width="320">
</p>

<p align="center">
  <a href="https://github.com/blackwell-systems/mcp-assert">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="assets/star-cta.png">
      <source media="(prefers-color-scheme: light)" srcset="assets/star-cta-light.png">
      <img src="assets/star-cta-light.png" alt="Star mcp-assert on GitHub" width="600">
    </picture>
  </a>
</p>

## Лицензия

MIT
