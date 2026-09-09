#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Три экрана приложения «Амфора» как SVG для импорта в Figma."""

import base64
import io
import os
from PIL import Image

ROOT = "/home/claude/rip-lab1"
OUT = os.path.join(ROOT, "docs", "figma")
os.makedirs(OUT, exist_ok=True)

INK = "#1A1A1A"
SHELL = "#111110"
PAPER = "#E8E8E1"
SAND = "#B8B8A8"
MUTED = "#4C4C4C"
YELLOW = "#FECF00"
FONT = "Inter, 'Helvetica Neue', Arial, sans-serif"


def esc(s):
    return (s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;"))


def T(x, y, s, size=14, fill=PAPER, weight=400, anchor="start", ls=None, op=None):
    a = f' text-anchor="{anchor}"' if anchor != "start" else ""
    l = f' letter-spacing="{ls}"' if ls else ""
    o = f' opacity="{op}"' if op else ""
    return (f'<text x="{x}" y="{y}" font-family="{FONT}" font-size="{size}" '
            f'font-weight="{weight}" fill="{fill}"{a}{l}{o}>{esc(s)}</text>')


def R(x, y, w, h, fill="none", rx=0, stroke=None, sw=1, op=None):
    s = f' stroke="{stroke}" stroke-width="{sw}"' if stroke else ""
    o = f' opacity="{op}"' if op else ""
    return f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="{rx}" fill="{fill}"{s}{o}/>'


def LINE(x1, y1, x2, y2, stroke, sw=1, op=None):
    o = f' opacity="{op}"' if op else ""
    return (f'<line x1="{x1}" y1="{y1}" x2="{x2}" y2="{y2}" '
            f'stroke="{stroke}" stroke-width="{sw}"{o}/>')


def img_data(slug, box_w, box_h, mode="cover45"):
    """Кроп исходного кадра 1080x1920 под нужную рамку -> data URI."""
    im = Image.open(os.path.join(ROOT, "media", "img", slug + ".jpg"))
    W, H = im.size
    target = box_w / box_h
    src = W / H
    if src > target:                      # шире — режем по бокам
        nw = int(H * target)
        im = im.crop(((W - nw) // 2, 0, (W - nw) // 2 + nw, H))
    else:                                 # выше — режем сверху и снизу
        nh = int(W / target)
        im = im.crop((0, (H - nh) // 2, W, (H - nh) // 2 + nh))
    im = im.resize((int(box_w * 2), int(box_h * 2)), Image.LANCZOS)
    buf = io.BytesIO()
    im.save(buf, "JPEG", quality=78, subsampling=1)
    b64 = base64.b64encode(buf.getvalue()).decode()
    return "data:image/jpeg;base64," + b64


def IMG(slug, x, y, w, h, clip=None):
    c = f' clip-path="url(#{clip})"' if clip else ""
    return (f'<image x="{x}" y="{y}" width="{w}" height="{h}" '
            f'preserveAspectRatio="xMidYMid slice"{c} '
            f'href="{img_data(slug, w, h)}"/>')


def tabbar(y, active, w=390):
    """Нижняя навигационная панель, 74 px."""
    o = [f'<g id="Tabbar">',
         R(0, y, w, 74, fill=SHELL),
         LINE(0, y, w, y, PAPER, 1, op=0.14)]
    tabs = [("Лента", "feed"), ("Добавление", "add"), ("Плитка", "catalog")]
    for i, (label, key) in enumerate(tabs):
        cx = w / 6 + i * w / 3
        col = YELLOW if key == active else SAND
        iy = y + 18
        if key == "feed":
            o.append(R(cx - 11, iy, 22, 22, rx=5, stroke=col, sw=1.6))
            o.append(f'<path d="M{cx-3.5} {iy+6.5} L{cx-3.5} {iy+15.5} '
                     f'L{cx+5.5} {iy+11} Z" fill="{col}"/>')
        elif key == "add":
            o.append(R(cx - 11, iy, 22, 22, rx=6, stroke=col, sw=1.6))
            o.append(LINE(cx, iy + 5, cx, iy + 17, col, 1.7))
            o.append(LINE(cx - 6, iy + 11, cx + 6, iy + 11, col, 1.7))
        else:
            for dx, dy in ((-11, 0), (1, 0), (-11, 12), (1, 12)):
                o.append(R(cx + dx, iy + dy, 10, 10, rx=2.5, stroke=col, sw=1.6))
        o.append(T(cx, y + 58, label, 11, col, 500, anchor="middle"))
    o.append("</g>")
    return o


def topbar(site="Елизаветовское городище, дельта Дона", w=390):
    return [f'<g id="Topbar">',
            T(16, 34, "АМФОРА", 17, PAPER, 700, ls="2.4"),
            T(w - 16, 33, site, 12, SAND, 400, anchor="end"),
            LINE(0, 50, w, 50, PAPER, 1, op=0.12),
            "</g>"]


def pill(x, y, text, w=None, size=13, filled=False):
    tw = w or (len(text) * (size * 0.56) + 32)
    if filled:
        return [R(x, y, tw, 30, fill=YELLOW, rx=999),
                T(x + tw / 2, y + 20, text, size, INK, 600, anchor="middle")]
    return [R(x, y, tw, 26, fill="#111110", rx=999, stroke=PAPER, sw=1, op=0.9),
            R(x, y, tw, 26, fill="none", rx=999, stroke=PAPER, sw=1, op=0.55),
            T(x + tw / 2, y + 17.5, text, size, PAPER, 500, anchor="middle")]


def svg(name, w, h, body, title):
    doc = [f'<svg xmlns="http://www.w3.org/2000/svg" '
           f'xmlns:xlink="http://www.w3.org/1999/xlink" '
           f'width="{w}" height="{h}" viewBox="0 0 {w} {h}">',
           f'<title>{title}</title>',
           R(0, 0, w, h, fill=SHELL)]
    doc += body
    doc.append("</svg>")
    path = os.path.join(OUT, name)
    with open(path, "w", encoding="utf-8") as f:
        f.write("\n".join(doc))
    print(name, os.path.getsize(path) // 1024, "KB")


# ---------------------------------------------------------------- 1. Лента
def screen_feed():
    W, H = 390, 844
    b = ['<defs>',
         '<linearGradient id="shadeTop" x1="0" y1="0" x2="0" y2="1">',
         '<stop offset="0" stop-color="#000" stop-opacity="0.55"/>',
         '<stop offset="1" stop-color="#000" stop-opacity="0"/>',
         '</linearGradient>',
         '<linearGradient id="shadeBottom" x1="0" y1="1" x2="0" y2="0">',
         '<stop offset="0" stop-color="#000" stop-opacity="0.88"/>',
         '<stop offset="0.45" stop-color="#000" stop-opacity="0.55"/>',
         '<stop offset="1" stop-color="#000" stop-opacity="0"/>',
         '</linearGradient>',
         '</defs>',
         '<g id="Video">',
         IMG("attic-black-figure", 0, 0, W, H),
         R(0, 0, W, 220, fill="url(#shadeTop)"),
         R(0, 320, W, 524, fill="url(#shadeBottom)"),
         '</g>',
         '<g id="Header">',
         T(18, 34, "АМФОРА", 17, PAPER, 700, ls="2.4"),
         T(W - 18, 34, "1 / 14", 13, PAPER, 400, anchor="end", op=0.75),
         '</g>',
         '<g id="Info">']
    b += pill(18, 418, "Аттика")
    b += [T(18, 472, "Аттическая", 28, PAPER, 700),
          T(18, 502, "чернофигурная керамика", 28, PAPER, 700),
          T(18, 530, "Афины, квартал Керамик", 15, PAPER, 500),
          T(18, 552, "620–480 гг. до н. э. · Килик, амфора, лекиф", 13, PAPER,
            400, op=0.72),
          T(18, 580, "Чёрный блестящий лак силуэтом по оранжевой", 14, PAPER,
            400, op=0.9),
          T(18, 600, "глине; детали процарапаны иглой до глины,", 14, PAPER,
            400, op=0.9),
          T(18, 620, "добавлены пурпур и белила.", 14, PAPER, 400, op=0.9),
          LINE(18, 646, W - 18, 646, PAPER, 1, op=0.22),
          '<g id="Stats">',
          T(18, 668, "НАХОДОК", 11, PAPER, 500, ls="0.66", op=0.6),
          T(18, 690, "214", 19, PAPER, 700),
          T(110, 668, "ДОЛЯ ИМПОРТА", 11, PAPER, 500, ls="0.66", op=0.6),
          T(110, 690, "16.4%", 19, PAPER, 700),
          T(230, 668, "ЛАЙКОВ", 11, PAPER, 500, ls="0.66", op=0.6),
          T(230, 690, "♥ 14", 19, PAPER, 700),
          '</g>',
          R(18, 706, 168, 46, fill=PAPER, rx=999),
          T(102, 735, "Следующий →", 15, INK, 600, anchor="middle"),
          '</g>']
    b += tabbar(H - 74, "feed")
    svg("screen-1-feed.svg", W, H, b, "Лента")


# ------------------------------------------------------------ 2. Добавление
def screen_add():
    W, H = 390, 1420
    b = topbar()
    b.append('<g id="Content">')
    b.append(T(16, 96, "Новая категория импорта", 26, PAPER, 700))
    b += pill(16, 112, "Черновик", w=104, filled=True)
    b.append(T(132, 133, "черновик №15 загружен с сервера", 12, SAND))

    def fieldset(y, h, legend, gid):
        return [f'<g id="{gid}">',
                R(14, y, 362, h, fill=PAPER, rx=20, stroke=PAPER, sw=1, op=0.04),
                R(14, y, 362, h, fill="none", rx=20, stroke=PAPER, sw=1, op=0.16),
                R(28, y - 8, len(legend) * 7.2 + 16, 16, fill=SHELL),
                T(36, y + 4, legend, 11, YELLOW, 600, ls="0.88"),
                '</g>']

    # Изображение
    b += fieldset(168, 172, "ИЗОБРАЖЕНИЕ", "Fieldset-Image")
    b += ['<defs><clipPath id="thumb1"><rect x="32" y="190" width="76" height="95" rx="10"/></clipPath></defs>',
          f'<g clip-path="url(#thumb1)">{IMG("knidian-amphora", 32, 190, 76, 95)}</g>',
          T(122, 240, "knidian-amphora.jpg —", 11, SAND),
          T(122, 256, "объект в Minio", 11, SAND),
          R(32, 300, 140, 32, fill=PAPER, rx=999),
          T(102, 321, "Выберите файл", 12, INK, 600, anchor="middle"),
          T(184, 321, "Файл не выбран", 12, SAND)]

    # Видео
    b += fieldset(360, 172, "ВИДЕО", "Fieldset-Video")
    b += ['<defs><clipPath id="thumb2"><rect x="32" y="382" width="76" height="95" rx="10"/></clipPath></defs>',
          f'<g clip-path="url(#thumb2)">{IMG("knidian-amphora", 32, 382, 76, 95)}</g>',
          T(122, 432, "knidian-amphora.mp4 —", 11, SAND),
          T(122, 448, "объект в Minio", 11, SAND),
          R(32, 492, 140, 32, fill=PAPER, rx=999),
          T(102, 513, "Выберите файл", 12, INK, 600, anchor="middle"),
          T(184, 513, "Файл не выбран", 12, SAND)]

    # Описание
    b += fieldset(552, 336, "ОПИСАНИЕ", "Fieldset-Text")
    fields = [(586, "Название категории", "Книдские амфоры"),
              (664, "Центр производства", "Книд, Кария")]
    for y, label, value in fields:
        b += [T(32, y, label, 12, SAND),
              R(32, y + 10, 326, 44, fill=PAPER, rx=10, op=0.06),
              R(32, y + 10, 326, 44, fill="none", rx=10, stroke=PAPER, sw=1, op=0.22),
              T(46, y + 38, value, 15, PAPER)]
    b += [T(32, 742, "Признаки-маркёры", 12, SAND),
          R(32, 752, 326, 116, fill=PAPER, rx=10, op=0.06),
          R(32, 752, 326, 116, fill="none", rx=10, stroke=PAPER, sw=1, op=0.22),
          T(46, 778, "Ручки с «книдским» коленчатым", 15, PAPER),
          T(46, 800, "изломом; клейма с именем", 15, PAPER),
          T(46, 822, "фрурарха и эмблемой —", 15, PAPER),
          T(46, 844, "бычьей головой.", 15, PAPER)]

    # Параметры сводки
    b += fieldset(912, 316, "ПАРАМЕТРЫ СВОДКИ", "Fieldset-Params")
    b += [T(32, 946, "Регион-поставщик", 12, SAND),
          R(32, 956, 326, 44, fill=PAPER, rx=10, op=0.06),
          R(32, 956, 326, 44, fill="none", rx=10, stroke=PAPER, sw=1, op=0.22),
          T(46, 984, "Восточная Эгеида", 15, PAPER),
          f'<path d="M336 976 l6 8 l6 -8" fill="none" stroke="{SAND}" stroke-width="1.6"/>']
    for y, label, value in [(1024, "Морфологический тип", "Тарная амфора"),
                            (1102, "Датировка", "III–I вв. до н. э."),
                            (1180, "Число фрагментов на памятнике", "0")]:
        b += [T(32, y, label, 12, SAND),
              R(32, y + 10, 326, 44, fill=PAPER, rx=10, op=0.06),
              R(32, y + 10, 326, 44, fill="none", rx=10, stroke=PAPER, sw=1, op=0.22),
              T(46, y + 38, value, 15, PAPER)]

    b += [R(14, 1256, 362, 48, fill=YELLOW, rx=999),
          T(195, 1286, "Сохранить категорию", 15, INK, 700, anchor="middle"),
          T(14, 1318, "На этом этапе форма ничего не сохраняет: запись", 11, SAND, op=0.75),
          T(14, 1334, "в базу появится в лабораторной работе №3.", 11, SAND, op=0.75),
          '</g>']
    b += tabbar(H - 74, "add")
    svg("screen-2-add.svg", W, H, b, "Добавление")


# --------------------------------------------------------------- 3. Плитка
SHARES = [("Аттика", "29.2%", 382, 2, 1.00),
          ("Южное Причерноморье", "25.4%", 332, 2, 0.869),
          ("Восточная Эгеида", "16.5%", 215, 3, 0.563),
          ("Северная Эгеида", "10.1%", 132, 1, 0.346),
          ("Левант", "5.9%", 77, 2, 0.202),
          ("Египет", "4.7%", 62, 1, 0.162),
          ("Этрурия", "3.6%", 47, 1, 0.123),
          ("Малая Азия", "2.5%", 33, 1, 0.086),
          ("Италия", "2.0%", 26, 1, 0.068)]

CARDS = [("attic-black-figure", ["Аттическая", "чернофигурная", "керамика"],
          "Афины, квартал Керамик", "Аттика", 14, "214 фр."),
         ("heraclean-amphora", ["Гераклейские", "амфоры"],
          "Гераклея Понтийская", "Ю. Причерноморье", 11, "187 фр."),
         ("attic-red-figure", ["Аттическая", "краснофигурная", "керамика"],
          "Афины, квартал Керамик", "Аттика", 12, "168 фр."),
         ("sinopean-amphora", ["Синопские", "амфоры"],
          "Синопа", "Ю. Причерноморье", 9, "145 фр.")]


def screen_catalog():
    W, H = 390, 1660
    b = topbar()
    b.append('<g id="Content">')
    b += [T(14, 96, "Категории импорта", 26, PAPER, 700),
          T(14, 124, "Учтено 1306 фрагментов импортной посуды и утвари.", 13, SAND),
          T(14, 143, "Удельный вес региона в этой массе — исходная величина", 13, SAND),
          T(14, 162, "для реконструкции торговых путей.", 13, SAND)]

    # блок долей
    top = 182
    b += ['<g id="RegionShares">',
          R(14, top, 362, 518, fill=PAPER, rx=20),
          T(28, top + 28, "УДЕЛЬНЫЙ ВЕС РЕГИОНОВ", 12, MUTED, 600, ls="0.96")]
    y = top + 56
    for name, pct, finds, cats, frac in SHARES:
        colour = YELLOW if frac == 1.00 else INK
        b += [T(28, y, name, 14, INK, 700),
              T(362, y, pct, 14, INK, 700, anchor="end"),
              R(28, y + 8, 334, 8, fill=INK, rx=999, op=0.1),
              R(28, y + 8, round(334 * frac, 1), 8, fill=colour, rx=999),
              T(28, y + 32, f"{finds} фр. · категорий: {cats}", 11, MUTED)]
        y += 50
    b.append('</g>')

    # фильтр
    b += ['<g id="Filter">',
          T(14, 738, "НЕ МЕНЕЕ, ФРАГМЕНТОВ", 12, SAND, 400, ls="0.48"),
          R(14, 750, 240, 44, fill=PAPER, rx=10, op=0.06),
          R(14, 750, 240, 44, fill="none", rx=10, stroke=PAPER, sw=1, op=0.22),
          T(28, 778, "0", 15, SAND, op=0.5),
          R(262, 750, 114, 44, fill=PAPER, rx=10),
          T(319, 778, "Показать", 15, INK, 600, anchor="middle"),
          T(14, 820, "Найдено категорий: 14", 12, SAND),
          '</g>']

    # сетка карточек 2 колонки
    b.append('<g id="Grid">')
    defs = ['<defs>']
    for i, (slug, lines, center, region, likes, finds) in enumerate(CARDS):
        col, row = i % 2, i // 2
        x = 14 + col * 187
        y = 838 + row * 344
        cid = f"card{i}"
        defs.append(f'<clipPath id="{cid}"><rect x="{x}" y="{y}" '
                    f'width="175" height="219" rx="20"/></clipPath>')
        b += [R(x, y, 175, 322, fill=PAPER, rx=20),
              f'<g clip-path="url(#{cid})">{IMG(slug, x, y, 175, 219)}</g>',
              R(x, y + 199, 175, 20, fill=PAPER)]
        b += pill(x + 10, y + 10, region, size=11)
        ty = y + 240
        for ln in lines:
            b.append(T(x + 12, ty, ln, 14, INK, 700))
            ty += 17
        b.append(T(x + 12, ty + 4, center, 11, MUTED))
        fy = y + 308
        b += [LINE(x + 12, fy - 14, x + 163, fy - 14, INK, 1, op=0.12),
              T(x + 12, fy, f"♡ {likes}", 12, INK, 600),
              T(x + 163, fy, finds, 12, MUTED, anchor="end")]
    defs.append('</defs>')
    b = defs + b
    b.append('</g>')

    b += [T(14, 1554, "Количества фрагментов — учебная сводка по материалам", 11, SAND, op=0.7),
          T(14, 1570, "памятника.", 11, SAND, op=0.7),
          '</g>']
    b += tabbar(H - 74, "catalog")
    svg("screen-3-catalog.svg", W, H, b, "Плитка")


screen_feed()
screen_add()
screen_catalog()
print("готово ->", OUT)
