'use strict';
/**
 * デモ：ステルス構成で起動し、人間らしくマウスを動かしてから
 * ボット検出テストページのスクリーンショットを保存する。
 *
 * 実行: node src/index.js
 * 事前: npm install && npx playwright install chromium
 */
const path = require('path');
const fs = require('fs');
const { launchStealth } = require('./stealthBrowser');
const { moveMouseHuman } = require('./humanMouse');
const { sleep, rand } = require('./utils');

const RESULTS_DIR = path.join(__dirname, '..', 'results');

async function main() {
  fs.mkdirSync(RESULTS_DIR, { recursive: true });

  const { browser, context, page, profile } = await launchStealth({
    headless: false, // 検出回避は headed の方が有利
  });
  console.log('[i] 起動プロファイル seed=%d renderer=%s', profile.seed, profile.webglRenderer);

  try {
    // 1) 検出テストページへ
    await page.goto('https://bot.sannysoft.com/', { waitUntil: 'networkidle' });

    // 2) 人間らしくマウスをうろうろさせる（行動の自然さを足す）
    for (let i = 0; i < 4; i++) {
      await moveMouseHuman(page, rand(100, 1200), rand(100, 600));
      await sleep(rand(200, 700));
    }

    // 3) スクリーンショットで結果を保存（緑＝検出を回避できた項目）
    const out = path.join(RESULTS_DIR, 'sannysoft.png');
    await page.screenshot({ path: out, fullPage: true });
    console.log('[✓] 検出テスト結果を保存:', out);
    console.log('    緑が多いほどステルスが効いています。');
  } finally {
    await context.close();
    if (browser) await browser.close();
  }
}

main().catch((err) => {
  console.error('[x] 失敗:', err);
  process.exit(1);
});
