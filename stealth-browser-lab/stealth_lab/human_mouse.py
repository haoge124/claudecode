"""人間らしいマウス軌跡の生成（WindMouse アルゴリズム）。

目標への「重力(gravity)」とランダムに揺れる「風(wind)」を速度ベクトルに
加え続けることで、直線でない自然な弧・微妙な揺れ・終端の減速を再現する。
検出側は「点A→点Bへ完全な直線で一瞬移動」を疑うため、ここが効く。

参考: BenLand100 による WindMouse（公開アルゴリズム）。
"""
import math
import random

from .utils import rand, gaussian_clamped, sleep_ms

SQRT3 = math.sqrt(3)
SQRT5 = math.sqrt(5)

# ページごとの最後のマウス座標を覚えておく（id(page) -> (x, y)）
_last_pos: dict[int, tuple[float, float]] = {}


def wind_mouse(x0, y0, x1, y1, G=9.0, W=3.0, max_step=15.0, target_area=12.0):
    """始点(x0,y0)→終点(x1,y1) の軌跡座標列 [(x, y), ...] を返す。

    G: 重力（大きいほど真っ直ぐ目標へ）
    W: 風（大きいほど揺れる）
    max_step: 1ステップの最大移動量（大きいほど速い）
    target_area: 終端で減速し始める距離
    """
    points = []
    vx = vy = wx = wy = 0.0
    x, y = float(x0), float(y0)
    dist = math.hypot(x1 - x, y1 - y)
    guard = 0
    while dist >= 1 and guard < 10000:
        guard += 1
        w_mag = min(W, dist)
        if dist >= target_area:
            # 目標から遠い間は風がランダムに揺らす
            wx = wx / SQRT3 + (2 * random.random() - 1) * w_mag / SQRT5
            wy = wy / SQRT3 + (2 * random.random() - 1) * w_mag / SQRT5
        else:
            # 目標に近づいたら風を減衰、ステップも細かくして「吸い付く」
            wx /= SQRT3
            wy /= SQRT3
            if max_step < 3:
                max_step = rand(3, 6)
            else:
                max_step /= SQRT5
        # 風 + 重力（目標方向）を速度に加える
        vx += wx + G * (x1 - x) / dist
        vy += wy + G * (y1 - y) / dist
        v_mag = math.hypot(vx, vy)
        if v_mag > max_step:
            v_clip = max_step / 2 + rand(0, max_step / 2)
            vx = vx / v_mag * v_clip
            vy = vy / v_mag * v_clip
        x += vx
        y += vy
        points.append((round(x), round(y)))
        dist = math.hypot(x1 - x, y1 - y)
    points.append((round(x1), round(y1)))
    return points


async def move_mouse_human(page, to_x, to_y, frm=None):
    """page のマウスを人間らしく移動。各点間に微小なゆらぎ待機を入れる。"""
    start = frm or _last_pos.get(id(page)) or (rand(0, 200), rand(0, 200))
    for px, py in wind_mouse(start[0], start[1], to_x, to_y):
        await page.mouse.move(px, py)
        await sleep_ms(gaussian_clamped(7, 4, 1, 25))
    _last_pos[id(page)] = (to_x, to_y)


async def click_human(page, selector, **kw):
    """要素まで人間らしく動いてクリック。中心ぴったりでなく少しズラす。"""
    el = await page.wait_for_selector(selector, state="visible")
    box = await el.bounding_box()
    if not box:
        raise RuntimeError(f"要素の座標が取れません: {selector}")
    tx = box["x"] + box["width"] / 2 + rand(-box["width"] * 0.2, box["width"] * 0.2)
    ty = box["y"] + box["height"] / 2 + rand(-box["height"] * 0.2, box["height"] * 0.2)
    await move_mouse_human(page, tx, ty)
    await sleep_ms(gaussian_clamped(90, 40, 30, 250))  # 押下前の「ためらい」
    await page.mouse.down()
    await sleep_ms(gaussian_clamped(60, 25, 20, 160))   # 押している時間
    await page.mouse.up()
