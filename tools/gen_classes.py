# Диаграмма классов бэкенда ЛР3: страницы фронтенда → домены API (интерфейсы)
# → модели → таблицы БД. Пишет docs/classes.svg и docs/classes.drawio.
from xml.sax.saxutils import escape

PAGE = ('#fff2cc', '#d6b656')
API = ('#dae8fc', '#6c8ebf')
MODEL = ('#d5e8d4', '#82b366')
TABLE = ('#f5f5f5', '#666666')
LH, HEAD = 17, 40          # высота строки и шапки
COLS = {'page': 20, 'api': 290, 'model': 830, 'table': 1150}
WID = {'page': 240, 'api': 510, 'model': 290, 'table': 230}

boxes = []                 # (key, col, y, stereotype, title, sections, colors)

def add(key, col, y, stereo, title, sections, colors):
    h = HEAD + sum(8 + len(s) * LH for s in sections)
    boxes.append(dict(key=key, x=COLS[col], y=y, w=WID[col], h=h, stereo=stereo,
                      title=title, sections=sections, colors=colors))
    return y + h

def box(key):
    return next(b for b in boxes if b['key'] == key)

# --- страницы фронтенда (SPA, ЛР5) и методы API, которые они вызывают
y = 60
for key, title, lines in [
    ('p_feed', 'Лента', ['/feed', 'GET  feed?id=&next=true', 'POST :id/like']),
    ('p_tiles', 'Плитка', ['/', 'GET  ?date_start=', 'DELETE :id']),
    ('p_add', 'Добавление', ['/add', 'GET  draft', 'POST (кнопка «Далее»)', 'PUT  :id/publish']),
    ('p_auth', 'Вход и регистрация', ['/login', 'POST register', 'POST login, logout']),
]:
    y = add(key, 'page', y, '«page»', title, [lines], PAGE) + 30

# --- домены API
y = add('d_cat', 'api', 60, '«interface» /api/import_categories', 'ImportCategoriesDomain', [[
    'GET    /api/import_categories?date_start=',
    'GET    /api/import_categories/feed?id=&next=true',
    'GET    /api/import_categories/draft',
    'POST   /api/import_categories      (multipart)',
    'PUT    /api/import_categories/:id/publish',
    'DELETE /api/import_categories/:id',
    'POST   /api/import_categories/:id/like',
], [
    '+GetImportCategories(c) []ImportCategorySerializer',
    '+GetFeed(c) FeedSerializer',
    '+GetDraft(c) DraftSerializer',
    '+CreateImportCategory(c) DraftSerializer',
    '+PublishImportCategory(c) ImportCategorySerializer',
    '+DeleteImportCategory(c)',
    '+LikeImportCategory(c) LikeSerializer',
]], API) + 30
y = add('d_auth', 'api', y, '«singleton» internal/app/auth', 'auth', [[
    'const creatorID uint = 1',
    '+CurrentUserID() uint   // sync.Once',
]], API) + 30
y = add('d_usr', 'api', y, '«interface» /api/users', 'UsersDomain', [[
    'POST   /api/users/register',
    'POST   /api/users/login     (заглушка до ЛР4)',
    'POST   /api/users/logout    (заглушка до ЛР4)',
], [
    '+Register(c) UserSerializer',
    '+Login(c)',
    '+Logout(c)',
]], API)

# --- модели
y = add('m_cat', 'model', 60, '«model» ds', 'ImportCategory', [[
    '+ID uint', '+Title string', '+Image *string', '+Video *string',
    '+DateStart *time.Time', '+DateEnd *time.Time', '+Description *string',
    '+Status Status', '+CreatorID uint']], MODEL) + 30
y = add('m_user', 'model', y, '«model» ds', 'User', [[
    '+ID uint', '+Login string', '+Password string', '+IsModerator bool']], MODEL) + 30
add('m_like', 'model', y, '«model» ds', 'Like', [[
    '+ID uint', '+UserID uint', '+ImportCategoryID uint']], MODEL)

