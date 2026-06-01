"""ステルス・ブラウザの起動（Playwright async）。

3段構えで「本物のブラウザを、本物のユーザーらしく」見せる：
  1) 起動オプション … --enable-automation を外し、AutomationControlled を無効化
  2) コンテキスト   … 実在の UA / viewport / locale / timezone を整合させる
  3) 指紋偽装       … fingerprint.py の JS を add_init_script で注入
  （+ playwright-stealth が入っていれば追加で適用）

⚠️ 整合性が命：UA を Windows Chrome にするなら platform も Win32、
   locale=ja-JP なら timezone=Asia/Tokyo、と矛盾を作らないこと。
"""
from playwright.async_api import async_playwright

from .fingerprint import apply_fingerprint

# 実在の Chrome に寄せた UA（実際の Chromium バージョンに合わせて更新する）
REALISTIC_UA = (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
)

STEALTH_ARGS = [
    "--disable-blink-features=AutomationControlled",
    "--no-first-run",
    "--no-default-browser-check",
    "--disable-infobars",
]


async def launch_stealth(
    headless: bool = False,
    user_data_dir: str | None = None,
    user_agent: str | None = None,
    fingerprint: dict | None = None,
) -> dict:
    """ステルス構成でブラウザを起動し、ハンドル一式を返す。

    戻り値の dict: {pw, browser, context, page, profile}
    終了時は close_all(handles) を呼ぶこと。
    """
    pw = await async_playwright().start()
    context_kwargs = dict(
        user_agent=user_agent or REALISTIC_UA,
        viewport={"width": 1366, "height": 768},
        locale="ja-JP",
        timezone_id="Asia/Tokyo",
        device_scale_factor=1,
    )

    browser = None
    if user_data_dir:
        # 永続コンテキスト：最も「生活感」が出る（推奨）
        context = await pw.chromium.launch_persistent_context(
            user_data_dir,
            headless=headless,
            args=STEALTH_ARGS,
            ignore_default_args=["--enable-automation"],
            **context_kwargs,
        )
    else:
        browser = await pw.chromium.launch(
            headless=headless,
            args=STEALTH_ARGS,
            ignore_default_args=["--enable-automation"],
        )
        context = await browser.new_context(**context_kwargs)

    page = context.pages[0] if context.pages else await context.new_page()

    # 任意：playwright-stealth が入っていれば追加適用（無ければ自前指紋のみ）
    try:
        from playwright_stealth import stealth_async  # type: ignore

        await stealth_async(page)
    except Exception:
        pass

    profile = await apply_fingerprint(context, **(fingerprint or {}))
    return {"pw": pw, "browser": browser, "context": context, "page": page, "profile": profile}


async def close_all(handles: dict) -> None:
    """launch_stealth が返したハンドルを安全に閉じる。"""
    try:
        await handles["context"].close()
    finally:
        if handles.get("browser"):
            await handles["browser"].close()
        await handles["pw"].stop()
