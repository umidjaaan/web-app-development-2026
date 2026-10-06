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
        return [R(x, y, tw, 30, fill=YELLOW, rx=15),
                T(x + tw / 2, y + 20, text, size, INK, 600, anchor="middle")]
    return [R(x, y, tw, 26, fill="#111110", rx=13, stroke=PAPER, sw=1, op=0.9),
            R(x, y, tw, 26, fill="none", rx=13, stroke=PAPER, sw=1, op=0.55),
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
         '<linearGradient id="infoScrim" x1="0" y1="1" x2="0" y2="0">',
         '<stop offset="0" stop-color="#000" stop-opacity="0.82"/>',
         '<stop offset="0.55" stop-color="#000" stop-opacity="0.72"/>',
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
         T(W - 18, 34, "3 / 15", 13, PAPER, 400, anchor="end", op=0.75),
         '</g>',
         '<g id="Info">',
         R(0, 494, W, 258, fill="url(#infoScrim)")]
    b += pill(18, 520, "Аттика")
    b += [T(18, 572, "Аттическая", 26, PAPER, 700),
          T(18, 600, "чернофигурная", 26, PAPER, 700),
          T(18, 628, "керамика", 26, PAPER, 700),
          T(18, 654, "Афины, квартал Керамик", 15, PAPER, 500),
          T(18, 676, "620–480 гг. до н. э. · Килик, амфора, лекиф", 13, PAPER,
            400, op=0.72),
          '<g id="Description">',
          T(18, 702, "Массовый показатель прямых связей с", 14, PAPER, 400, op=0.9),
          T(18, 722, "Афинами в архаический период. Столов…", 14, PAPER, 400, op=0.9),
          T(18, 748, "Ещё", 13, YELLOW, 600),
          '</g>',
          '</g>']

    # правый рельс: параметры категории и обе иконки
    cx = 344
    b += ['<g id="Rail">',
          T(cx, 544, "НАХОДОК", 10, PAPER, 500, ls="0.6", anchor="middle", op=0.6),
          T(cx, 564, "214", 16, PAPER, 700, anchor="middle"),
          T(cx, 592, "ДОЛЯ", 10, PAPER, 500, ls="0.6", anchor="middle", op=0.6),
          T(cx, 612, "15.9%", 16, PAPER, 700, anchor="middle"),
          '<g id="Icon-Like">',
          f'<path d="M{cx} {cx*0+664} L{cx-15} 650 a8.6 8.6 0 0 1 12.2 -12.2 '
          f'l2.8 2.8 l2.8 -2.8 a8.6 8.6 0 0 1 12.2 12.2 Z" fill="{PAPER}"/>',
          T(cx, 682, "14", 11, PAPER, 600, anchor="middle"),
          '</g>',
          '<g id="Icon-Next">',
          f'<circle cx="{cx}" cy="717" r="16" fill="none" stroke="{PAPER}" stroke-width="1.6"/>',
          f'<path d="M{cx} 709 L{cx} 725 M{cx-6} 719 L{cx} 725 L{cx+6} 719" '
          f'fill="none" stroke="{PAPER}" stroke-width="1.6" '
          f'stroke-linecap="round" stroke-linejoin="round"/>',
          T(cx, 748, "След.", 11, PAPER, 600, anchor="middle"),
          '</g>',
          '</g>']
    b += tabbar(H - 74, "feed")
    svg("screen-1-feed.svg", W, H, b, "Лента")


# ------------------------------------------------------------ 2. Добавление
def field(y, label, value, ph=False, h=44, w=326, x=32, icon=None):
    """Подпись + поле ввода. value рисуется как placeholder, если ph=True."""
    col = SAND if ph else PAPER
    op = 0.5 if ph else None
    out = [T(x, y, label, 12, SAND),
           R(x, y + 10, w, h, fill=PAPER, rx=10, op=0.06),
           R(x, y + 10, w, h, fill="none", rx=10, stroke=PAPER, sw=1, op=0.22),
           T(x + 14, y + 38, value, 15, col, op=op)]
    if icon == "calendar":
        ix = x + w - 36
        out += [R(ix, y + 22, 20, 20, fill="none", rx=3, stroke=SAND, sw=1.4),
                LINE(ix, y + 28, ix + 20, y + 28, SAND, 1.4),
                LINE(ix + 6, y + 19, ix + 6, y + 24, SAND, 1.4),
                LINE(ix + 14, y + 19, ix + 14, y + 24, SAND, 1.4)]
    if icon == "select":
        out.append(f'<path d="M{x + w - 22} {y + 28} l6 8 l6 -8" fill="none" '
                   f'stroke="{SAND}" stroke-width="1.6"/>')
    return out