# --- таблицы
add('t_cat', 'table', box('m_cat')['y'], '«table»', 'import_categories', [[
    'id PK', 'title', 'image NULL', 'video NULL', 'date_start NULL', 'date_end NULL',
    'description NULL', 'status', 'creator_id FK']], TABLE)
add('t_user', 'table', box('m_user')['y'], '«table»', 'users', [[
    'id PK', 'login', 'password', 'is_moderator']], TABLE)
add('t_like', 'table', box('m_like')['y'], '«table»', 'likes', [[
    'id PK', 'user_id FK', 'import_category_id FK']], TABLE)

# --- зависимости: (откуда, куда, доля высоты у источника, доля у цели)
deps = [
    ('p_feed', 'd_cat', .5, .25), ('p_tiles', 'd_cat', .5, .45), ('p_add', 'd_cat', .5, .65),
    ('p_auth', 'd_usr', .5, .5),
    ('d_cat', 'm_cat', .3, .4), ('d_cat', 'm_like', .75, .5), ('d_cat', 'd_auth', None, None),
    ('d_usr', 'm_user', .5, .4), ('d_auth', 'm_user', .5, .75),
    ('m_cat', 't_cat', .5, .5), ('m_user', 't_user', .5, .5), ('m_like', 't_like', .5, .5),
]

W = COLS['table'] + WID['table'] + 20
H = max(b['y'] + b['h'] for b in boxes) + 20
TITLES = [('page', 'Страницы фронтенда'), ('api', 'Домены API (internal/app/handler)'),
          ('model', 'Модели (internal/app/ds)'), ('table', 'Таблицы PostgreSQL')]

# ---------------------------------------------------------------- SVG
o = [f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}" font-family="Helvetica, Arial, sans-serif">',
     '<defs><marker id="arr" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="9" markerHeight="9" orient="auto">'
     '<path d="M0,0 L10,5 L0,10" fill="none" stroke="#333" stroke-width="1.4"/></marker></defs>',
     f'<rect width="{W}" height="{H}" fill="#fff"/>']
for col, t in TITLES:
    o.append(f'<text x="{COLS[col] + WID[col] / 2}" y="34" text-anchor="middle" font-size="14" font-weight="bold" fill="#333">{t}</text>')
for b in boxes:
    x, y, w, h = b['x'], b['y'], b['w'], b['h']
    fill, stroke = b['colors']
    o.append(f'<rect x="{x}" y="{y}" width="{w}" height="{h}" fill="#fff" stroke="{stroke}" stroke-width="1.3"/>')
    o.append(f'<rect x="{x}" y="{y}" width="{w}" height="{HEAD}" fill="{fill}" stroke="{stroke}" stroke-width="1.3"/>')
    o.append(f'<text x="{x + w / 2}" y="{y + 16}" text-anchor="middle" font-size="11" font-style="italic" fill="#444">{escape(b["stereo"])}</text>')
    o.append(f'<text x="{x + w / 2}" y="{y + 33}" text-anchor="middle" font-size="13" font-weight="bold">{escape(b["title"])}</text>')
    cy = y + HEAD
    for sec in b['sections']:
        o.append(f'<line x1="{x}" y1="{cy}" x2="{x + w}" y2="{cy}" stroke="{stroke}"/>')
        cy += 4
        for line in sec:
            cy += LH
            o.append(f'<text x="{x + 8}" y="{cy - 4}" font-size="11.5" font-family="DejaVu Sans Mono, monospace">{escape(line)}</text>')
        cy += 4

def anchor(b, side, frac):
    x = b['x'] + (b['w'] if side == 'r' else 0)
    return x, b['y'] + b['h'] * frac

for a, z, fa, fz in deps:
    A, Z = box(a), box(z)
    if fa is None:   # вертикальная зависимость внутри колонки доменов
        x1 = A['x'] + A['w'] / 2
        o.append(f'<path d="M{x1},{A["y"] + A["h"]} L{x1},{Z["y"]}" fill="none" stroke="#333" stroke-width="1.2" stroke-dasharray="6 4" marker-end="url(#arr)"/>')
        continue
    (x1, y1), (x2, y2) = anchor(A, 'r', fa), anchor(Z, 'l', fz)
    solid = a.startswith('m_')
    xm = (x1 + x2) / 2
    dash = '' if solid else ' stroke-dasharray="6 4"'
    o.append(f'<path d="M{x1},{y1} C{xm},{y1} {xm},{y2} {x2},{y2}" fill="none" stroke="#333" stroke-width="1.2"{dash} marker-end="url(#arr)"/>')
