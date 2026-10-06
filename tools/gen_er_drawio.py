# ER-диаграмма в стиле draw.io: docs/er.drawio (редактируемая) + docs/er_drawio.svg (картинка)
from xml.sax.saxutils import escape as esc
RH=30; KW=40; NW=160; TW=110; W=KW+NW+TW
HEAD='#dae8fc'; STROKE='#6c8ebf'
T={
 'users':(40,40,[("PK","id","bigint"),("","login","varchar(64)"),("","password","varchar(128)"),
   ("","full_name","varchar(128)"),("","is_moderator","boolean"),("","created_at","timestamptz")]),
 'likes':(420,330,[("PK","id","bigint"),("FK","user_id","bigint"),("FK","import_category_id","bigint"),("","created_at","timestamptz")]),
 'import_categories':(800,40,[("PK","id","bigint"),("","slug","varchar(64)"),("","title","varchar(128)"),
   ("","shape","varchar(64)"),("","production_center","varchar(128)"),("","region","varchar(64)"),
   ("","date_start","date"),("","date_end","date"),("","diagnostics","text"),("","description","text"),
   ("","finds_count","bigint"),("","image_url","varchar(256)"),("","video_url","varchar(256)"),
   ("","status","varchar(16)"),("FK","creator_id","bigint"),("","created_at","timestamptz"),("","updated_at","timestamptz")]),
}
def H(n): return RH+len(T[n][2])*RH
def rowy(n,i): return T[n][1]+RH+i*RH+RH/2
ux,uy,_=T['users']; lx,ly,_=T['likes']; ix,iy,_=T['import_categories']
# рёбра: (id, source, target, точки, start, end, подпись, позиция подписи)
E=[('e1','users','import_categories',[(ux+W,rowy('users',0)),(ix,rowy('users',0))],'ERzeroToOne','ERzeroToMany','создаёт',((ux+W+ix)/2,rowy('users',0)-8)),
   ('e2','users','likes',[(ux+W,uy+170),(ux+W+45,uy+170),(ux+W+45,rowy('likes',1)),(lx,rowy('likes',1))],'ERmandOne','ERzeroToMany','ставит',(ux+W+50,(uy+170+rowy('likes',1))/2)),
   ('e3','import_categories','likes',[(ix,iy+280),(ix-45,iy+280),(ix-45,rowy('likes',2)),(lx+W,rowy('likes',2))],'ERmandOne','ERzeroToMany','получает',(ix-50,iy+270))]

# ---------- draw.io ----------
c=['<mxfile host="app.diagrams.net"><diagram name="ER" id="er"><mxGraphModel dx="1200" dy="800" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="1169" pageHeight="827" math="0" shadow="0"><root><mxCell id="0"/><mxCell id="1" parent="0"/>']
for n,(x,y,rows) in T.items():
    c.append(f'<mxCell id="{n}" value="{n}" style="shape=table;startSize={RH};container=1;collapsible=0;childLayout=tableLayout;fixedRows=1;rowLines=0;fontStyle=1;align=center;resizeLast=1;html=1;fillColor={HEAD};strokeColor={STROKE};fontSize=13;" vertex="1" parent="1"><mxGeometry x="{x}" y="{y}" width="{W}" height="{H(n)}" as="geometry"/></mxCell>')
    for i,(k,nm,tp) in enumerate(rows):
        r=f'{n}_r{i}'; bot=1 if k=="PK" else 0
        c.append(f'<mxCell id="{r}" value="" style="shape=tableRow;horizontal=0;startSize=0;swimlaneHead=0;swimlaneBody=0;fillColor=#ffffff;collapsible=0;dropTarget=0;points=[[0,0.5],[1,0.5]];portConstraint=eastwest;top=0;left=0;right=0;bottom={bot};strokeColor={STROKE};" vertex="1" parent="{n}"><mxGeometry y="{RH+i*RH}" width="{W}" height="{RH}" as="geometry"/></mxCell>')
        base='shape=partialRectangle;connectable=0;fillColor=none;top=0;left=0;bottom=0;right=0;overflow=hidden;whiteSpace=wrap;html=1;'
        fs=5 if k=="PK" else 0
        c.append(f'<mxCell id="{r}_k" value="{k}" style="{base}fontStyle=1;" vertex="1" parent="{r}"><mxGeometry width="{KW}" height="{RH}" as="geometry"><mxRectangle width="{KW}" height="{RH}" as="alternateBounds"/></mxGeometry></mxCell>')
        c.append(f'<mxCell id="{r}_n" value="{nm}" style="{base}align=left;spacingLeft=6;fontStyle={fs};" vertex="1" parent="{r}"><mxGeometry x="{KW}" width="{NW}" height="{RH}" as="geometry"><mxRectangle width="{NW}" height="{RH}" as="alternateBounds"/></mxGeometry></mxCell>')
        c.append(f'<mxCell id="{r}_t" value="{tp}" style="{base}align=left;spacingLeft=6;fontColor=#555555;" vertex="1" parent="{r}"><mxGeometry x="{KW+NW}" width="{TW}" height="{RH}" as="geometry"><mxRectangle width="{TW}" height="{RH}" as="alternateBounds"/></mxGeometry></mxCell>')
