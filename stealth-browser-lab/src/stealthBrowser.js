'use strict';
/**
 * ステルス・ブラウザの起動。
 *
 * 3段構えで「本物のブラウザを、本物のユーザーらしく」見せる：
 *   1) playwright-extra + stealth プラグイン … 既知の自動化フラグを自動で潰す
 *   2) 起動オプション … --enable-automation を外し、AutomationControlled を無効化
 *   3) コンテキスト … 実在の UA / viewport / locale / timezone を揃える
 *      ＋ fingerprint.js で Canvas/WebGL/Audio の指紋を偽装
 *
 * ⚠️ 整合性が命：UA を Windows Chrome にするなら platform も Win32、
 *   locale=ja-JP なら timezone=Asia/Tokyo、というように矛盾を作らないこと。
 */
const { chromium } = require('playwright-extra');
const StealthPlugin = require('puppeteer-extra-plugin-stealth');
const { applyFingerprint } = require('./fingerprint');

chromium.use(StealthPlugin());

// 実在の Chrome に寄せた User-Agent（バージョンは実際の Chromium に合わせて更新する）
const REALISTIC_UA =
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 ' +
  '(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36';

/** 自動化臭を減らす起動引数 */
const STEALTH_ARGS = [
  '--disable-blink-features=AutomationControlled',
  '--no-first-run',
  '--no-default-browser-check',
  '--disable-infobars',
];

/**
 * ステルス構成でブラウザ・コンテキスト・ページを起動する。
 * @param {object} [opts]
 *   headless: boolean（既定 false。検出回避は headed の方が有利）
 *   userDataDir: string（指定すると永続プロファイル＝Cookie/履歴が残る）
 *   fingerprint: object（fingerprint プロファイル上書き）
 * @returns {Promise<{browser, context, page, profile}>}
 */
async function launchStealth(opts = {}) {
  const headless = opts.headless ?? false;
  const contextOptions = {
    userAgent: opts.userAgent || REALISTIC_UA,
    viewport: opts.viewport || { width: 1366, height: 768 },
    locale: opts.locale || 'ja-JP',
    timezoneId: opts.timezoneId || 'Asia/Tokyo',
    // 実画面に近い色域・端末スケール
    deviceScaleFactor: 1,
  };

  let browser = null;
  let context;
  if (opts.userDataDir) {
    // 永続コンテキスト：最も「生活感」が出る（推奨）
    context = await chromium.launchPersistentContext(opts.userDataDir, {
      headless,
      args: STEALTH_ARGS,
      ignoreDefaultArgs: ['--enable-automation'],
      ...contextOptions,
    });
  } else {
    browser = await chromium.launch({
      headless,
      args: STEALTH_ARGS,
      ignoreDefaultArgs: ['--enable-automation'],
    });
    context = await browser.newContext(contextOptions);
  }

  const profile = await applyFingerprint(context, opts.fingerprint || {});
  const page = context.pages()[0] || (await context.newPage());

  return { browser, context, page, profile };
}

module.exports = { launchStealth, REALISTIC_UA, STEALTH_ARGS };
