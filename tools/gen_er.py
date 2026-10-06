# ER-диаграмма ЛР2 (нотация «воронья лапка», стиль draw.io):
#   docs/er.svg — картинка для отчёта (формат под ширину листа A4)
#   docs/er.drawio — та же схема для редактирования в draw.io
RH, HH = 26, 30
HEAD, STROKE, NEXT = '#dae8fc', '#6c8ebf', '#fff2cc'
KW, NW, TW_, NUW = 34, 118, 92, 40

# поля: (ключ, имя, тип, необязательное, заполняется по «Далее»)
T = {
    'users': (20, 20, [
        ('PK', 'id', 'bigint', False, False),
        ('', 'login', 'varchar(64)', False, False),
        ('', 'password', 'varchar(128)', False, False),
        ('', 'is_moderator', 'boolean', False, False)]),
    'import_categories': (500, 20, [
        ('PK', 'id', 'bigint', False, False),
        ('', 'title', 'varchar(128)', False, True),
        ('', 'image_url', 'varchar(256)', True, True),
        ('', 'video_url', 'varchar(256)', True, True),
        ('', 'date_start', 'date', True, False),
        ('', 'date_end', 'date', True, False),
        ('', 'description', 'text', True, False),
        ('', 'status', 'varchar(16)', False, False),
        ('FK', 'creator_id', 'bigint', False, False)]),
    'likes': (262, 300, [
        ('PK', 'id', 'bigint', False, False),
        ('FK', 'user_id', 'bigint', False, False),
        ('FK', 'import_category_id', 'bigint', False, False)]),
}
W = KW + NW + TW_ + NUW
def H(n): return HH + len(T[n][2]) * RH
def ry(n, i): return T[n][1] + HH + i * RH + RH / 2
ux, uy, _ = T['users']; ix, iy, _ = T['import_categories']; lx, ly, _ = T['likes']

# связи: точки ломаной, направление выхода у начала и у конца (наружу от таблицы)
E = [
    ('e1', 'users', 0, 'import_categories', 8,
     [(ux + W, ry('users', 0)), (ix - 30, ry('users', 0)), (ix - 30, ry('import_categories', 8)), (ix, ry('import_categories', 8))],
     (1, 0), (-1, 0)),
    ('e2', 'users', None, 'likes', 1,
     [(ux + W / 2, uy + H('users')), (ux + W / 2, ry('likes', 1)), (lx, ry('likes', 1))],
     (0, 1), (-1, 0)),
    ('e3', 'import_categories', None, 'likes', 2,
     [(ix + W / 2, iy + H('import_categories')), (ix + W / 2, ry('likes', 2)), (lx + W, ry('likes', 2))],
     (0, 1), (1, 0)),
]

# ---------------- SVG ----------------
CW, CH = ix + W + 20, ly + H('likes') + 70
o = [f'<svg xmlns="http://www.w3.org/2000/svg" width="{CW}" height="{CH}" viewBox="0 0 {CW} {CH}" font-family="Helvetica, Arial, sans-serif">',
     f'<rect width="{CW}" height="{CH}" fill="#fff"/>']
for n, (x, y, rows) in T.items():
    o.append(f'<rect x="{x}" y="{y}" width="{W}" height="{H(n)}" fill="#fff" stroke="{STROKE}"/>')
    for i, (k, nm, tp, null, nxt) in enumerate(rows):
        r = y + HH + i * RH
        if nxt:
            o.append(f'<rect x="{x+0.5}" y="{r}" width="{W-1}" height="{RH}" fill="{NEXT}"/>')
        if k == 'PK':
            o.append(f'<line x1="{x}" y1="{r+RH}" x2="{x+W}" y2="{r+RH}" stroke="{STROKE}"/>')
        o.append(f'<text x="{x+KW/2}" y="{r+17}" text-anchor="middle" font-size="11" font-weight="bold">{k}</text>')
        dec = ' font-weight="bold" text-decoration="underline"' if k == 'PK' else ''
        o.append(f'<text x="{x+KW+4}" y="{r+17}" font-size="12"{dec}>{nm}</text>')
        o.append(f'<text x="{x+KW+NW+4}" y="{r+17}" font-size="11" fill="#555">{tp}</text>')
        if null:
            o.append(f'<text x="{x+W-6}" y="{r+17}" text-anchor="end" font-size="10" font-style="italic" fill="#888">NULL</text>')
    o.append(f'<rect x="{x}" y="{y}" width="{W}" height="{HH}" fill="{HEAD}" stroke="{STROKE}"/>')
    o.append(f'<text x="{x+W/2}" y="{y+20}" text-anchor="middle" font-size="13" font-weight="bold">{n}</text>')

S = 'stroke="#000" stroke-width="1" fill="none"'
def mark(kind, p, d):
    (x, y), (dx, dy) = p, d
    nx, ny = -dy, dx
    at = lambda off, side: (x + dx * off + nx * side, y + dy * off + ny * side)
    bar = lambda off: '<line x1="%s" y1="%s" x2="%s" y2="%s" %s/>' % (*at(off, 7), *at(off, -7), S)
    circ = lambda off: '<circle cx="%s" cy="%s" r="5" fill="#fff" stroke="#000"/>' % at(off, 0)
    if kind == 'one':
        return [bar(7), bar(12)]
    return ['<line x1="%s" y1="%s" x2="%s" y2="%s" %s/>' % (*at(0, 7), *at(10, 0), S),
            '<line x1="%s" y1="%s" x2="%s" y2="%s" %s/>' % (*at(0, -7), *at(10, 0), S),
            '<line x1="%s" y1="%s" x2="%s" y2="%s" %s/>' % (*at(0, 0), *at(10, 0), S), circ(15)]