for eid,s,t,pts,sa,ea,lab,_ in E:
    sx,sy,_=T[s]; tx,ty,_=T[t]
    ex=(pts[0][0]-sx)/W; ey=(pts[0][1]-sy)/H(s); nx=(pts[-1][0]-tx)/W; ny=(pts[-1][1]-ty)/H(t)
    mids=''.join(f'<mxPoint x="{a}" y="{b}"/>' for a,b in pts[1:-1])
    arr='<Array as="points">'+mids+'</Array>' if mids else ''
    c.append(f'<mxCell id="{eid}" value="{lab}" style="edgeStyle=orthogonalEdgeStyle;rounded=0;html=1;fontSize=12;fontStyle=2;labelBackgroundColor=#ffffff;startArrow={sa};endArrow={ea};startFill=0;endFill=0;startSize=10;endSize=10;exitX={ex:.4f};exitY={ey:.4f};exitDx=0;exitDy=0;entryX={nx:.4f};entryY={ny:.4f};entryDx=0;entryDy=0;" edge="1" parent="1" source="{s}" target="{t}"><mxGeometry relative="1" as="geometry">{arr}</mxGeometry></mxCell>')
c.append('</root></mxGraphModel></diagram></mxfile>')
open('docs/er.drawio','w').write('\n'.join(c))

# ---------- SVG в виде экспорта draw.io ----------
CW,CH=1150,640
o=[f'<svg xmlns="http://www.w3.org/2000/svg" width="{CW}" height="{CH}" viewBox="0 0 {CW} {CH}" font-family="Helvetica, Arial, sans-serif">',f'<rect width="{CW}" height="{CH}" fill="#fff"/>']
for n,(x,y,rows) in T.items():
    h=H(n)
    o.append(f'<rect x="{x}" y="{y}" width="{W}" height="{h}" fill="#fff" stroke="{STROKE}"/>')
    o.append(f'<rect x="{x}" y="{y}" width="{W}" height="{RH}" fill="{HEAD}" stroke="{STROKE}"/>')
    o.append(f'<text x="{x+W/2}" y="{y+20}" text-anchor="middle" font-size="13" font-weight="bold">{n}</text>')
    for i,(k,nm,tp) in enumerate(rows):
        ry=y+RH+i*RH
        if k=="PK": o.append(f'<line x1="{x}" y1="{ry+RH}" x2="{x+W}" y2="{ry+RH}" stroke="{STROKE}"/>')
        o.append(f'<text x="{x+KW/2}" y="{ry+19}" text-anchor="middle" font-size="12" font-weight="bold">{k}</text>')
        dec=' font-weight="bold" text-decoration="underline"' if k=="PK" else ''
        o.append(f'<text x="{x+KW+6}" y="{ry+19}" font-size="12"{dec}>{nm}</text>')
        o.append(f'<text x="{x+KW+NW+6}" y="{ry+19}" font-size="12" fill="#555">{tp}</text>')
S='stroke="#000" stroke-width="1" fill="none"'
def mark(kind,x,y,d):  # d: направление от таблицы наружу (+1 вправо, -1 влево)
    r=[]
    bar=lambda off: f'<line x1="{x+d*off}" y1="{y-7}" x2="{x+d*off}" y2="{y+7}" {S}/>'
    circ=lambda off: f'<circle cx="{x+d*off}" cy="{y}" r="5" fill="#fff" stroke="#000"/>'
    if kind=='ERmandOne': r+= [bar(7),bar(12)]
    if kind=='ERzeroToOne': r+= [bar(7),circ(17)]
    if kind=='ERzeroToMany':
        r+= [f'<line x1="{x}" y1="{y-7}" x2="{x+d*10}" y2="{y}" {S}/>',f'<line x1="{x}" y1="{y+7}" x2="{x+d*10}" y2="{y}" {S}/>',
             f'<line x1="{x}" y1="{y}" x2="{x+d*10}" y2="{y}" {S}/>',circ(15)]
    return r
for eid,s,t,pts,sa,ea,lab,(lx_,ly_) in E:
    o.append(f'<polyline points="{" ".join(f"{a},{b}" for a,b in pts)}" {S}/>')
    d0=1 if pts[1][0]>pts[0][0] else -1
    d1=1 if pts[-2][0]>pts[-1][0] else -1
    o+=mark(sa,*pts[0],d0); o+=mark(ea,*pts[-1],d1)
    anchor='middle' if eid=='e1' else ('start' if eid=='e2' else 'end')
    o.append(f'<text x="{lx_}" y="{ly_}" text-anchor="{anchor}" font-size="12" font-style="italic">{lab}</text>')
o.append('</svg>')
open('docs/er_drawio.svg','w').write('\n'.join(o))
