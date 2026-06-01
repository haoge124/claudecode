"""デモ：ステルス構成で起動し、人間らしくマウスを動かしてから
ボット検出テストページのスクリーンショットを保存する。

実行: python main.py
事前: pip install -r requirements.txt && playwright install chromium
"""
import asyncio
import os

from stealth_lab.stealth_browser import launch_stealth, close_all
from stealth_lab.human_mouse import move_mouse_human
from stealth_lab.utils import rand, sleep_ms

RESULTS_DIR = "results"


async def main() -> None:
    os.makedirs(RESULTS_DIR, exist_ok=True)
    handles = await launch_stealth(headless=False)  # 検出回避は headed が有利
    page = handles["page"]
    print(f"[i] 起動プロファイル seed={handles['profile']['seed']} "
          f"renderer={handles['profile']['webglRenderer']}")
    try:
        await page.goto("https://bot.sannysoft.com/", wait_until="networkidle")
        # 人間らしくマウスをうろうろさせる
        for _ in range(4):
            await move_mouse_human(page, rand(100, 1200), rand(100, 600))
            await sleep_ms(rand(200, 700))

        out = os.path.join(RESULTS_DIR, "sannysoft.png")
        await page.screenshot(path=out, full_page=True)
        print(f"[✓] 検出テスト結果を保存: {out}（緑が多いほどステルスが効いています）")
    finally:
        await close_all(handles)


if __name__ == "__main__":
    asyncio.run(main())
