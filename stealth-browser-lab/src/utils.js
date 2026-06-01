'use strict';
/**
 * 汎用ユーティリティ。乱数・待機・正規分布など、
 * 「人間っぽさ」を作るための土台になる小さな関数群。
 */

/** [min, max) の一様乱数 */
function rand(min, max) {
  return min + Math.random() * (max - min);
}

/** [min, max] の整数乱数 */
function randInt(min, max) {
  return Math.floor(rand(min, max + 1));
}

/** ガウス分布（Box-Muller 法）。平均 mu, 標準偏差 sigma。
 *  人間の操作間隔は一様乱数より正規分布の方が自然。 */
function gaussian(mu = 0, sigma = 1) {
  let u = 0;
  let v = 0;
  while (u === 0) u = Math.random();
  while (v === 0) v = Math.random();
  const z = Math.sqrt(-2.0 * Math.log(u)) * Math.cos(2.0 * Math.PI * v);
  return mu + sigma * z;
}

/** 下限・上限でクランプした正規乱数（負の待機時間などを防ぐ） */
function gaussianClamped(mu, sigma, min, max) {
  return Math.max(min, Math.min(max, gaussian(mu, sigma)));
}

/** Promise ベースの sleep */
function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/** 配列からランダムに1つ選ぶ */
function choice(arr) {
  return arr[randInt(0, arr.length - 1)];
}

module.exports = { rand, randInt, gaussian, gaussianClamped, sleep, choice };