o.append('</svg>')
open('docs/classes.svg', 'w').write('\n'.join(o))

# ---------------------------------------------------------------- draw.io
c = ['<mxfile host="app.diagrams.net"><diagram name="Классы" id="classes"><mxGraphModel grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="1169" pageHeight="827" math="0" shadow="0"><root><mxCell id="0"/><mxCell id="1" parent="0"/>']
for i, (col, t) in enumerate(TITLES):
    c.append(f'<mxCell id="title{i}" value="{escape(t)}" style="text;html=1;align=center;fontStyle=1;fontSize=14;" vertex="1" parent="1">'
             f'<mxGeometry x="{COLS[col]}" y="14" width="{WID[col]}" height="30" as="geometry"/></mxCell>')
for b in boxes:
    fill, stroke = b['colors']
    label = escape(f'<i>{escape(b["stereo"])}</i><br><b>{escape(b["title"])}</b>')
    c.append(f'<mxCell id="{b["key"]}" value="{label}" style="swimlane;html=1;fontStyle=0;childLayout=stackLayout;horizontal=1;startSize={HEAD};'
             f'horizontalStack=0;resizeParent=1;resizeParentMax=0;resizeLast=0;collapsible=0;marginBottom=0;fillColor={fill};strokeColor={stroke};" vertex="1" parent="1">'
             f'<mxGeometry x="{b["x"]}" y="{b["y"]}" width="{b["w"]}" height="{b["h"]}" as="geometry"/></mxCell>')
    yy = HEAD
    for si, sec in enumerate(b['sections']):
        if si:
            c.append(f'<mxCell id="{b["key"]}_sep{si}" value="" style="line;strokeWidth=1;fillColor=none;align=left;verticalAlign=middle;'
                     f'spacingTop=-1;spacingLeft=3;spacingRight=3;rotatable=0;labelPosition=right;points=[];portConstraint=eastwest;strokeColor={stroke};" vertex="1" parent="{b["key"]}">'
                     f'<mxGeometry y="{yy}" width="{b["w"]}" height="8" as="geometry"/></mxCell>')
            yy += 8
        text = '<br>'.join(escape(l) for l in sec)
        hh = len(sec) * LH + (8 if not si else 0)
        c.append(f'<mxCell id="{b["key"]}_s{si}" value="{escape(text)}" style="text;html=1;strokeColor=none;fillColor=none;align=left;verticalAlign=top;'
                 f'spacingLeft=6;spacingTop=2;overflow=hidden;rotatable=0;fontFamily=Courier New;fontSize=11;" vertex="1" parent="{b["key"]}">'
                 f'<mxGeometry y="{yy}" width="{b["w"]}" height="{hh}" as="geometry"/></mxCell>')
        yy += hh
for i, (a, z, fa, fz) in enumerate(deps):
    solid = a.startswith('m_')
    if fa is None:
        st = 'exitX=0.5;exitY=1;exitDx=0;exitDy=0;entryX=0.5;entryY=0;entryDx=0;entryDy=0;'
    else:
        st = f'exitX=1;exitY={fa};exitDx=0;exitDy=0;entryX=0;entryY={fz};entryDx=0;entryDy=0;'
    c.append(f'<mxCell id="dep{i}" style="edgeStyle=orthogonalEdgeStyle;rounded=1;html=1;endArrow=open;endSize=10;{"" if solid else "dashed=1;"}{st}" '
             f'edge="1" parent="1" source="{a}" target="{z}"><mxGeometry relative="1" as="geometry"/></mxCell>')
c.append('</root></mxGraphModel></diagram></mxfile>')
open('docs/classes.drawio', 'w').write('\n'.join(c))
print(W, H)
