"""ブラウザ指紋(fingerprint)の偽装。

ボット検出はブラウザ固有の「指紋」で個体や自動化を見分ける：
  - navigator.* の各種プロパティ（webdriver, languages...）
  - Canvas 描画結果のピクセル（GPU/ドライバ差で微妙に異なる）
  - WebGL の VENDOR / RENDERER 文字列
  - AudioContext の波形

ここでは context.add_init_script で「ページの全スクリプトより前」に
これらを上書き/ノイズ付与する。注入する中身は“ブラウザ内で動く JS”なので、
JS版でも Python版でも全く同じものを使える（ここは言語非依存）。

⚠️ 学習用。実在ブラウザと矛盾する値を入れると逆に怪しまれるので、
   「整合性」が最重要（UA と platform、languages と locale 等を揃える）。
"""
import json
import random

# 実在しそうな構成に寄せた既定プロファイル
DEFAULT_PROFILE = {
    "seed": random.randint(0, 10**9),
    "hardwareConcurrency": 8,
    "deviceMemory": 8,
    "platform": "Win32",
    "languages": ["ja-JP", "ja", "en-US", "en"],
    "webglVendor": "Google Inc. (NVIDIA)",
    "webglRenderer": (
        "ANGLE (NVIDIA, NVIDIA GeForce RTX 3060 Direct3D11 vs_5_0 ps_5_0, D3D11)"
    ),
}

# ブラウザに注入する JS（__PROFILE__ をプロファイルJSONに置換して使う）
_FP_JS = """
(function (p) {
  try {
    Object.defineProperty(Navigator.prototype, 'webdriver', { get: () => false, configurable: true });
  } catch (e) {}

  const defineNav = (k, v) => {
    try { Object.defineProperty(navigator, k, { get: () => v, configurable: true }); } catch (e) {}
  };
  defineNav('hardwareConcurrency', p.hardwareConcurrency);
  defineNav('deviceMemory', p.deviceMemory);
  defineNav('platform', p.platform);
  defineNav('languages', Object.freeze(p.languages.slice()));

  // シード付き擬似乱数（決定論的ノイズ用 / mulberry32）
  let s = p.seed >>> 0;
  const rng = () => {
    s |= 0; s = (s + 0x6d2b79f5) | 0;
    let t = Math.imul(s ^ (s >>> 15), 1 | s);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };

  // Canvas 指紋に極小ノイズ
  try {
    const toDataURL = HTMLCanvasElement.prototype.toDataURL;
    const getImageData = CanvasRenderingContext2D.prototype.getImageData;
    const noisify = (ctx, w, h) => {
      const img = getImageData.call(ctx, 0, 0, w, h);
      for (let i = 0; i < img.data.length; i += 4) {
        const n = Math.floor(rng() * 3) - 1;
        img.data[i] = Math.max(0, Math.min(255, img.data[i] + n));
      }
      ctx.putImageData(img, 0, 0);
    };
    HTMLCanvasElement.prototype.toDataURL = function (...args) {
      try { const ctx = this.getContext('2d'); if (ctx) noisify(ctx, this.width, this.height); } catch (e) {}
      return toDataURL.apply(this, args);
    };
  } catch (e) {}

  // WebGL の VENDOR / RENDERER を偽装
  try {
    const patchGL = (proto) => {
      const gp = proto.getParameter;
      proto.getParameter = function (param) {
        if (param === 37445) return p.webglVendor;   // UNMASKED_VENDOR_WEBGL
        if (param === 37446) return p.webglRenderer;  // UNMASKED_RENDERER_WEBGL
        return gp.call(this, param);
      };
    };
    if (window.WebGLRenderingContext) patchGL(WebGLRenderingContext.prototype);
    if (window.WebGL2RenderingContext) patchGL(WebGL2RenderingContext.prototype);
  } catch (e) {}

  // AudioContext の波形に極小ノイズ
  try {
    const gcd = AudioBuffer.prototype.getChannelData;
    AudioBuffer.prototype.getChannelData = function (...args) {
      const data = gcd.apply(this, args);
      for (let i = 0; i < data.length; i += 100) { data[i] = data[i] + (rng() - 0.5) * 1e-7; }
      return data;
    };
  } catch (e) {}

  // permissions.query の矛盾解消
  try {
    const orig = navigator.permissions.query.bind(navigator.permissions);
    navigator.permissions.query = (params) =>
      (params && params.name === 'notifications')
        ? Promise.resolve({ state: Notification.permission })
        : orig(params);
  } catch (e) {}
})(__PROFILE__);
"""


def build_init_script(profile: dict) -> str:
    """プロファイルを焼き込んだ、注入用 JS 文字列を返す。"""
    return _FP_JS.replace("__PROFILE__", json.dumps(profile))


async def apply_fingerprint(context, **override) -> dict:
    """context に指紋偽装を適用し、使用プロファイルを返す。"""
    profile = {**DEFAULT_PROFILE, **override}
    await context.add_init_script(build_init_script(profile))
    return profile
