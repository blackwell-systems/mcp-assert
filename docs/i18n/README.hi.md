[English](../../README.md) · [简体中文](README.zh-CN.md) · [Русский](README.ru.md) · **हिन्दी** · [العربية](README.ar.md)

<p align="center">
  <img src="../../assets/social-preview.png" alt="mcp-assert" width="600">
</p>

<p align="center">
  <a href="https://github.com/blackwell-systems"><img src="https://raw.githubusercontent.com/blackwell-systems/blackwell-docs-theme/main/badge-trademark.svg" alt="Blackwell Systems"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/go-1.23+-blue.svg" alt="Go"></a>
  <a href="../../LICENSE"><img src="https://img.shields.io/badge/license-MIT-green.svg" alt="License: MIT"></a>
  <a href="https://github.com/blackwell-systems/mcp-assert"><img src="https://raw.githubusercontent.com/blackwell-systems/mcp-assert/main/assets/badge-passing.svg?v=3" alt="mcp-assert: passing" height="20"></a>
  <a href="https://github.com/blackwell-systems/mcp-assert"><img src="https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/blackwell-systems/mcp-assert/main/assets/downloads-badge.json" alt="Downloads"></a>
</p>

**अपने MCP सर्वर को असली प्रोटोकॉल के विरुद्ध टेस्ट करें। कोई mock नहीं। कोई import नहीं। किसी भाषा से बंधन नहीं।**

mcp-assert आपके सर्वर से ठीक वैसे ही कनेक्ट होता है जैसे Claude, Cursor या कोई भी MCP क्लाइंट करता है: असली stdio/SSE/HTTP ट्रांसपोर्ट, पूरा initialize handshake, वास्तविक टूल कॉल। यह प्रतिक्रियाओं को उन अपेक्षाओं के विरुद्ध जाँचता है जिन्हें आप YAML में परिभाषित करते हैं। अगर यह mcp-assert पास कर लेता है, तो यह हर MCP क्लाइंट के साथ काम करता है।

