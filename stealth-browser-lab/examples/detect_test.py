"""複数の検出テストサイトを巡回し、結果のスクショを保存する学習用スクリプト。

実行（リポジトリ直下から）: python -m examples.detect_test
  または: python stealth-browser-lab/examples/detect_test.py

見るべきサイト：
  - bot.sannysoft.com … 自動化フラグ一覧（緑=OK / 赤=検出）
  - abrahamjuliot.github.io/creepjs … 指紋の一貫性スコア（trust score）
"""
import asyncio
import os
import sys

# リポジトリ直下を import パスに追加（直接実行された場合の保険）
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from stealth_lab.stealth_browser import launch_stealth, close_all  # noqa: E402
from stealth_lab.human_mouse import move_mouse_human  # noqa: E402
from stealth_lab.utils import rand, sleep_ms  # noqa: E402

RESULTS_DIR = "results"
TARGETS = [
    {"name": "sannysoft", "url": "https://bot.sannysoft.com/"},
    {"name": "creepjs", "url": "https://abrahamjuliot.github.io/creepjs/"},
]


async def main() -> None:
    os.makedirs(RESULTS_DIR, exist_ok=True)
    handles = await launch_stealth(headless=False)
    page = handles["page"]
    try:
        for t in TARGETS:
            print(f"[i] 検査中: {t['name']}")
            await page.goto(t["url"], wait_until="networkidle", timeout=60000)
            # creepjs はスコア計算に時間がかかるので少し待つ
            await sleep_ms(8000 if t["name"] == "creepjs" else 1500)
            await move_mouse_human(page, rand(200, 1000), rand(150, 550))
            out = os.path.join(RESULTS_DIR, f"{t['name']}.png")
            await page.screenshot(path=out, full_page=True)
            print(f"[✓] 保存: {out}")
    finally:
        await close_all(handles)


if __name__ == "__main__":
    asyncio.run(main())