for eid, s, si, t, ti, pts, d0, d1 in E:
    o.append(f'<polyline points="{" ".join(f"{a},{b}" for a, b in pts)}" {S}/>')
    o += mark('one', pts[0], d0) + mark('many', pts[-1], d1)

# легенда
ly2 = ly + H('likes') + 30
o.append(f'<rect x="20" y="{ly2-11}" width="22" height="14" fill="{NEXT}" stroke="#d6b656"/>')
o.append(f'<text x="50" y="{ly2}" font-size="12">заполняется по кнопке «Далее» (title обязательно)</text>')
o.append(f'<text x="20" y="{ly2+22}" font-size="12"><tspan font-style="italic" fill="#888">NULL</tspan> — необязательное поле; остальные поля NOT NULL</text>')
o.append('</svg>')
open('docs/er.svg', 'w').write('\n'.join(o))

# ---------------- draw.io ----------------
c = ['<mxfile host="app.diagrams.net"><diagram name="ER" id="er"><mxGraphModel grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0"><root><mxCell id="0"/><mxCell id="1" parent="0"/>']
base = 'shape=partialRectangle;connectable=0;fillColor=none;top=0;left=0;bottom=0;right=0;overflow=hidden;whiteSpace=wrap;html=1;fontSize=12;'
for n, (x, y, rows) in T.items():
    c.append(f'<mxCell id="{n}" value="{n}" style="shape=table;startSize={HH};container=1;collapsible=0;childLayout=tableLayout;fixedRows=1;rowLines=0;fontStyle=1;align=center;resizeLast=1;html=1;fillColor={HEAD};strokeColor={STROKE};" vertex="1" parent="1"><mxGeometry x="{x}" y="{y}" width="{W}" height="{H(n)}" as="geometry"/></mxCell>')
    for i, (k, nm, tp, null, nxt) in enumerate(rows):
        r = f'{n}_{i}'
        fill = NEXT if nxt else '#ffffff'
        c.append(f'<mxCell id="{r}" value="" style="shape=tableRow;horizontal=0;startSize=0;swimlaneHead=0;swimlaneBody=0;fillColor={fill};collapsible=0;dropTarget=0;points=[[0,0.5],[1,0.5]];portConstraint=eastwest;top=0;left=0;right=0;bottom={1 if k == "PK" else 0};strokeColor={STROKE};" vertex="1" parent="{n}"><mxGeometry y="{HH+i*RH}" width="{W}" height="{RH}" as="geometry"/></mxCell>')
        cells = [(k, KW, 'fontStyle=1;'), (nm, NW, 'align=left;spacingLeft=4;' + ('fontStyle=5;' if k == 'PK' else '')),
                 (tp, TW_, 'align=left;spacingLeft=4;fontColor=#555555;fontSize=11;'), ('NULL' if null else '', NUW, 'align=right;spacingRight=4;fontStyle=2;fontColor=#888888;fontSize=10;')]
        xx = 0
        for j, (v, w, st) in enumerate(cells):
            c.append(f'<mxCell id="{r}_{j}" value="{v}" style="{base}{st}" vertex="1" parent="{r}"><mxGeometry x="{xx}" width="{w}" height="{RH}" as="geometry"><mxRectangle width="{w}" height="{RH}" as="alternateBounds"/></mxGeometry></mxCell>')
            xx += w
for eid, s, si, t, ti, pts, d0, d1 in E:
    src = f'{s}_{si}' if si is not None else s
    ex, ey = ((1, 0.5) if d0 == (1, 0) else (0.5, 1))
    nx_ = 0 if d1 == (-1, 0) else 1
    mids = ''.join(f'<mxPoint x="{a}" y="{b}"/>' for a, b in pts[1:-1])
    c.append(f'<mxCell id="{eid}" style="edgeStyle=orthogonalEdgeStyle;rounded=0;html=1;startArrow=ERmandOne;endArrow=ERzeroToMany;startFill=0;endFill=0;startSize=10;endSize=10;exitX={ex};exitY={ey};exitDx=0;exitDy=0;entryX={nx_};entryY=0.5;entryDx=0;entryDy=0;" edge="1" parent="1" source="{src}" target="{t}_{ti}"><mxGeometry relative="1" as="geometry"><Array as="points">{mids}</Array></mxGeometry></mxCell>')
c.append(f'<mxCell id="lg1" value="заполняется по кнопке «Далее» (title обязательно)" style="text;html=1;align=left;verticalAlign=middle;fontSize=12;" vertex="1" parent="1"><mxGeometry x="50" y="{ly2-14}" width="420" height="20" as="geometry"/></mxCell>')
c.append(f'<mxCell id="lg0" value="" style="rounded=0;whiteSpace=wrap;html=1;fillColor={NEXT};strokeColor=#d6b656;" vertex="1" parent="1"><mxGeometry x="20" y="{ly2-11}" width="22" height="14" as="geometry"/></mxCell>')
c.append(f'<mxCell id="lg2" value="NULL — необязательное поле; остальные поля NOT NULL" style="text;html=1;align=left;verticalAlign=middle;fontSize=12;" vertex="1" parent="1"><mxGeometry x="20" y="{ly2+8}" width="460" height="20" as="geometry"/></mxCell>')
c.append('</root></mxGraphModel></diagram></mxfile>')
open('docs/er.drawio', 'w').write('\n'.join(c))
print(CW, CH)