> [!WARNING]
> हमने 102 MCP सर्वर स्कैन किए और AWS, Serena और Grafana समेत 55 सर्वरों में **4,794 schema समस्याएँ** (2,239 errors) पाईं। सबसे आम विफलता: पैरामीटरों में टाइप परिभाषाएँ न होने से agent गलत मान प्रकार भेजते हैं। देखें [scorecard](https://blackwell-systems.github.io/mcp-assert/scorecard/)।

```
Your YAML        ──→  mcp-assert  ──→  MCP Server
(inputs + assertions)    (client)        (any language)
                            │
                        Pass / Fail
```

### आपका सर्वर फ़र्क नहीं बता पाएगा

mcp-assert पूरा MCP प्रोटोकॉल बोलता है: initialize handshake, `tools/list` से खोज, असली आर्ग्युमेंट्स के साथ `tools/call`। यह ऐसे bug खोजता है जो यूनिट टेस्ट छोड़ देते हैं, क्योंकि यह प्रोसेस के भीतर नहीं, बल्कि वायर पर टेस्ट करता है।

### उत्पादन में अपनाया गया

- **[Wyre Technology](https://github.com/wyre-technology)**: `mcp-assert-action` का उपयोग करते हुए साझा baseline वर्कफ़्लो के ज़रिए 25 MCP सर्वर टेस्ट किए गए
- **[Ant Group (AntV)](https://github.com/antvis/mcp-server-chart)**: लॉन्च के 3 दिनों के भीतर CI में एकीकृत
- **[Vera](https://github.com/aallan/vera)**: प्रोजेक्ट रोडमैप पर अनुशंसित टेस्ट harness ([#529](https://github.com/aallan/vera/issues/529))
- **मर्ज किए गए Fix PR**: Google, Grafana, LangChain, आधिकारिक MCP SDK

MCP के लिए टेस्टिंग मानक, जैसे Python के लिए pytest या JavaScript के लिए Jest।

इसे किसी भी MCP सर्वर प्रोजेक्ट में एक ही लाइन में जोड़ें:

```yaml
- uses: blackwell-systems/mcp-assert-action@v1
  with:
    suite: evals/
```

<p align="center">
  <img src="../../assets/demo.gif" alt="mcp-assert demo" width="720">
</p>

> [!NOTE]
> LLM व्यक्तिपरक आउटपुट के लिए हैं। Assertions नियतात्मक आउटपुट के लिए हैं। अधिकांश MCP टूल नियतात्मक होते हैं। mcp-assert उन्हें कवर करता है।

## इंस्टॉल

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

## त्वरित शुरुआत

### किसी भी MCP सर्वर का सेकंडों में ऑडिट करें। कोई सेटअप नहीं।

इसे किसी भी सर्वर की ओर इंगित करें:

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

संरचित एरर कोड समस्याओं को तुरंत वर्गीकृत कर देते हैं। सभी 24 कोड के लिए देखें [Error Reference](../../docs/ERROR_REFERENCE.md)।

> [!TIP]
> ऑडिट कनेक्ट होता है, `tools/list` के ज़रिए हर टूल की खोज करता है, schema से बने इनपुट के साथ हर टूल को कॉल करता है, और रिपोर्ट करता है कि कौन-से टूल क्रैश होते हैं बनाम कौन-से एरर को सही तरह से संभालते हैं। कोई YAML की ज़रूरत नहीं। और गहराई में जाने के लिए, assertion फ़ाइलें जेनरेट करें और उन्हें कस्टमाइज़ करें:


```bash
# Audit + generate starter YAML for CI
mcp-assert audit --server "npx my-mcp-server" --output evals/

# Edit the generated YAMLs: add expected content, setup steps, multi-step flows

# Run in CI with regression detection
mcp-assert ci --suite evals/ --threshold 95
```

### शून्य से assertions लिखें

```bash
# Scaffold your first assertion
mcp-assert init evals                   # Or: init evals --server "my-server" for auto-generation

# Run it
mcp-assert run --suite evals/ --fixture evals/fixtures
```

पूरे walkthrough के लिए देखें [Getting Started guide](https://blackwell-systems.github.io/mcp-assert/getting-started/)।

### पहले से Vitest, Jest, Bun, PHPUnit या pytest का उपयोग कर रहे हैं?

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
> वही YAML फ़ाइलें CLI, Vitest, Jest, Bun, PHPUnit, pytest और Go test में एक जैसी काम करती हैं। कोई migration नहीं। एक बार लिखें, कहीं भी चलाएँ।

## आप जो कुछ भी कर सकते हैं

| कमांड | यह क्या करता है | ज़रूरी सेटअप |
|---------|-------------|----------------|
| `audit --server "..."` | किसी भी सर्वर को स्कैन करें, हर टूल को स्वस्थ/क्रैश/टाइम-आउट के रूप में वर्गीकृत करें | कोई नहीं |
| `fuzz --server "..."` | हर टूल पर विरोधात्मक इनपुट फेंकें, क्रैश और हैंग खोजें | कोई नहीं |
| `init --server "..."` | tools/list से एक पूरा टेस्ट सूट जेनरेट करें + स्नैपशॉट कैप्चर करें | कोई नहीं |
| `run --suite evals/` | YAML assertions चलाएँ, pass/fail रिपोर्ट करें | YAML फ़ाइलें |
| `ci --suite evals/` | thresholds, baselines, JUnit XML, GitHub Step Summary के साथ चलाएँ | YAML फ़ाइलें |
| `coverage --suite evals/ --server "..."` | रिपोर्ट करें कि किन टूल में assertions हैं और किनमें नहीं | YAML फ़ाइलें |
| `snapshot --suite evals/ --update` | रिग्रेशन डिटेक्शन के लिए प्रतिक्रियाओं को golden files के रूप में कैप्चर करें | YAML फ़ाइलें |
| `watch --suite evals/` | YAML बदलावों पर फिर से चलाएँ, स्थिति पलटने पर diff दिखाएँ | YAML फ़ाइलें |
| `matrix --languages go:gopls,ts:tsserver` | कई language servers पर एक ही सूट | YAML फ़ाइलें |
| `intercept --server "..." --trajectory t.yaml` | agent और सर्वर के बीच प्रॉक्सी करें, live टूल कॉल ट्रेस कैप्चर करें | Trajectory YAML |
| `lint --server "..."` | agent उपयोगिता के लिए 24 स्टैटिक विश्लेषण नियम; `--fix` स्वतः schema सुधार जेनरेट करता है | कोई नहीं |

`audit` (शून्य सेटअप) से शुरू करें, फिर `fuzz` (विरोधात्मक टेस्टिंग), फिर `init` (सब कुछ जेनरेट करता है), फिर अपने विशिष्ट assertions के लिए YAML को कस्टमाइज़ करें।

## शून्य-प्रयास कवरेज

```bash
# Generate stub assertions for every tool the server exposes
mcp-assert generate --server "my-mcp-server" --output evals/ --fixture ./fixtures

# Capture actual outputs as snapshots
mcp-assert snapshot --suite evals/ --server "my-mcp-server" --update

# Assert nothing changed
mcp-assert run --suite evals/ --server "my-mcp-server"
```

## Lint + ऑटो-फ़िक्स

स्टैटिक विश्लेषण टूल को चलाए बिना schema समस्याएँ पकड़ता है। 24 नियम उन समस्याओं का पता लगाते हैं जिनसे agent विफल होते हैं:

```bash
mcp-assert lint --server "npx my-mcp-server"
```

```
  E  E103   create_entities       Required parameter "entities" has no description
  W  W114   generate_chart        Input schema is 5 levels deep. LLMs struggle with nesting
  W  W112   (server)              Server exposes 27 tools. LLM accuracy degrades beyond 20

5 error(s), 11 warning(s)
```

फ़िक्स स्वतः जेनरेट करें:

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

चेतावनियों पर विफल होने के लिए CI में `--strict` का उपयोग करें:

```bash
mcp-assert lint --server "..." --strict --threshold 0
```

## यह LLM-as-Judge फ़्रेमवर्क से किस तरह भिन्न है

नियतात्मक टूल के लिए mcp-assert बेहतर उपयुक्त है। व्यक्तिपरक आउटपुट के लिए LLM-as-judge फ़्रेमवर्क सही विकल्प बने रहते हैं। अगर आपका सर्वर टूल प्रकारों को मिलाता है तो दोनों का उपयोग करें।

| आयाम | LLM-as-judge eval फ़्रेमवर्क | mcp-assert |
|---|---|---|
| किसके लिए सर्वोत्तम | व्यक्तिपरक आउटपुट (गद्य, रचनात्मक सामग्री) | नियतात्मक आउटपुट (डेटा, स्थिति, सत्यापन) |
| ग्रेडिंग | भाषा मॉडल स्कोरिंग (लचीली, महँगी) | assertion-आधारित (सटीक, मुफ़्त) |
| गति | प्रति टेस्ट सेकंड (LLM राउंड-ट्रिप) | प्रति टेस्ट मिलीसेकंड (कोई LLM नहीं) |
| CI लागत | हर रन पर API कॉल | शून्य बाहरी निर्भरताएँ |
| विश्वसनीयता | मापी नहीं गई | प्रति assertion pass@k / pass^k |
| रिग्रेशन | समर्थित नहीं | baseline तुलना, पीछे खिसकने पर विफल |
| बहु-भाषा | समर्थित नहीं | N language servers पर एक ही assertion |

## टेस्ट बस लिख क्यों न लें?

आपको MCP प्रोटोकॉल bootstrapping, एक सर्वर-अज्ञेय रनर (आपके Go टेस्ट आपके TypeScript सर्वर को टेस्ट नहीं कर सकते), और eval सुविधाएँ (रिग्रेशन डिटेक्शन, Docker आइसोलेशन, JUnit आउटपुट) चाहिए होंगी। mcp-assert यह सब संभालता है। एक YAML फ़ाइल, कोई भी सर्वर, कोई भी भाषा।

## CI एकीकरण

शून्य-सेटअप CI के लिए [mcp-assert GitHub Action](https://github.com/blackwell-systems/mcp-assert-action) का उपयोग करें:

```yaml
- uses: blackwell-systems/mcp-assert-action@v1
  with:
    suite: evals/
    threshold: 95
```

बाइनरी डाउनलोड करता है, assertions चलाता है, JUnit XML परिणाम अपलोड करता है, GitHub Step Summary लिखता है। आपके runners पर Go toolchain की ज़रूरत नहीं।

या सीधे चलाएँ:

```bash
mcp-assert ci --suite evals/ --threshold 95 --junit results.xml
```

JUnit XML, markdown सारांश, बैज और रिग्रेशन डिटेक्शन के लिए देखें [CI Integration guide](https://blackwell-systems.github.io/mcp-assert/ci-integration/)।

## pytest एकीकरण

mcp-assert assertions को pytest टेस्ट आइटम के रूप में चलाएँ:

```bash
pip install pytest-mcp-assert
pytest --mcp-suite evals/
```

हर YAML फ़ाइल pass/fail/skip सेमांटिक्स के साथ एक pytest Item बन जाती है। `pyproject.toml` के ज़रिए कॉन्फ़िगर करें:

```toml
[tool.pytest.ini_options]
mcp_suite = "evals/"
mcp_fixture = "fixtures/"
```

फिर बस `pytest` चलाएँ। सभी विकल्पों के लिए `pytest-plugin/README.md` देखें।

## Vitest एकीकरण

mcp-assert assertions को Vitest टेस्ट के रूप में चलाएँ:

```bash
npm install -D @blackwell-systems/vitest-mcp-assert
```

किसी डायरेक्टरी में सभी YAML फ़ाइलें स्वतः खोजें:

```ts
// mcp.test.ts
import { describeMcpSuite } from '@blackwell-systems/vitest-mcp-assert'
describeMcpSuite('mcp server', 'evals/')
```

या व्यक्तिगत assertions चलाएँ:

```ts
import { test } from 'vitest'
import { runMcpAssert } from '@blackwell-systems/vitest-mcp-assert'
test('echo tool', () => runMcpAssert('evals/echo.yaml'))
```

वही YAML फ़ाइलें Vitest, pytest और CLI में काम करती हैं। सभी विकल्पों के लिए `vitest-plugin/README.md` देखें।

## दस्तावेज़ीकरण

पूरा दस्तावेज़ीकरण [blackwell-systems.github.io/mcp-assert](https://blackwell-systems.github.io/mcp-assert) पर उपलब्ध है:

- [Getting Started](https://blackwell-systems.github.io/mcp-assert/getting-started/): इंस्टॉल, scaffold, पहला रन
- [Writing Assertions](https://blackwell-systems.github.io/mcp-assert/writing-assertions/): YAML प्रारूप, सभी 18 assertion प्रकार + 4 trajectory प्रकार, 8 block प्रकार, 6 टेस्ट फ़्रेमवर्क प्लगइन (pytest, Vitest, Jest, Bun, PHPUnit, Go test), setup चरण, capture, fixtures
- [CLI Reference](https://blackwell-systems.github.io/mcp-assert/cli/): फ़्लैग और उदाहरणों के साथ पूरा कमांड संदर्भ
- [Examples](https://blackwell-systems.github.io/mcp-assert/examples/): 8 भाषाओं में 65 उदाहरण सूट (606 assertions)
- [CI Integration](https://blackwell-systems.github.io/mcp-assert/ci-integration/): GitHub Action, JUnit XML, रिग्रेशन डिटेक्शन
- [Badge](https://blackwell-systems.github.io/mcp-assert/badge/): अपने सर्वर README में "Works with mcp-assert" बैज जोड़ें
- [Architecture](https://blackwell-systems.github.io/mcp-assert/architecture/): आंतरिक संरचना और डिज़ाइन निर्णय
- [Roadmap](https://blackwell-systems.github.io/mcp-assert/roadmap/): क्या शिप हो चुका है और आगे क्या है
- [Scorecard](https://blackwell-systems.github.io/mcp-assert/scorecard/): 13 सर्वरों में 32 bug पाए गए, 9 fix PR सबमिट किए गए, 58 सर्वर स्कैन किए गए

<p align="center">
  <img src="../../assets/download-stats.svg?v=2" alt="Download stats" width="320">
</p>

<p align="center">
  <a href="https://github.com/blackwell-systems/mcp-assert">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="assets/star-cta.png">
      <source media="(prefers-color-scheme: light)" srcset="assets/star-cta-light.png">
      <img src="../../assets/star-cta-light.png" alt="Star mcp-assert on GitHub" width="600">
    </picture>
  </a>
</p>

## लाइसेंस

MIT