def screen_add():
    W, H = 390, 1580
    b = topbar()
    b.append('<g id="Content">')
    b += [T(16, 96, "Новая категория импорта", 26, PAPER, 700),
          T(16, 124, "Форма заполняется вручную: изображение и видео уходят", 13, SAND),
          T(16, 143, "в Minio, остальные поля — в коллекцию import_categories.", 13, SAND)]

    def fieldset(y, h, legend, gid):
        return [f'<g id="{gid}">',
                R(14, y, 362, h, fill=PAPER, rx=20, stroke=PAPER, sw=1, op=0.04),
                R(14, y, 362, h, fill="none", rx=20, stroke=PAPER, sw=1, op=0.16),
                R(28, y - 8, len(legend) * 7.2 + 16, 16, fill=SHELL),
                T(36, y + 4, legend, 11, YELLOW, 600, ls="0.88"),
                '</g>']

    # Изображение
    b += fieldset(170, 116, "ИЗОБРАЖЕНИЕ", "Fieldset-Image")
    b += [R(32, 196, 140, 32, fill=PAPER, rx=16),
          T(102, 217, "Выберите файл", 12, INK, 600, anchor="middle"),
          T(184, 217, "Файл не выбран", 12, SAND),
          T(32, 258, "JPEG, вертикальный кадр 9:16 — объект бакета в Minio.", 11,
            SAND, op=0.75)]

    # Видео
    b += fieldset(306, 116, "ВИДЕО", "Fieldset-Video")
    b += [R(32, 332, 140, 32, fill=PAPER, rx=16),
          T(102, 353, "Выберите файл", 12, INK, 600, anchor="middle"),
          T(184, 353, "Файл не выбран", 12, SAND),
          T(32, 394, "MP4 без звука, зацикленное — объект бакета в Minio.", 11,
            SAND, op=0.75)]

    # Описание
    b += fieldset(442, 460, "ОПИСАНИЕ", "Fieldset-Text")
    b += field(476, "Название категории", "Например: Книдские амфоры", ph=True)
    b += field(554, "Центр производства", "Город, область", ph=True)
    b += [T(32, 632, "Признаки-маркёры", 12, SAND),
          R(32, 642, 326, 100, fill=PAPER, rx=10, op=0.06),
          R(32, 642, 326, 100, fill="none", rx=10, stroke=PAPER, sw=1, op=0.22),
          T(46, 670, "По каким признакам категория", 15, SAND, op=0.5),
          T(46, 692, "опознаётся во фрагменте", 15, SAND, op=0.5),
          T(32, 766, "Справка", 12, SAND),
          R(32, 776, 326, 100, fill=PAPER, rx=10, op=0.06),
          R(32, 776, 326, 100, fill="none", rx=10, stroke=PAPER, sw=1, op=0.22),
          T(46, 804, "Значение категории для", 15, SAND, op=0.5),
          T(46, 826, "реконструкции связей", 15, SAND, op=0.5)]

    # Параметры сводки
    b += fieldset(930, 460, "ПАРАМЕТРЫ СВОДКИ", "Fieldset-Params")
    b += field(964, "Регион-поставщик", "Аттика", icon="select")
    b += field(1042, "Морфологический тип", "Амфора, килик, чаша", ph=True)
    b += field(1120, "Дата начала бытования типа", "дд.мм.гггг", ph=True,
               icon="calendar")
    b += field(1198, "Дата конца бытования типа", "дд.мм.гггг", ph=True,
               icon="calendar")
    b += [T(32, 1292, "Год в календаре — это год до нашей эры: 0250-01-01", 11,
            SAND, op=0.75),
          T(32, 1308, "читается как 250 г. до н. э.", 11, SAND, op=0.75)]
    b += field(1332, "Число фрагментов на памятнике", "0", ph=True)

    b += [R(14, 1428, 362, 48, fill=YELLOW, rx=24),
          T(195, 1458, "Сохранить категорию", 15, INK, 700, anchor="middle"),
          T(14, 1488, "На этом этапе форма ничего не сохраняет: запись", 11, SAND, op=0.75),
          T(14, 1504, "в базу появится в лабораторной работе №3.", 11, SAND, op=0.75),
          '</g>']
    b += tabbar(H - 74, "add")
    svg("screen-2-add.svg", W, H, b, "Добавление")


# --------------------------------------------------------------- 3. Плитка
SHARES = [("Аттика", "28.4%", 382, 2, 1.000),
          ("Южное Причерноморье", "24.7%", 332, 2, 0.869),
          ("Восточная Эгеида", "18.8%", 252, 4, 0.660),
          ("Северная Эгеида", "9.8%", 132, 1, 0.346),
          ("Левант", "5.7%", 77, 2, 0.202),
          ("Египет", "4.6%", 62, 1, 0.162),
          ("Этрурия", "3.5%", 47, 1, 0.123),
          ("Малая Азия", "2.5%", 33, 1, 0.086),
          ("Италия", "1.9%", 26, 1, 0.068)]

