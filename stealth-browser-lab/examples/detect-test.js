'use strict';
/**
 * 複数の検出テストサイトを巡回し、結果のスクショを保存する学習用スクリプト。
 *
 * 実行: node examples/detect-test.js
 *
 * 見るべきサイト：
 *   - bot.sannysoft.com … 自動化フラグ一覧（緑=OK / 赤=検出）
 *   - abrahamjuliot.github.io/creepjs … 指紋の一貫性スコア（trust score）
 */
const path = require('path');
const fs = require('fs');
const { launchStealth } = require('../src/stealthBrowser');
const { moveMouseHuman } = require('../src/humanMouse');
const { sleep, rand } = require('../src/utils');

const RESULTS_DIR = path.join(__dirname, '..', 'results');

const TARGETS = [
  { name: 'sannysoft', url: 'https://bot.sannysoft.com/' },
  { name: 'creepjs', url: 'https://abrahamjuliot.github.io/creepjs/' },
];

async function main() {
  fs.mkdirSync(RESULTS_DIR, { recursive: true });
  const { browser, context, page } = await launchStealth({ headless: false });

  try {
    for (const t of TARGETS) {
      console.log('[i] 検査中:', t.name);
      await page.goto(t.url, { waitUntil: 'networkidle', timeout: 60000 });
      // creepjs はスコア計算に時間がかかるので少し待つ
      await sleep(t.name === 'creepjs' ? 8000 : 1500);
      await moveMouseHuman(page, rand(200, 1000), rand(150, 550));

      const out = path.join(RESULTS_DIR, `${t.name}.png`);
      await page.screenshot({ path: out, fullPage: true });
      console.log('[✓] 保存:', out);
    }
  } finally {
    await context.close();
    if (browser) await browser.close();
  }
}

main().catch((err) => {
  console.error('[x] 失敗:', err);
  process.exit(1);
});
