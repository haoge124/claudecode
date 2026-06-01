"""汎用ユーティリティ。乱数・正規分布・非同期sleep など、
「人間っぽさ」を作るための土台になる小さな関数群。"""
import asyncio
import random


def rand(a: float, b: float) -> float:
    """[a, b) の一様乱数。"""
    return random.uniform(a, b)


def rand_int(a: int, b: int) -> int:
    """[a, b] の整数乱数。"""
    return random.randint(a, b)


def gaussian(mu: float = 0.0, sigma: float = 1.0) -> float:
    """正規分布。人間の操作間隔は一様乱数より正規分布が自然。"""
    return random.gauss(mu, sigma)


def gaussian_clamped(mu: float, sigma: float, lo: float, hi: float) -> float:
    """下限・上限でクランプした正規乱数（負の待機などを防ぐ）。"""
    return max(lo, min(hi, random.gauss(mu, sigma)))


async def sleep_ms(ms: float) -> None:
    """ミリ秒指定の非同期 sleep。"""
    await asyncio.sleep(ms / 1000.0)