CARDS = [("etruscan-bucchero", ["Этрусское буккеро", "(bucchero nero)"],
          ["Цере (Черветери) и Вульчи,", "Южная Этрурия"], "Этрурия", 9,
          "675–500 до н. э."),
         ("naukratis-faience", ["Египетский фаянс:", "скарабеи и бусы"],
          ["Навкратис, дельта Нила"], "Египет", 5, "650–525 до н. э."),
         ("attic-black-figure", ["Аттическая", "чернофигурная", "керамика"],
          ["Афины, квартал Керамик"], "Аттика", 14, "620–480 до н. э."),
         ("chian-amphora", ["Хиосские амфоры"],
          ["Остров Хиос"], "В. Эгеида", 10, "600–300 до н. э.")]


def screen_catalog():
    W, H = 390, 1700
    b = topbar()
    b.append('<g id="Content">')
    b += [T(14, 96, "Категории импорта", 26, PAPER, 700),
          T(14, 124, "Учтено 1343 фрагментов импортной посуды и утвари.", 13, SAND),
          T(14, 143, "Удельный вес региона в этой массе — исходная величина", 13, SAND),
          T(14, 162, "для реконструкции торговых путей.", 13, SAND)]

    # блок долей регионов
    top = 182
    b += ['<g id="RegionShares">',
          R(14, top, 362, 518, fill=PAPER, rx=20),
          T(28, top + 28, "УДЕЛЬНЫЙ ВЕС РЕГИОНОВ", 12, MUTED, 600, ls="0.96")]
    y = top + 56
    for name, pct, finds, cats, frac in SHARES:
        colour = YELLOW if frac == 1.000 else INK
        b += [T(28, y, name, 14, INK, 700),
              T(362, y, pct, 14, INK, 700, anchor="end"),
              R(28, y + 8, 334, 8, fill=INK, rx=4, op=0.1),
              R(28, y + 8, round(334 * frac, 1), 8, fill=colour, rx=4),
              T(28, y + 32, f"{finds} фр. · категорий: {cats}", 11, MUTED)]
        y += 50
    b.append('</g>')

    # фильтр: календарь с кнопкой
    b += ['<g id="Filter">',
          T(14, 738, "ДАТА НАЧАЛА — НЕ ПОЗДНЕЕ", 12, SAND, 400, ls="0.48"),
          R(14, 750, 240, 44, fill=PAPER, rx=10, op=0.06),
          R(14, 750, 240, 44, fill="none", rx=10, stroke=PAPER, sw=1, op=0.22),
          T(28, 778, "01.01.0500", 15, PAPER),
          R(218, 762, 20, 20, fill="none", rx=3, stroke=SAND, sw=1.4),
          LINE(218, 768, 238, 768, SAND, 1.4),
          LINE(224, 759, 224, 764, SAND, 1.4),
          LINE(232, 759, 232, 764, SAND, 1.4),
          R(262, 750, 114, 44, fill=PAPER, rx=10),
          T(319, 778, "Показать", 15, INK, 600, anchor="middle"),
          T(14, 816, "Год в календаре — это год до нашей эры: 0500-01-01", 11,
            SAND, op=0.75),
          T(14, 832, "читается как 500 г. до н. э.", 11, SAND, op=0.75),
          T(14, 862, "Найдено категорий: 6", 12, SAND),
          T(376, 862, "сбросить фильтр", 12, YELLOW, 600, anchor="end"),
          '</g>']

    # сетка карточек: две колонки
    b.append('<g id="Grid">')
    defs = ['<defs>']
    for i, (slug, lines, center, region, likes, period) in enumerate(CARDS):
        col, row = i % 2, i // 2
        x = 14 + col * 187
        y = 884 + row * 344
        cid = f"card{i}"
        defs.append(f'<clipPath id="{cid}"><rect x="{x}" y="{y}" '
                    f'width="175" height="219" rx="20"/></clipPath>')
        b += [R(x, y, 175, 322, fill=PAPER, rx=20),
              f'<g clip-path="url(#{cid})">{IMG(slug, x, y, 175, 219)}</g>',
              R(x, y + 199, 175, 20, fill=PAPER)]
        b += pill(x + 10, y + 10, region, size=11)
        ty = y + 234
        for ln in lines:
            b.append(T(x + 12, ty, ln, 14, INK, 700))
            ty += 17
        ty += 2
        for ln in center:
            b.append(T(x + 12, ty, ln, 11, MUTED))
            ty += 14
        fy = y + 308
        b += [LINE(x + 12, fy - 14, x + 163, fy - 14, INK, 1, op=0.12),
              T(x + 12, fy, f"♡ {likes}", 12, INK, 600),
              T(x + 163, fy, period, 11, MUTED, anchor="end")]
    defs.append('</defs>')
    b = defs + b
    b.append('</g>')

    b += [T(14, 1594, "Количества фрагментов — учебная сводка по материалам", 11,
            SAND, op=0.7),
          T(14, 1610, "памятника.", 11, SAND, op=0.7),
          '</g>']
    b += tabbar(H - 74, "catalog")
    svg("screen-3-catalog.svg", W, H, b, "Плитка")


screen_feed()
screen_add()
screen_catalog()
print("готово ->", OUT)
