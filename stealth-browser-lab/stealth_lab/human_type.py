"""人間らしいタイピング。

ボットは「全文字を等間隔・超高速」で入力しがち。人間は：
  - 文字ごとに打鍵間隔がばらつく（正規分布）
  - 単語の区切り・句読点で少し長く止まる
  - たまにタイプミス→Backspace で訂正する
を再現する。
"""
import random

from .utils import sleep_ms, gaussian_clamped, gaussian

NEIGHBORS = {
    "a": "s", "s": "d", "d": "f", "e": "r", "r": "t",
    "t": "y", "o": "i", "n": "m", "i": "o", "l": "k",
}

PAUSE_CHARS = set(" 、。,.!?！？")


async def type_human(page, selector, text, typo_rate: float = 0.03):
    """セレクタの入力欄に人間らしく文字を打ち込む。"""
    await page.click(selector)
    await sleep_ms(gaussian_clamped(180, 60, 60, 400))  # クリック後の「構え」

    for ch in text:
        low = ch.lower()
        # たまにミスタイプ → 訂正
        if random.random() < typo_rate and low in NEIGHBORS:
            await page.keyboard.type(NEIGHBORS[low])
            await sleep_ms(gaussian_clamped(140, 50, 60, 320))
            await page.keyboard.press("Backspace")
            await sleep_ms(gaussian_clamped(120, 40, 50, 260))

        await page.keyboard.type(ch)

        delay = gaussian_clamped(110, 45, 35, 320)
        # 空白・句読点の後は少し長めに「考える」
        if ch in PAUSE_CHARS:
            delay += abs(gaussian(140, 60))
        await sleep_ms(delay)
