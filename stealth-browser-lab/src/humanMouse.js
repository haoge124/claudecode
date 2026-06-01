'use strict';
/**
 * 人間らしいマウス軌跡の生成。
 *
 * 中核は「WindMouse」アルゴリズム。
 * 目標地点への「重力(gravity)」と、ランダムに揺れる「風(wind)」を
 * 速度ベクトルに加え続けることで、直線ではない自然な弧・微妙な揺れ・
 * 終端での減速を再現する。ボット検出は「点A→点Bへ完全な直線で一瞬移動」
 * のような非人間的な軌跡を疑うため、ここが効く。
 *
 * 参考: BenLand100 による WindMouse（公開アルゴリズム）。
 */
const { sleep, gaussianClamped, rand } = require('./utils');

const SQRT3 = Math.sqrt(3);
const SQRT5 = Math.sqrt(5);

/**
 * 始点(x0,y0)→終点(x1,y1) の軌跡座標列を返す。
 * @param {object} o 調整パラメータ
 *   G: 重力（大きいほど真っ直ぐ目標へ）
 *   W: 風（大きいほど揺れる）
 *   maxStep: 1ステップの最大移動量（大きいほど速い）
 *   targetArea: 終端で減速し始める距離
 * @returns {Array<[number,number]>}
 */
function windMouse(x0, y0, x1, y1, o = {}) {
  const G = o.G ?? 9;
  const W = o.W ?? 3;
  let maxStep = o.maxStep ?? 15;
  const targetArea = o.targetArea ?? 12;

  const points = [];
  let vx = 0;
  let vy = 0;
  let wx = 0;
  let wy = 0;
  let x = x0;
  let y = y0;

  let dist = Math.hypot(x1 - x, y1 - y);
  // 無限ループ保護
  let guard = 0;
  while (dist >= 1 && guard++ < 10000) {
    const wMag = Math.min(W, dist);
    if (dist >= targetArea) {
      // 目標から遠い間は風がランダムに揺らす
      wx = wx / SQRT3 + ((2 * Math.random() - 1) * wMag) / SQRT5;
      wy = wy / SQRT3 + ((2 * Math.random() - 1) * wMag) / SQRT5;
    } else {
      // 目標に近づいたら風を減衰させ、ステップも細かくして「吸い付く」
      wx /= SQRT3;
      wy /= SQRT3;
      if (maxStep < 3) maxStep = rand(3, 6);
      else maxStep /= SQRT5;
    }
    // 風 + 重力（目標方向）を速度に加える
    vx += wx + (G * (x1 - x)) / dist;
    vy += wy + (G * (y1 - y)) / dist;

    // 速度が maxStep を超えたらクリップ（=最高速の制限）
    const vMag = Math.hypot(vx, vy);
    if (vMag > maxStep) {
      const vClip = maxStep / 2 + rand(0, maxStep / 2);
      vx = (vx / vMag) * vClip;
      vy = (vy / vMag) * vClip;
    }

    x += vx;
    y += vy;
    points.push([Math.round(x), Math.round(y)]);
    dist = Math.hypot(x1 - x, y1 - y);
  }
  points.push([Math.round(x1), Math.round(y1)]);
  return points;
}

/**
 * Playwright の page でマウスを人間らしく移動させる。
 * 各点の間に微小なゆらぎ待機を入れて、移動速度に「ムラ」を作る。
 */
async function moveMouseHuman(page, toX, toY, opts = {}) {
  const from = opts.from || page.__lastMouse || { x: rand(0, 200), y: rand(0, 200) };
  const path = windMouse(from.x, from.y, toX, toY, opts);
  for (const [px, py] of path) {
    await page.mouse.move(px, py);
    // 1点ごとの待機 (ms)。短いが一定でないのがポイント。
    await sleep(gaussianClamped(7, 4, 1, 25));
  }
  page.__lastMouse = { x: toX, y: toY };
}

/**
 * 要素まで人間らしく動いてクリック。
 * 要素中心ぴったりではなく、少しズラした点を狙うのも非ボットらしさ。
 */
async function clickHuman(page, selector, opts = {}) {
  const el = await page.waitForSelector(selector, { state: 'visible' });
  const box = await el.boundingBox();
  if (!box) throw new Error(`要素の座標が取れません: ${selector}`);
  const targetX = box.x + box.width / 2 + rand(-box.width * 0.2, box.width * 0.2);
  const targetY = box.y + box.height / 2 + rand(-box.height * 0.2, box.height * 0.2);
  await moveMouseHuman(page, targetX, targetY, opts);
  await sleep(gaussianClamped(90, 40, 30, 250)); // 押下前の「ためらい」
  await page.mouse.down();
  await sleep(gaussianClamped(60, 25, 20, 160)); // 押している時間
  await page.mouse.up();
}

module.exports = { windMouse, moveMouseHuman, clickHuman };
