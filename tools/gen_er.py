# Генерирует ER-диаграмму (нотация «воронья лапка») в docs/er.svg
W, H = 1240, 700
RH, HH, TW = 28, 40, 300
tables = {
 "users": (40, 60, [
  ("PK","id","bigint"),("UQ","login","varchar(64)"),("","password","varchar(128)"),
  ("","full_name","varchar(128)"),("","is_moderator","boolean"),("","created_at","timestamptz")]),
 "likes": (470, 330, [
  ("PK","id","bigint"),("FK","user_id","bigint"),("FK","import_category_id","bigint"),
  ("","created_at","timestamptz")]),
 "import_categories": (900, 60, [
  ("PK","id","bigint"),("UQ","slug","varchar(64)"),("","title","varchar(128)"),
  ("","shape","varchar(64)"),("","production_center","varchar(128)"),("","region","varchar(64)"),
  ("","date_start","date"),("","date_end","date"),("","diagnostics","text"),
  ("","description","text"),("","finds_count","bigint"),("","image_url","varchar(256)"),
  ("","video_url","varchar(256)"),("","status","varchar(16)"),("FK","creator_id","bigint NULL"),
  ("","created_at","timestamptz"),("","updated_at","timestamptz")]),
}
o = [f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}" font-family="DejaVu Sans, Arial, sans-serif">',
     f'<rect width="{W}" height="{H}" fill="#ffffff"/>',
     '<text x="40" y="36" font-size="20" font-weight="bold" fill="#1a1a1a">ER-диаграмма БД «Амфора» (ЛР2)</text>']
for name,(x,y,rows) in tables.items():
    h = HH+len(rows)*RH
    o.append(f'<rect x="{x}" y="{y}" width="{TW}" height="{h}" fill="#fff" stroke="#1a1a1a" stroke-width="1.5"/>')
    o.append(f'<rect x="{x}" y="{y}" width="{TW}" height="{HH}" fill="#fecf00" stroke="#1a1a1a" stroke-width="1.5"/>')
    o.append(f'<text x="{x+TW/2}" y="{y+26}" text-anchor="middle" font-size="16" font-weight="bold" fill="#1a1a1a">{name}</text>')
    for i,(k,c,t) in enumerate(rows):
        ry = y+HH+i*RH
        if i: o.append(f'<line x1="{x}" y1="{ry}" x2="{x+TW}" y2="{ry}" stroke="#ddd"/>')
        if k=="PK" and i==0:
            o.append(f'<line x1="{x}" y1="{ry+RH}" x2="{x+TW}" y2="{ry+RH}" stroke="#1a1a1a" stroke-width="1.5"/>')
        col = {"PK":"#b8860b","FK":"#1f5fa8","UQ":"#6a6a6a"}.get(k,"#000")
        o.append(f'<text x="{x+10}" y="{ry+19}" font-size="11" font-weight="bold" fill="{col}">{k}</text>')
        u = ' text-decoration="underline"' if k=="PK" else ''
        o.append(f'<text x="{x+40}" y="{ry+19}" font-size="13" fill="#1a1a1a"{u}>{c}</text>')
        o.append(f'<text x="{x+TW-10}" y="{ry+19}" text-anchor="end" font-size="12" fill="#4c4c4c">{t}</text>')

S='stroke="#1a1a1a" stroke-width="1.6" fill="none"'
def one(x,y,d):      # «ровно один»: две черты; d=+1 линия уходит вправо
    return (f'<line x1="{x+d*10}" y1="{y-8}" x2="{x+d*10}" y2="{y+8}" {S}/>'
            f'<line x1="{x+d*16}" y1="{y-8}" x2="{x+d*16}" y2="{y+8}" {S}/>')
def zero_one(x,y,d): # «ноль или один»: черта + кружок
    return (f'<line x1="{x+d*10}" y1="{y-8}" x2="{x+d*10}" y2="{y+8}" {S}/>'
            f'<circle cx="{x+d*22}" cy="{y}" r="6" fill="#fff" stroke="#1a1a1a" stroke-width="1.6"/>')
def zero_many(x,y,d):# «ноль или много»: лапка + кружок
    return (f'<line x1="{x}" y1="{y-9}" x2="{x+d*14}" y2="{y}" {S}/>'
            f'<line x1="{x}" y1="{y+9}" x2="{x+d*14}" y2="{y}" {S}/>'
            f'<circle cx="{x+d*22}" cy="{y}" r="6" fill="#fff" stroke="#1a1a1a" stroke-width="1.6"/>')
def label(x,y,t): return f'<text x="{x}" y="{y}" text-anchor="middle" font-size="13" font-style="italic" fill="#1a1a1a">{t}</text>'

# users 0..1 ── 0..N import_categories (creator_id)
o.append(f'<line x1="340" y1="114" x2="900" y2="114" {S}/>')
o += [zero_one(340,114,1), zero_many(900,114,-1), label(620,104,"создаёт (creator_id)")]
# users 1 ── 0..N likes (user_id)
o.append(f'<polyline points="340,240 405,240 405,412 470,412" {S}/>')
o += [one(340,240,1), zero_many(470,412,-1), label(405,230,"ставит")]
# import_categories 1 ── 0..N likes (import_category_id)
o.append(f'<polyline points="900,300 835,300 835,440 770,440" {S}/>')
o += [one(900,300,-1), zero_many(770,440,1), label(835,290,"получает")]

# легенда и ограничения
lx, ly = 40, 612
o.append(f'<text x="{lx}" y="{ly}" font-size="13" font-weight="bold" fill="#1a1a1a">Ограничения</text>')
notes = ["likes: UNIQUE (user_id, import_category_id) — один лайк от пользователя на категорию",
         "import_categories: UNIQUE INDEX (creator_id) WHERE status = 'draft' — один черновик на пользователя",
         "import_categories.status ∈ {draft, published, deleted}; удаление — логическое (SQL UPDATE через курсор)"]
for i,n in enumerate(notes):
    o.append(f'<text x="{lx}" y="{ly+22+i*20}" font-size="12" fill="#4c4c4c">{n}</text>')
o.append('</svg>')
open("docs/er.svg","w").write("\n".join(o))
