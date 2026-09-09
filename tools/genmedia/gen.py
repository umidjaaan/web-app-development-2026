#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Генератор медиа для ЛР1 (РИП, ИУ5).

Рисует импортную керамику как тела вращения: профиль задаёт силуэт,
цилиндрическая светотень даёт объём, орнамент «заворачивается» по окружности
(шаг сжимается к краям через arcsin), поверх ложатся следы гончарного круга,
блик на лаке и тень на плоскости.

Запуск:  python3 tools/genmedia/gen.py
Выход:   media/img/<slug>.jpg      (1080x1920, постер ленты и карточка плитки)
         media/video/<slug>.mp4    (720x1280, лента)
"""

import math
import os
import subprocess
import sys

import numpy as np
from PIL import Image, ImageDraw, ImageFilter

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
IMG_DIR = os.path.join(ROOT, "media", "img")
VID_DIR = os.path.join(ROOT, "media", "video")

BEIGE = (232, 232, 225)
GRAPHITE = (26, 26, 26)
YELLOW = (254, 207, 0)
SAND = (184, 184, 168)

SS = 2                      # суперсэмплинг силуэта
UNIT = 96                   # размер ячейки орнамента в пикселях


# ==========================================================================
# профиль сосуда
# ==========================================================================
def smooth_profile(points, n=1200):
    """[(доля высоты, доля радиуса)] -> сглаженный профиль из n точек."""
    ys = np.array([p[0] for p in points], dtype=np.float64)
    rs = np.array([p[1] for p in points], dtype=np.float64)
    t = np.linspace(0.0, 1.0, n)
    r = np.interp(t, ys, rs)
    k = 61
    ker = np.hanning(k)
    ker /= ker.sum()
    r = np.convolve(np.pad(r, (k, k), mode="edge"), ker, mode="same")[k:-k]
    return t, r


def bezier(p0, p1, p2, n=80):
    out = []
    for i in range(n + 1):
        s = i / n
        x = (1 - s) ** 2 * p0[0] + 2 * (1 - s) * s * p1[0] + s ** 2 * p2[0]
        y = (1 - s) ** 2 * p0[1] + 2 * (1 - s) * s * p1[1] + s ** 2 * p2[1]
        out.append((x, y))
    return out


# ==========================================================================
# формы
# ==========================================================================
SHAPES = {
    "amphora_bulbous_neck": dict(
        profile=[(0.00, 0.30), (0.03, 0.32), (0.06, 0.26), (0.12, 0.34),
                 (0.18, 0.30), (0.22, 0.22), (0.30, 0.62), (0.45, 0.95),
                 (0.60, 0.98), (0.75, 0.72), (0.88, 0.34), (0.96, 0.13),
                 (1.00, 0.07)],
        handles=(0.16, 0.34, 1.15), ratio=0.42, mouth=0.30),
    "amphora_biconic": dict(
        profile=[(0.00, 0.26), (0.03, 0.28), (0.07, 0.21), (0.20, 0.20),
                 (0.26, 0.30), (0.40, 0.78), (0.52, 1.00), (0.62, 0.88),
                 (0.80, 0.48), (0.93, 0.20), (1.00, 0.10)],
        handles=(0.11, 0.30, 1.10), ratio=0.44, mouth=0.26),
    "amphora_ovoid": dict(
        profile=[(0.00, 0.28), (0.04, 0.30), (0.08, 0.22), (0.22, 0.21),
                 (0.28, 0.34), (0.42, 0.82), (0.55, 0.97), (0.70, 0.86),
                 (0.85, 0.52), (0.95, 0.24), (1.00, 0.13)],
        handles=(0.12, 0.32, 1.05), ratio=0.45, mouth=0.28),
    "amphora_horned": dict(
        profile=[(0.00, 0.22), (0.03, 0.25), (0.07, 0.18), (0.26, 0.18),
                 (0.32, 0.30), (0.46, 0.72), (0.58, 0.92), (0.72, 0.82),
                 (0.86, 0.48), (0.95, 0.20), (1.00, 0.09)],
        handles=(0.05, 0.30, 1.35), ratio=0.42, mouth=0.22),
    "amphora_double_handle": dict(
        profile=[(0.00, 0.20), (0.03, 0.23), (0.07, 0.16), (0.28, 0.16),
                 (0.34, 0.28), (0.48, 0.70), (0.60, 0.88), (0.74, 0.78),
                 (0.88, 0.44), (0.96, 0.18), (1.00, 0.08)],
        handles=(0.08, 0.33, 1.10), double=True, ratio=0.40, mouth=0.20),
    "kylix": dict(
        profile=[(0.00, 1.00), (0.05, 0.99), (0.13, 0.92), (0.23, 0.76),
                 (0.31, 0.55), (0.38, 0.31), (0.44, 0.13), (0.62, 0.10),
                 (0.78, 0.11), (0.88, 0.20), (0.95, 0.40), (1.00, 0.48)],
        handles=(0.03, 0.11, 0.34), horizontal=True, ratio=1.45, mouth=1.00),
    "kantharos": dict(
        profile=[(0.00, 0.86), (0.10, 0.80), (0.22, 0.66), (0.34, 0.62),
                 (0.46, 0.72), (0.56, 0.66), (0.66, 0.40), (0.76, 0.20),
                 (0.86, 0.16), (0.94, 0.34), (1.00, 0.44)],
        handles=(-0.16, 0.42, 0.95), ratio=0.80, mouth=0.86),
    "bowl_hemisphere": dict(
        profile=[(0.00, 1.00), (0.10, 0.99), (0.30, 0.93), (0.50, 0.82),
                 (0.70, 0.63), (0.85, 0.40), (0.95, 0.18), (1.00, 0.04)],
        ratio=1.55, mouth=1.00),
    "plate": dict(
        profile=[(0.00, 1.00), (0.08, 0.98), (0.24, 0.88), (0.44, 0.72),
                 (0.62, 0.52), (0.78, 0.36), (0.90, 0.30), (1.00, 0.28)],
        ratio=2.30, mouth=1.00),
    "bowl_incurved": dict(
        profile=[(0.00, 0.90), (0.06, 0.98), (0.18, 1.00), (0.38, 0.94),
                 (0.58, 0.78), (0.76, 0.54), (0.90, 0.28), (1.00, 0.10)],
        ratio=1.60, mouth=0.90),
    "alabastron": dict(
        profile=[(0.00, 0.52), (0.04, 0.56), (0.09, 0.24), (0.20, 0.22),
                 (0.28, 0.46), (0.42, 0.72), (0.60, 0.80), (0.78, 0.72),
                 (0.92, 0.48), (1.00, 0.30)],
        ratio=0.34, mouth=0.52),
    "aryballos": dict(
        profile=[(0.00, 0.60), (0.05, 0.64), (0.11, 0.26), (0.22, 0.24),
                 (0.32, 0.56), (0.48, 0.88), (0.64, 0.96), (0.80, 0.82),
                 (0.92, 0.52), (1.00, 0.26)],
        ratio=0.62, mouth=0.60),
}


# ==========================================================================
# орнаменты: ячейка рисуется один раз и потом сэмплируется
# ==========================================================================
def _unit_meander():
    im = Image.new("L", (UNIT, UNIT), 0)
    d = ImageDraw.Draw(im)
    u, w = UNIT, max(3, UNIT // 9)
    d.line([(0, u * 0.86), (u * 0.80, u * 0.86), (u * 0.80, u * 0.16),
            (u * 0.22, u * 0.16), (u * 0.22, u * 0.60), (u * 0.56, u * 0.60),
            (u * 0.56, u * 0.40)], fill=255, width=w, joint="curve")
    return im


def _unit_rays():
    im = Image.new("L", (UNIT, UNIT), 0)
    ImageDraw.Draw(im).polygon(
        [(UNIT * 0.5, 0), (UNIT * 0.92, UNIT), (UNIT * 0.08, UNIT)], fill=255)
    return im


def _unit_tongues():
    im = Image.new("L", (UNIT, UNIT), 0)
    d = ImageDraw.Draw(im)
    d.pieslice([UNIT * 0.06, -UNIT * 0.55, UNIT * 0.94, UNIT * 0.95],
               0, 180, fill=255)
    d.line([(UNIT * 0.5, 0), (UNIT * 0.5, UNIT)], fill=0,
           width=max(2, UNIT // 22))
    return im


def _unit_zigzag():
    im = Image.new("L", (UNIT, UNIT), 0)
    ImageDraw.Draw(im).line(
        [(0, UNIT * 0.85), (UNIT * 0.5, UNIT * 0.15), (UNIT, UNIT * 0.85)],
        fill=255, width=max(4, UNIT // 8))
    return im


def _unit_figure():
    """Силуэт в чернофигурной манере: фигура в профиль."""
    im = Image.new("L", (UNIT, UNIT), 0)
    d = ImageDraw.Draw(im)
    u = UNIT
    d.ellipse([u * .40, u * .06, u * .60, u * .26], fill=255)          # голова
    d.polygon([(u * .34, u * .26), (u * .66, u * .26),
               (u * .60, u * .60), (u * .40, u * .60)], fill=255)      # торс
    d.line([(u * .46, u * .58), (u * .30, u * .96)], fill=255, width=max(4, u // 12))
    d.line([(u * .56, u * .58), (u * .74, u * .96)], fill=255, width=max(4, u // 12))
    d.line([(u * .62, u * .32), (u * .88, u * .18)], fill=255, width=max(3, u // 16))
    d.line([(u * .38, u * .32), (u * .16, u * .46)], fill=255, width=max(3, u // 16))
    return im


def _unit_palmette():
    im = Image.new("L", (UNIT, UNIT), 0)
    d = ImageDraw.Draw(im)
    for i in range(7):
        a = math.pi * (0.12 + 0.76 * i / 6)
        x = UNIT * 0.5 + math.cos(a) * UNIT * 0.42
        y = UNIT * 0.95 - math.sin(a) * UNIT * 0.80
        d.line([(UNIT * 0.5, UNIT * 0.95), (x, y)], fill=255,
               width=max(3, UNIT // 16))
    return im


def _unit_dots():
    im = Image.new("L", (UNIT, UNIT), 0)
    ImageDraw.Draw(im).ellipse(
        [UNIT * 0.32, UNIT * 0.32, UNIT * 0.68, UNIT * 0.68], fill=255)
    return im


def _unit_solid():
    return Image.new("L", (UNIT, UNIT), 255)


PATTERNS = {}


def patterns():
    if not PATTERNS:
        for name, fn in (("meander", _unit_meander), ("rays", _unit_rays),
                         ("tongues", _unit_tongues), ("zigzag", _unit_zigzag),
                         ("figure", _unit_figure), ("palmette", _unit_palmette),
                         ("dots", _unit_dots), ("solid", _unit_solid)):
            PATTERNS[name] = np.asarray(fn(), dtype=np.float32) / 255.0
    return PATTERNS


# ==========================================================================
# рендер сосуда
# ==========================================================================
def _tube(draw, pts, width, colour):
    """Ручка как трубка: тёмная основа, тело, узкий блик."""
    r, g, b = colour
    dark = (max(0, int(r * .55)), max(0, int(g * .55)), max(0, int(b * .55)), 255)
    light = (min(255, int(r * 1.28) + 12), min(255, int(g * 1.28) + 12),
             min(255, int(b * 1.28) + 12), 255)
    draw.line(pts, fill=dark, width=int(width), joint="curve")
    draw.line(pts, fill=colour + (255,), width=max(1, int(width * 0.78)),
              joint="curve")
    shifted = [(x - width * 0.16, y - width * 0.14) for x, y in pts]
    draw.line(shifted, fill=light, width=max(1, int(width * 0.24)),
              joint="curve")


def render_vessel(shape_key, box_w, box_h, clay, gloss, decor, seed=0):
    """Возвращает RGBA-изображение сосуда, вписанного в box_w x box_h."""
    sh = SHAPES[shape_key]
    ratio = sh.get("ratio", 0.45)

    h = box_h
    w = h * ratio
    if w > box_w:
        w, h = box_w, box_w / ratio

    pad_x = w * 0.42                     # запас на ручки
    W = int(w + 2 * pad_x)
    H = int(h * 1.02)
    cx = W / 2.0
    top = (H - h) / 2.0

    # --- радиус по строкам -------------------------------------------------
    t, rprof = smooth_profile(sh["profile"])
    ys = np.arange(H, dtype=np.float32)
    tt = (ys - top) / h
    inside_rows = (tt >= 0.0) & (tt <= 1.0)
    idx = np.clip((tt * (len(rprof) - 1)).astype(np.int32), 0, len(rprof) - 1)
    rr = rprof[idx].astype(np.float32) * (w / 2.0)
    rr[~inside_rows] = 0.0

    xs = np.arange(W, dtype=np.float32)
    dx = xs[None, :] - cx
    safe = np.maximum(rr, 1e-3)[:, None]
    U = dx / safe                                     # -1..1 поперёк тела
    body = (np.abs(U) <= 1.0) & (rr[:, None] > 0.6)

    Uc = np.clip(U, -1.0, 1.0)
    nz = np.sqrt(np.maximum(0.0, 1.0 - Uc * Uc))      # нормаль по глубине

    # --- цилиндрическая светотень (свет слева сверху) ----------------------
    lx, lz = -0.48, 0.88
    lam = np.clip(Uc * lx + nz * lz, 0.0, 1.0)
    shade = 0.34 + 0.82 * lam
    shade -= 0.16 * np.clip((Uc - 0.55) / 0.45, 0.0, 1.0) ** 2   # тень у края
    # блик
    spec = np.exp(-((Uc + 0.42) ** 2) / 0.045)

    # следы гончарного круга
    grooves = 1.0 + 0.030 * np.sin(ys[:, None] / max(2.0, h / 120.0))
    shade = shade * grooves

    # мягкое затемнение к низу тулова
    vshade = 1.0 - 0.10 * np.clip((tt - 0.55) / 0.45, 0.0, 1.0)[:, None]
    shade = shade * vshade

    # --- где лежит лак/роспись --------------------------------------------
    pats = patterns()
    S = (np.arcsin(Uc) / (math.pi / 2.0) + 1.0) / 2.0   # 0..1 по окружности
    glossmap = np.zeros((H, W), dtype=np.float32)
    for band in decor:
        y0, y1, kind, units = band[0], band[1], band[2], band[3]
        if kind == "stamp":
            continue
        invert = len(band) > 4 and band[4] == "invert"
        rows = (tt >= y0) & (tt <= y1)
        if not rows.any():
            continue
        V = np.zeros(H, dtype=np.float32)
        span = max(1e-6, y1 - y0)
        V[rows] = (tt[rows] - y0) / span
        pat = pats[kind]
        px = ((S * units) % 1.0 * (UNIT - 1)).astype(np.int32)
        py = np.clip((V[:, None] * (UNIT - 1)).astype(np.int32), 0, UNIT - 1)
        sel = pat[py, px]
        if invert:
            sel = 1.0 - sel
        glossmap[rows] = np.maximum(glossmap[rows], sel[rows])

    glossmap *= body

    # --- сборка цвета ------------------------------------------------------
    clay_a = np.array(clay, dtype=np.float32)
    gloss_a = np.array(gloss, dtype=np.float32)
    base = clay_a[None, None, :] * (1.0 - glossmap[..., None]) + \
        gloss_a[None, None, :] * glossmap[..., None]

    shine = 0.16 + 0.42 * glossmap                     # лак блестит сильнее
    rgb = base * shade[..., None] + 255.0 * (spec * shine)[..., None]

    # зернистость черепка
    rnd = np.random.default_rng(seed)
    rgb += rnd.normal(0.0, 4.2, (H, W, 1)).astype(np.float32)
    rgb = np.clip(rgb, 0, 255).astype(np.uint8)

    # --- альфа с антиалиасингом краёв --------------------------------------
    edge = (np.abs(U) - 1.0) * safe                    # расстояние до контура
    alpha = np.clip(-edge + 0.5, 0.0, 1.0)
    alpha[rr[:, None].repeat(W, axis=1) <= 0.6] = 0.0
    alpha = (alpha * 255).astype(np.uint8)

    vessel = Image.fromarray(np.dstack([rgb, alpha[..., None]]), "RGBA")

    # --- ручки (за корпусом) ----------------------------------------------
    layer = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    if "handles" in sh:
        d = ImageDraw.Draw(layer)
        y0f, y1f, bulge = sh["handles"]
        i0 = int(np.clip(abs(y0f), 0, 1) * (len(rprof) - 1))
        i1 = int(np.clip(y1f, 0, 1) * (len(rprof) - 1))
        r0 = rprof[i0] * w / 2.0
        r1 = rprof[i1] * w / 2.0
        y0, y1 = top + y0f * h, top + y1f * h
        hw = max(5.0, w * 0.085)
        for sgn in (-1, 1):
            p0 = (cx + sgn * r0, y0)
            p2 = (cx + sgn * r1, y1)
            if sh.get("horizontal"):
                ctrl = (cx + sgn * (max(r0, r1) + w * bulge * 0.45), (y0 + y1) / 2)
            else:
                ctrl = (cx + sgn * (max(r0, r1) + w * bulge * 0.42), y0 - h * 0.02)
            if sh.get("double"):
                for off in (-hw * 0.6, hw * 0.6):
                    _tube(d, bezier((p0[0] + off * .2, p0[1]),
                                    (ctrl[0] + off, ctrl[1]),
                                    (p2[0] + off * .2, p2[1])), hw * 0.7, clay)
            else:
                _tube(d, bezier(p0, ctrl, p2), hw, clay)

    layer.alpha_composite(vessel)

    # --- клеймо на плече ---------------------------------------------------
    if decor and decor[-1][2] == "stamp":
        ty = decor[-1][0]
        i = int(np.clip(ty, 0, 1) * (len(rprof) - 1))
        rx = rprof[i] * w / 2.0
        yy = top + ty * h
        sw_, sh_ = rx * 0.86, h * 0.040
        ds_ = ImageDraw.Draw(layer, "RGBA")
        st = (int(gloss[0] * .8), int(gloss[1] * .8), int(gloss[2] * .8), 210)
        ds_.rounded_rectangle([cx - sw_ * .5, yy, cx + sw_ * .5, yy + sh_],
                              radius=sh_ * .28, outline=st,
                              width=max(2, int(w * 0.008)))
        for j in range(4):
            gx = cx - sw_ * .32 + j * sw_ * .21
            ds_.line([(gx, yy + sh_ * .30), (gx, yy + sh_ * .70)], fill=st,
                     width=max(2, int(w * 0.007)))

    # --- устье: тёмный эллипс внутри --------------------------------------
    mouth = sh.get("mouth", 0.0)
    if mouth > 0.05:
        rx = rprof[0] * w / 2.0
        ry = rx * (0.20 if mouth > 0.8 else 0.26)
        d = ImageDraw.Draw(layer, "RGBA")
        inner = (int(clay[0] * .34), int(clay[1] * .34), int(clay[2] * .34), 255)
        if sum(gloss) < sum(clay):
            inner = (int(gloss[0] * .7), int(gloss[1] * .7), int(gloss[2] * .7), 255)
        d.ellipse([cx - rx, top - ry, cx + rx, top + ry], fill=inner)
        d.ellipse([cx - rx * .9, top - ry * .78, cx + rx * .9, top + ry * .82],
                  outline=(255, 255, 255, 26), width=max(1, int(rx * .05)))
        lip = max(2, int(w * 0.012))
        d.arc([cx - rx, top - ry, cx + rx, top + ry], 0, 180,
              fill=(255, 255, 255, 40), width=lip)

    return layer, W, H


# ==========================================================================
# фон и оформление
# ==========================================================================
def paper(width, height, seed=7):
    rnd = np.random.default_rng(seed)
    yy = np.linspace(0.0, 1.0, height, dtype=np.float32)[:, None]
    base = np.zeros((height, width, 3), dtype=np.float32)
    base[:, :] = BEIGE
    base *= (1.0 - 0.06 * yy)[..., None]                    # лёгкий градиент
    xx = np.linspace(-1.0, 1.0, width, dtype=np.float32)[None, :]
    vign = 1.0 - 0.10 * (xx ** 2 + (2 * yy - 1) ** 2) / 2.0  # виньетка
    base *= vign[..., None]
    base += rnd.normal(0.0, 3.0, (height, width, 1))
    return Image.fromarray(np.clip(base, 0, 255).astype(np.uint8), "RGB") \
        .filter(ImageFilter.GaussianBlur(0.4))


def scale_bar(d, x, y, w, colour):
    d.line([(x, y), (x + w, y)], fill=colour, width=int(2 * SS))
    for i in range(6):
        xx = x + w * i / 5.0
        d.line([(xx, y - 7 * SS), (xx, y + 7 * SS)], fill=colour, width=int(2 * SS))
    for i in range(0, 5, 2):
        d.rectangle([x + w * i / 5.0, y - 6 * SS, x + w * (i + 1) / 5.0, y],
                    fill=colour)


def render(item, width, height, path, quality=88, vessel=0.42):
    """Вертикальный кадр 9:16 с безопасной зоной 4:5 под карточку плитки."""
    W, H = width * SS, height * SS
    im = paper(W, H, seed=item["seed"]).convert("RGBA")

    safe_h = min(H, W * 1.25)
    sy0 = (H - safe_h) / 2.0
    mx = int(W * 0.14)
    my = int(54 * SS)

    d = ImageDraw.Draw(im, "RGBA")
    d.line([(mx, sy0 + my), (W - mx, sy0 + my)], fill=SAND + (150,), width=int(1.5 * SS))
    d.line([(mx, sy0 + safe_h - my), (W - mx, sy0 + safe_h - my)],
           fill=SAND + (150,), width=int(1.5 * SS))
    d.rectangle([mx, sy0 + my - int(9 * SS), mx + int(74 * SS), sy0 + my], fill=YELLOW)

    obj, ow, oh = render_vessel(item["shape"], W * 0.50, H * vessel,
                                item["clay"], item["gloss"], item["decor"],
                                item["seed"])
    ox = int(W / 2 - ow / 2)
    oy = int(H / 2 - oh / 2)

    # тень на плоскости
    shadow = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    ds = ImageDraw.Draw(shadow)
    sr = ow * 0.30
    sy = oy + oh * 0.985
    ds.ellipse([W / 2 - sr, sy - sr * 0.13, W / 2 + sr * 0.92, sy + sr * 0.13],
               fill=(0, 0, 0, 78))
    silhouette = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    silhouette.paste((0, 0, 0, 34), (ox + int(ow * .05), oy + int(oh * .02)),
                     obj.split()[3])
    shadow.alpha_composite(silhouette)
    im = Image.alpha_composite(im, shadow.filter(ImageFilter.GaussianBlur(9 * SS)))

    im.alpha_composite(obj, (ox, oy))

    im = im.convert("RGB").resize((width, height), Image.LANCZOS)
    im.save(path, quality=quality, subsampling=1)
    return path


# ==========================================================================
# каталог: соответствует internal/app/repository/data.go
# ==========================================================================
GROOVE = [(0.30, 0.315, "solid", 1), (0.62, 0.633, "solid", 1)]

ITEMS = [
    dict(slug="attic-black-figure", shape="kylix",
         clay=(196, 106, 62), gloss=(24, 21, 20), seed=11,
         decor=[(0.00, 0.035, "solid", 1),
                (0.07, 0.24, "figure", 7),
                (0.26, 0.30, "meander", 16),
                (0.33, 1.00, "solid", 1)]),
    dict(slug="attic-red-figure", shape="kylix",
         clay=(200, 110, 64), gloss=(26, 23, 21), seed=12,
         decor=[(0.00, 0.05, "solid", 1),
                (0.05, 0.28, "figure", 6, "invert"),
                (0.28, 1.00, "solid", 1)]),
    dict(slug="etruscan-bucchero", shape="kantharos",
         clay=(52, 49, 53), gloss=(96, 92, 96), seed=13,
         decor=[(0.06, 0.30, "tongues", 20),
                (0.34, 0.37, "solid", 1),
                (0.40, 0.52, "dots", 14)]),
    dict(slug="chian-amphora", shape="amphora_bulbous_neck",
         clay=(216, 205, 176), gloss=(166, 68, 48), seed=14,
         decor=[(0.06, 0.09, "solid", 1),
                (0.30, 0.335, "solid", 1),
                (0.40, 0.50, "rays", 22),
                (0.55, 0.575, "solid", 1),
                (0.62, 0.645, "solid", 1)]),
    dict(slug="thasian-amphora", shape="amphora_biconic",
         clay=(158, 110, 80), gloss=(74, 56, 44), seed=15,
         decor=GROOVE + [(0.44, 0.50, "stamp", 1)]),
    dict(slug="heraclean-amphora", shape="amphora_ovoid",
         clay=(170, 90, 63), gloss=(78, 50, 40), seed=16,
         decor=GROOVE + [(0.46, 0.52, "stamp", 1)]),
    dict(slug="sinopean-amphora", shape="amphora_ovoid",
         clay=(182, 78, 54), gloss=(64, 44, 38), seed=17,
         decor=GROOVE + [(0.47, 0.53, "stamp", 1)]),
    dict(slug="rhodian-amphora", shape="amphora_horned",
         clay=(206, 189, 150), gloss=(132, 106, 70), seed=18,
         decor=GROOVE + [(0.50, 0.57, "stamp", 1)]),
    dict(slug="koan-amphora", shape="amphora_double_handle",
         clay=(202, 182, 144), gloss=(128, 104, 74), seed=19,
         decor=GROOVE),
    dict(slug="megarian-bowl", shape="bowl_hemisphere",
         clay=(58, 52, 46), gloss=(124, 100, 62), seed=20,
         decor=[(0.00, 0.10, "solid", 1),
                (0.12, 0.34, "palmette", 12),
                (0.36, 0.96, "tongues", 16)]),
    dict(slug="arretine-sigillata", shape="plate",
         clay=(182, 52, 34), gloss=(138, 34, 22), seed=21,
         decor=[(0.00, 0.05, "solid", 1),
                (0.30, 0.40, "palmette", 14)]),
    dict(slug="eastern-sigillata-a", shape="bowl_incurved",
         clay=(184, 74, 50), gloss=(140, 50, 34), seed=22,
         decor=[(0.00, 0.07, "solid", 1),
                (0.44, 0.47, "solid", 1)]),
    dict(slug="naukratis-faience", shape="aryballos",
         clay=(84, 160, 158), gloss=(233, 228, 210), seed=23,
         decor=[(0.10, 0.14, "solid", 1),
                (0.30, 0.44, "zigzag", 10),
                (0.52, 0.56, "solid", 1),
                (0.62, 0.74, "dots", 9)]),
    dict(slug="phoenician-glass", shape="alabastron",
         clay=(38, 66, 126), gloss=(238, 206, 84), seed=24,
         decor=[(0.24, 0.40, "zigzag", 7),
                (0.46, 0.62, "zigzag", 7),
                (0.68, 0.82, "zigzag", 7)]),
    dict(slug="knidian-amphora", shape="amphora_biconic",
         clay=(190, 154, 116), gloss=(104, 82, 58), seed=25,
         decor=GROOVE + [(0.45, 0.51, "stamp", 1)]),
    dict(slug="mendean-amphora", shape="amphora_ovoid",
         clay=(166, 122, 90), gloss=(80, 58, 44), seed=26,
         decor=GROOVE),
]


def main():
    os.makedirs(IMG_DIR, exist_ok=True)
    os.makedirs(VID_DIR, exist_ok=True)
    have_ffmpeg = subprocess.run(["which", "ffmpeg"], capture_output=True).returncode == 0

    for it in ITEMS:
        jpg = os.path.join(IMG_DIR, it["slug"] + ".jpg")
        render(it, 1080, 1920, jpg, quality=86)
        print("img  ", os.path.relpath(jpg, ROOT))

        if not have_ffmpeg:
            continue
        mp4 = os.path.join(VID_DIR, it["slug"] + ".mp4")
        # Кадр неподвижен (никакого зума/пана — сам сосуд не двигается ни на пиксель).
        # «Живость» видео даёт только лёгкое дыхание света витрины и киношное
        # зерно — так снимают музейные экспонаты, и это не выглядит как
        # «картинка, которая приближается».
        vf = ("scale=720:1280,"
              "eq=brightness='0.018*sin(2*PI*t/5)':eval=frame,"
              "noise=alls=7:allf=t+u,"
              "format=yuv420p")
        subprocess.run(
            ["ffmpeg", "-y", "-loglevel", "error", "-loop", "1", "-i", jpg,
             "-t", "6", "-r", "25", "-vf", vf, "-c:v", "libx264", "-crf", "30",
             "-preset", "veryfast", "-movflags", "+faststart", "-an", mp4],
            check=True)
        print("video", os.path.relpath(mp4, ROOT))

    print("\nГотово: %d изображений и видео" % len(ITEMS))


if __name__ == "__main__":
    sys.exit(main())
