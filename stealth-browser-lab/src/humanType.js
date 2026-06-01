'use strict';
/**
 * 人間らしいタイピング。
 *
 * ボットは「全文字を等間隔・超高速」で入力しがち。人間は：
 *   - 文字ごとに打鍵間隔がばらつく（正規分布）
 *   - 単語の区切り・句読点で少し長く止まる
 *   - たまにタイプミス→Backspace で訂正する
 * を再現する。
 */
const { sleep, gaussianClamped, gaussian } = require('./utils');

const NEIGHBORS = {
  a: 's', s: 'd', d: 'f', e: 'r', r: 't', t: 'y', o: 'i', n: 'm', i: 'o', l: 'k',
};

/**
 * セレクタの入力欄に人間らしく文字を打ち込む。
 * @param {import('playwright').Page} page
 * @param {string} selector
 * @param {string} text
 * @param {object} [opts] typoRate: タイプミス確率(0-1)
 */
async function typeHuman(page, selector, text, opts = {}) {
  const typoRate = opts.typoRate ?? 0.03;
  await page.click(selector);
  await sleep(gaussianClamped(180, 60, 60, 400)); // クリック後の「構え」

  for (const ch of text) {
    // たまにミスタイプ → 訂正
    if (Math.random() < typoRate && NEIGHBORS[ch.toLowerCase()]) {
      await page.keyboard.type(NEIGHBORS[ch.toLowerCase()]);
      await sleep(gaussianClamped(140, 50, 60, 320));
      await page.keyboard.press('Backspace');
      await sleep(gaussianClamped(120, 40, 50, 260));
    }

    await page.keyboard.type(ch);

    // 打鍵間隔：平均 110ms 前後でばらつかせる
    let delay = gaussianClamped(110, 45, 35, 320);
    // 空白・句読点の後は少し長めに「考える」
    if (ch === ' ' || '、。,.!?！？'.includes(ch)) {
      delay += Math.abs(gaussian(140, 60));
    }
    await sleep(delay);
  }
}

module.exports = { typeHuman };
