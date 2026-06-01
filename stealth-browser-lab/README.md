# stealth-browser-lab 🥷 (Python / async)

**Playwright (async) によるステルス・ブラウザ自動化の学習＆練習ラボ。**
指紋(fingerprint)偽装、人間らしいマウス軌跡（WindMouse）、人間らしいタイピングを
小さなモジュールに分けて、「なぜそれが効くのか」を理解しながら触れるようにしています。

> ⚠️ **免責 / 利用上の注意**
> 本リポジトリは **学習・研究・自分が権限を持つ対象（自社サイトのQA、許可された公開データ取得）** を前提とした教材です。
> ボット検出は不正利用を防ぐために存在します。**利用規約・robots.txt・レート制限を必ず尊重**し、
> 認証突破・スクレイピング規約違反・不正アクセス等には使用しないでください。法的責任は利用者にあります。

---

## 📚 まず全体像（レイヤーで理解する）

```
┌ AIエージェント層     : agent-browser / Codex （AIが操作を判断）
├ 自動化ライブラリ層   : Playwright / Puppeteer  ← このラボはここ（Python API）
├ 通信プロトコル層     : CDP (Chrome DevTools Protocol)
└ ブラウザ本体         : Chromium (headless / headed)
```

- **Playwright** は「本物のブラウザを運転する道具」。Python 公式 API がある。
- デフォルトの自動化は **検出される**（`navigator.webdriver` などの足跡）。
- そこで「足跡を消す」「指紋を整える」「人間らしく動く」の3点を足すのがこのラボ。

### 🐍 sync と async の違い（よくある誤解）
- これは **プロセス数の話ではありません**。Playwright API の“書き方のスタイル”の違い。
- **sync**: `await` 不要で逐次実行。読みやすく学習向き。
- **async**: `async`/`await` + `asyncio`。複数ページ/サイトを **1スレッドで並行** に捌きやすい。
- どちらも基本シングルスレッド。`multiprocessing`（複数プロセス並列）とは別物。
- **このラボは async 版**。多サイトを同時に回したい時に強い。

---

## 🗂 ディレクトリ構成

```
stealth-browser-lab/
├── stealth_lab/
│   ├── utils.py            # 乱数・正規分布・非同期sleep
│   ├── human_mouse.py      # WindMouse による人間らしいマウス軌跡
│   ├── human_type.py       # 打鍵間隔のばらつき＋タイプミス訂正
│   ├── fingerprint.py      # Canvas/WebGL/Audio/navigator の指紋偽装
│   └── stealth_browser.py  # 上記を束ねてステルス構成で起動 (async)
├── examples/
│   └── detect_test.py      # sannysoft / creepjs を巡回して結果保存
├── main.py                 # デモのエントリポイント
├── requirements.txt
└── README.md
```

> 💡 **指紋偽装(JS)は言語非依存**：`fingerprint.py` がブラウザに注入するのは“ブラウザ内で動く JS”。
> Python からは `context.add_init_script(js_string)` に渡すだけ。JS版と中身は同じです。

---

## ⚙️ セットアップ & 実行

```bash
python -m venv .venv && source .venv/bin/activate   # 任意
pip install -r requirements.txt
playwright install chromium                         # ブラウザ本体を取得

python main.py                  # bot.sannysoft.com で効果を確認（results/sannysoft.png）
python -m examples.detect_test  # sannysoft + creepjs を巡回
```

`results/` のスクショで、緑（=検出回避OK）が多いほどステルスが効いています。

---

## 🧠 各テクニックの要点

### 1. 指紋偽装 (`fingerprint.py`)
`context.add_init_script` で **ページの全スクリプトより前** に JS を注入。
- `navigator.webdriver` を `false`
- `Canvas.toDataURL` に **seed固定の極小ノイズ**
- `WebGL` の VENDOR/RENDERER を実在GPU風に
- `AudioContext` の波形に極小ノイズ
- `permissions.query` の矛盾を解消
👉 **最重要は「整合性」**。UA=Windows なら `platform=Win32`、`locale=ja-JP` なら `timezone=Asia/Tokyo`。

### 2. 人間らしいマウス (`human_mouse.py`)
**WindMouse**：目標への「重力」＋ランダムな「風」を速度に加え続け、自然な弧・揺れ・終端の減速を再現。

### 3. 人間らしいタイピング (`human_type.py`)
打鍵間隔を正規分布でばらつかせ、句読点で長めに止まり、たまに隣キーをミスして Backspace 訂正。

### 4. 起動オプション (`stealth_browser.py`)
- `--disable-blink-features=AutomationControlled`、`ignore_default_args=['--enable-automation']`
- `launch_persistent_context` で **永続プロファイル**（Cookie/履歴の“生活感”）
- `playwright-stealth` が入っていれば追加適用（無くても自前指紋で動く）

---

## 🧭 学習ロードマップ（どう勉強するか）

**Step 0｜「検出される側」を見る**
対策なしの素の Playwright で `bot.sannysoft.com` を開き、赤い項目を確認 →「何が検出されるか」を体感。

**Step 1｜足跡を1つずつ潰す**
`fingerprint.py` の `_FP_JS` のブロックを**1つだけ有効化**しては再テスト。
`navigator.webdriver` → Canvas → WebGL… と緑に変わる様子を観察（原理を1つずつ理解）。

**Step 2｜挙動の自然さ**
`wind_mouse()` の `G`/`W`/`max_step` を変えて軌跡を観察。点列を matplotlib で描くと違いが一目瞭然。

**Step 3｜一貫性チェック**
`creepjs` の **trust score** を見る。「強く隠す」より「矛盾なく一貫」が効く、を確認。

**Step 4｜上級トピック（必要になったら）**
- CDP 痕跡の隠蔽：`undetected-playwright` / `rebrowser-patches`
- TLS/JA3 指紋（ブラウザ利用なら基本問題にならない理由）
- reCAPTCHA v3 / Cloudflare Turnstile が「挙動スコア」で年々強くなる現実

### 推奨の学び方
- **手を動かす → 再テスト → 差分を観察** のループ
- 1コミット＝1テクニックで小さく試す
- 「なぜ検出されるか（攻撃者視点）」「なぜ防ぐか（防御者視点）」の両面で考える

---

## 🔗 参考リンク

- Playwright (Python): https://playwright.dev/python/
- playwright-stealth (PyPI): https://pypi.org/project/playwright-stealth/
- 検出テスト（sannysoft）: https://bot.sannysoft.com/
- 指紋一貫性（CreepJS）: https://abrahamjuliot.github.io/creepjs/
- WindMouse 解説（BenLand100）: https://ben.land/post/2021/04/25/windmouse-human-mouse-movement/

---

## ⚖️ ライセンス
MIT（学習目的）。利用は各サイトの規約・法令の範囲内で。
