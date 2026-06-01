# stealth-browser-lab 🥷

**Playwright によるステルス・ブラウザ自動化の学習＆練習ラボ。**
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
├ 自動化ライブラリ層   : Playwright / Puppeteer  ← このラボはここ
├ 通信プロトコル層     : CDP (Chrome DevTools Protocol)
└ ブラウザ本体         : Chromium (headless / headed)
```

- **Playwright** は「本物のブラウザを運転する道具」。だから単純な検出は素通りできる。
- ただし **デフォルトは検出される**（`navigator.webdriver` などの足跡）。
- そこで「足跡を消す（ステルス）」「指紋を整える」「人間らしく動く」の3点を足すのがこのラボの主題。

---

## 🗂 ディレクトリ構成

```
stealth-browser-lab/
├── src/
│   ├── utils.js          # 乱数・正規分布・sleep（人間っぽさの土台）
│   ├── humanMouse.js     # WindMouse による人間らしいマウス軌跡
│   ├── humanType.js      # 打鍵間隔のばらつき＋タイプミス訂正
│   ├── fingerprint.js    # Canvas/WebGL/Audio/navigator の指紋偽装
│   ├── stealthBrowser.js # 上記を束ねてステルス構成で起動
│   └── index.js          # デモ（検出テストページのスクショ保存）
├── examples/
│   └── detect-test.js    # sannysoft / creepjs を巡回して結果保存
├── package.json
└── README.md
```

---

## ⚙️ セットアップ & 実行

```bash
npm install
npx playwright install chromium   # ブラウザ本体を取得

npm run demo          # bot.sannysoft.com で効果を確認（results/sannysoft.png）
npm run test:detect   # sannysoft + creepjs を巡回
```

`results/` に保存されたスクショで、緑（=検出回避OK）が多いほどステルスが効いています。

---

## 🧠 各テクニックの要点

### 1. 指紋偽装 (`fingerprint.js`)
`context.addInitScript` で **ページの全スクリプトより前** に注入する。
- `navigator.webdriver` を `false` に
- `Canvas` の `toDataURL` に**極小ノイズ**（seed固定で一貫性を保つ）
- `WebGL` の VENDOR/RENDERER を実在GPU風に
- `AudioContext` の波形に極小ノイズ
- `permissions.query` の矛盾を解消
👉 **最重要は「整合性」**。UA=Windows なら `platform=Win32`、`locale=ja-JP` なら `timezone=Asia/Tokyo`。矛盾は一発でバレる。

### 2. 人間らしいマウス (`humanMouse.js`)
**WindMouse** アルゴリズム：目標への「重力」＋ランダムな「風」を速度に加え続け、
直線でない弧・微妙な揺れ・終端の減速を再現する。検出側は「直線で一瞬移動」を疑う。

### 3. 人間らしいタイピング (`humanType.js`)
打鍵間隔を正規分布でばらつかせ、空白・句読点で長めに止まり、たまに隣キーをミスして Backspace で訂正する。

### 4. 起動オプション (`stealthBrowser.js`)
- `playwright-extra` + stealth プラグインで既知フラグを一括除去
- `--disable-blink-features=AutomationControlled`、`ignoreDefaultArgs: ['--enable-automation']`
- `launchPersistentContext` で **永続プロファイル**（Cookie/履歴の“生活感”）

---

## 🧭 学習ロードマップ（どう勉強するか）

**Step 0｜まず「検出される側」を見る**
何も対策しない素の Playwright で `bot.sannysoft.com` を開き、赤い項目を確認する。
→ 「何が検出されるのか」を体感するのが出発点。

**Step 1｜足跡を1つずつ潰す**
`fingerprint.js` のブロックを**1つだけ有効化**しては再テスト。
`navigator.webdriver` → Canvas → WebGL… と、項目が緑に変わる様子を観察する。
（“魔法の一括ライブラリ”ではなく、原理を1つずつ理解するのが目的）

**Step 2｜挙動の自然さ**
`humanMouse.js` の WindMouse のパラメータ（G/W/maxStep）を変えて、軌跡を可視化してみる。
`page.on('console')` や軌跡を canvas に描くと違いが分かる。

**Step 3｜一貫性チェック**
`creepjs` で **trust score** を見る。指紋は「強く隠す」より「矛盾なく一貫」が効く、を確認する。

**Step 4｜上級トピック（必要になったら）**
- CDP 痕跡の隠蔽：`rebrowser-patches` / `patchright`
- TLS/JA3 指紋（ブラウザ利用なら基本問題にならない理由）
- reCAPTCHA v3 / Cloudflare Turnstile が「挙動スコア」で年々強くなっている現実

### 推奨の学び方
- **手を動かす → 再テスト → 差分を観察** のループを回す
- 1コミット＝1テクニック、で小さく試す
- 「なぜ検出されるか（攻撃者視点）」と「なぜ防ぐか（防御者視点）」の両面で考える

---

## 🔗 参考リンク

- Playwright 公式: https://playwright.dev/
- playwright-extra: https://github.com/berstend/puppeteer-extra/tree/master/packages/playwright-extra
- 検出テスト（sannysoft）: https://bot.sannysoft.com/
- 指紋一貫性（CreepJS）: https://abrahamjuliot.github.io/creepjs/
- WindMouse アルゴリズム解説（BenLand100）: https://ben.land/post/2021/04/25/windmouse-human-mouse-movement/

---

## ⚖️ ライセンス
MIT（学習目的）。利用は各サイトの規約・法令の範囲内で。
