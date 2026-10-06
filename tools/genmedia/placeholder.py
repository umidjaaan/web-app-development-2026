#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Изображение-заглушка для категорий, добавленных через форму.

Запуск:  python3 tools/genmedia/placeholder.py
Выход:   media/img/placeholder.jpg  (1080x1920)
"""

import os

import numpy as np
from PIL import Image, ImageDraw, ImageFilter, ImageFont

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
OUT = os.path.join(ROOT, "media", "img", "placeholder.jpg")

BEIGE = (232, 232, 225)
GRAPHITE = (26, 26, 26)
SAND = (184, 184, 168)

W, H = 1080, 1920


def font(size):
    for path in ("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
                 "/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf"):
        if os.path.exists(path):
            return ImageFont.truetype(path, size)
    return ImageFont.load_default()


def paper():
    rnd = np.random.default_rng(7)
    yy = np.linspace(0.0, 1.0, H, dtype=np.float32)[:, None]
    base = np.zeros((H, W, 3), dtype=np.float32)
    base[:, :] = BEIGE
    base *= (1.0 - 0.06 * yy)[..., None]
    xx = np.linspace(-1.0, 1.0, W, dtype=np.float32)[None, :]
    vign = 1.0 - 0.10 * (xx ** 2 + (2 * yy - 1) ** 2) / 2.0
    base *= vign[..., None]
    base += rnd.normal(0.0, 3.0, (H, W, 1))
    return Image.fromarray(np.clip(base, 0, 255).astype(np.uint8), "RGB") \
        .filter(ImageFilter.GaussianBlur(0.4))


def main():
    im = paper()
    d = ImageDraw.Draw(im)

    # пунктирная рамка по центру кадра
    box = (W * 0.16, H * 0.34, W * 0.84, H * 0.66)
    dash, gap = 26, 18
    x = box[0]
    while x < box[2]:
        d.line([(x, box[1]), (min(x + dash, box[2]), box[1])], fill=SAND, width=4)
        d.line([(x, box[3]), (min(x + dash, box[2]), box[3])], fill=SAND, width=4)
        x += dash + gap
    y = box[1]
    while y < box[3]:
        d.line([(box[0], y), (box[0], min(y + dash, box[3]))], fill=SAND, width=4)
        d.line([(box[2], y), (box[2], min(y + dash, box[3]))], fill=SAND, width=4)
        y += dash + gap

    # силуэт амфоры внутри рамки
    cx, cy = W / 2, H * 0.47
    d.ellipse([cx - 90, cy - 70, cx + 90, cy + 130], outline=SAND, width=6)
    d.line([(cx - 34, cy - 150), (cx - 44, cy - 60)], fill=SAND, width=6)
    d.line([(cx + 34, cy - 150), (cx + 44, cy - 60)], fill=SAND, width=6)
    d.rectangle([cx - 36, cy - 168, cx + 36, cy - 144], outline=SAND, width=6)
    d.line([(cx - 30, cy + 128), (cx - 22, cy + 178)], fill=SAND, width=6)
    d.line([(cx + 30, cy + 128), (cx + 22, cy + 178)], fill=SAND, width=6)
    d.line([(cx - 26, cy + 178), (cx + 26, cy + 178)], fill=SAND, width=6)

    title = "Изображение не загружено"
    hint = "категория добавлена через форму"
    f1, f2 = font(46), font(30)
    d.text((cx, H * 0.71), title, font=f1, fill=GRAPHITE, anchor="mm")
    d.text((cx, H * 0.745), hint, font=f2, fill=SAND, anchor="mm")

    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    im.save(OUT, quality=86, subsampling=1)
    print("готово ->", os.path.relpath(OUT, ROOT))


if __name__ == "__main__":
    main()
