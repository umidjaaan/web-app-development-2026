# Схема архитектуры приложения «Амфора» → docs/architecture.svg
W,H=1240,720
o=[f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}" font-family="DejaVu Sans, Arial, sans-serif">',
 '<defs><marker id="a" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="8" markerHeight="8" orient="auto-start-reverse"><path d="M0,0 L10,5 L0,10 z" fill="#1a1a1a"/></marker></defs>',
 f'<rect width="{W}" height="{H}" fill="#fff"/>',
 '<text x="40" y="36" font-size="20" font-weight="bold" fill="#1a1a1a">Архитектура приложения «Амфора» (ЛР2)</text>']
def box(x,y,w,h,title,sub=(),fill="#fff",bold=True,ts=15):
    o.append(f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="10" fill="{fill}" stroke="#1a1a1a" stroke-width="1.5"/>')
    n=1+len(sub); cy=y+h/2-(n-1)*9+5
    o.append(f'<text x="{x+w/2}" y="{cy}" text-anchor="middle" font-size="{ts}" font-weight="{"bold" if bold else "normal"}" fill="#1a1a1a">{title}</text>')
    for i,s in enumerate(sub):
        o.append(f'<text x="{x+w/2}" y="{cy+19*(i+1)}" text-anchor="middle" font-size="12" fill="#4c4c4c">{s}</text>')
def group(x,y,w,h,label,fill):
    o.append(f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="16" fill="{fill}" stroke="#8a8a80" stroke-dasharray="6 5" stroke-width="1.5"/>')
    o.append(f'<text x="{x+16}" y="{y+24}" font-size="13" font-weight="bold" fill="#4c4c4c">{label}</text>')
def arrow(pts,label=None,lx=0,ly=0,both=False,anchor="middle"):
    s=" ".join(f"{a},{b}" for a,b in pts)
    ms='marker-start="url(#a)" ' if both else ''
    o.append(f'<polyline points="{s}" fill="none" stroke="#1a1a1a" stroke-width="1.6" {ms}marker-end="url(#a)"/>')
    if label:
        for i,l in enumerate(label.split("|")):
            o.append(f'<text x="{lx}" y="{ly+15*i}" text-anchor="{anchor}" font-size="12" font-style="italic" fill="#1a1a1a">{l}</text>')

group(270,80,560,600,"Go-приложение (Gin), :8080","#f4f4ee")
group(870,80,340,600,"Docker Compose","#eef3f8")

box(40,290,190,130,"Браузер",("пользователь",),fill="#fecf00")
box(300,120,500,64,"Router (Gin)",("internal/api/server.go · RegisterHandler: 3 GET + 3 POST",))
box(300,228,300,84,"Handler (контроллеры)",("handler/import_categories.go","разбор запроса, редирект, c.HTML"))
box(620,228,180,84,"Templates",("templates/*.html","html/template"))
box(300,356,500,84,"Repository (слой данных)",("repository/import_categories.go","GORM: Find, First, Create, Updates · db.Exec(UPDATE)"))
box(300,484,240,70,"Модели ds",("User · ImportCategory · Like",))
box(560,484,240,70,"config · dsn",(".env → строка подключения",))
box(300,590,500,64,"cmd/migrate",("AutoMigrate + частичный индекс + Seed",))

box(900,120,280,90,"MinIO  :9000",("бакет import-categories","img/*.jpg, video/*.mp4"))
box(900,340,280,110,"PostgreSQL 16  :5433",("база amphora","users · import_categories · likes"),fill="#fff8d6")
box(900,560,280,80,"Adminer  :8081",("панель администратора БД",))

# браузер <-> router: запрос и ответ
arrow([(230,330),(265,330),(265,152),(300,152)],"HTTP-запрос /|HTML-ответ",150,258,both=True)
arrow([(600,270),(620,270)])
arrow([(450,184),(450,228)])
arrow([(450,312),(450,356)],"вызов метода",460,338,anchor="start")
arrow([(800,398),(900,398)],"SQL",850,390)
arrow([(420,440),(420,484)])
arrow([(680,484),(680,440)])
arrow([(800,622),(850,622),(850,430),(900,430)],"миграция",856,530,anchor="start")
arrow([(1040,560),(1040,450)],"SQL-запросы",1050,510,anchor="start")
# браузер -> MinIO поверх приложения
arrow([(70,290),(70,62),(1040,62),(1040,120)],"&lt;img&gt; / &lt;video&gt; загружаются браузером напрямую по URL из БД",590,56)
o.append('</svg>')
open("docs/architecture.svg","w").write("\n".join(o))
