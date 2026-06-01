'use strict';
/**
 * ブラウザ指紋(fingerprint)の偽装。
 *
 * ボット検出は、ブラウザ固有の「指紋」で個体や自動化を見分ける：
 *   - navigator.* の各種プロパティ（webdriver, plugins, languages...）
 *   - Canvas 描画結果のピクセル（GPU/ドライバ差で微妙に異なる）
 *   - WebGL の VENDOR / RENDERER 文字列
 *   - AudioContext の波形
 *
 * ここでは context.addInitScript で「ページの全スクリプトより前」に
 * これらを上書き/ノイズ付与する。seed を固定すると、そのセッション内では
 * 指紋が安定する（毎アクセスでバラつくのも逆に不自然なため）。
 *
 * ⚠️ 学習用。実在ブラウザと矛盾する値を入れると逆に怪しまれるので、
 *   「整合性」が最重要（UA と platform、languages と locale 等を揃える）。
 */

/** 既定の偽装プロファイル（実在しそうな構成に寄せる） */
const DEFAULT_PROFILE = {
  seed: Math.floor(Math.random() * 1e9),
  hardwareConcurrency: 8,
  deviceMemory: 8,
  platform: 'Win32',
  languages: ['ja-JP', 'ja', 'en-US', 'en'],
  webglVendor: 'Google Inc. (NVIDIA)',
  webglRenderer:
    'ANGLE (NVIDIA, NVIDIA GeForce RTX 3060 Direct3D11 vs_5_0 ps_5_0, D3D11)',
};

/**
 * page/context の addInitScript に注入するスクリプト本体を組み立てる。
 * プロファイルをクロージャに焼き込み、ページ側で実行される文字列にする。
 */
function buildInitScript(profile) {
  // 関数を文字列化し、JSON プロファイルを引数として渡す
  const fn = function (p) {
    // --- 1) navigator.webdriver を消す（自動化の最有名フラグ） ---
    try {
      Object.defineProperty(Navigator.prototype, 'webdriver', {
        get: () => false,
        configurable: true,
      });
    } catch (e) {}

    // --- 2) navigator の基本プロパティを上書き ---
    const defineNav = (key, value) => {
      try {
        Object.defineProperty(navigator, key, { get: () => value, configurable: true });
      } catch (e) {}
    };
    defineNav('hardwareConcurrency', p.hardwareConcurrency);
    defineNav('deviceMemory', p.deviceMemory);
    defineNav('platform', p.platform);
    defineNav('languages', Object.freeze(p.languages.slice()));

    // --- 3) シード付き擬似乱数（決定論的ノイズ用 / mulberry32） ---
    let s = p.seed >>> 0;
    const rng = () => {
      s |= 0;
      s = (s + 0x6d2b79f5) | 0;
      let t = Math.imul(s ^ (s >>> 15), 1 | s);
      t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
      return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
    };

    // --- 4) Canvas 指紋にごく微小なノイズ ---
    try {
      const toDataURL = HTMLCanvasElement.prototype.toDataURL;
      const getImageData = CanvasRenderingContext2D.prototype.getImageData;
      const noisify = (ctx, w, h) => {
        const img = getImageData.call(ctx, 0, 0, w, h);
        for (let i = 0; i < img.data.length; i += 4) {
          // 各チャンネルを ±1 だけ確率的に動かす（人間の目では不可視）
          const n = Math.floor(rng() * 3) - 1;
          img.data[i] = Math.max(0, Math.min(255, img.data[i] + n));
        }
        ctx.putImageData(img, 0, 0);
      };
      HTMLCanvasElement.prototype.toDataURL = function (...args) {
        try {
          const ctx = this.getContext('2d');
          if (ctx) noisify(ctx, this.width, this.height);
        } catch (e) {}
        return toDataURL.apply(this, args);
      };
    } catch (e) {}

    // --- 5) WebGL の VENDOR / RENDERER を偽装 ---
    try {
      const patchGL = (proto) => {
        const getParameter = proto.getParameter;
        proto.getParameter = function (param) {
          // UNMASKED_VENDOR_WEBGL = 37445, UNMASKED_RENDERER_WEBGL = 37446
          if (param === 37445) return p.webglVendor;
          if (param === 37446) return p.webglRenderer;
          return getParameter.call(this, param);
        };
      };
      if (window.WebGLRenderingContext) patchGL(WebGLRenderingContext.prototype);
      if (window.WebGL2RenderingContext) patchGL(WebGL2RenderingContext.prototype);
    } catch (e) {}

    // --- 6) AudioContext の波形に極小ノイズ ---
    try {
      const getChannelData = AudioBuffer.prototype.getChannelData;
      AudioBuffer.prototype.getChannelData = function (...args) {
        const data = getChannelData.apply(this, args);
        for (let i = 0; i < data.length; i += 100) {
          data[i] = data[i] + (rng() - 0.5) * 1e-7;
        }
        return data;
      };
    } catch (e) {}

    // --- 7) permissions.query の矛盾解消（通知が "denied" のままになるバグ対策） ---
    try {
      const orig = navigator.permissions.query.bind(navigator.permissions);
      navigator.permissions.query = (params) =>
        params && params.name === 'notifications'
          ? Promise.resolve({ state: Notification.permission })
          : orig(params);
    } catch (e) {}
  };

  return `(${fn.toString()})(${JSON.stringify(profile)});`;
}

/**
 * context に指紋偽装を適用する。
 * @param {import('playwright').BrowserContext} context
 * @param {object} [override] プロファイル上書き
 * @returns {object} 適用したプロファイル
 */
async function applyFingerprint(context, override = {}) {
  const profile = { ...DEFAULT_PROFILE, ...override };
  await context.addInitScript(buildInitScript(profile));
  return profile;
}

module.exports = { applyFingerprint, buildInitScript, DEFAULT_PROFILE };
