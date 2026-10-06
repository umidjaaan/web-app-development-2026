# ER-диаграмма стандартными фигурами draw.io «Список» (swimlane + строки):
# docs/er_list.drawio (редактируемая) + docs/er_list.svg (картинка)
RH=30; W=180
T={
 'users':(60,40,["id (PK)","login","password","full_name","is_moderator","created_at"]),
 'likes':(330,630,["id (PK)","user_id (FK)","import_category_id (FK)","created_at"]),
 'import_categories':(620,40,["id (PK)","slug","title","shape","production_center","region",
   "date_start","date_end","diagnostics","description","finds_count","image_url","video_url",
   "status","creator_id (FK)","created_at","updated_at"]),
}
def H(n): return RH+len(T[n][2])*RH
def ry(n,i): return T[n][1]+RH+i*RH+RH/2
ux,uy,_=T['users']; lx,ly,_=T['likes']; ix,iy,_=T['import_categories']
# связи: от строки PK к строке FK
E=[('e1','users',0,'import_categories',14,
    [(ux+W,ry('users',0)),(ux+W+50,ry('users',0)),(ux+W+50,ry('users',0)-0)],'ERmandOne','ERzeroToMany','создаёт'),
   ('e2','users',0,'likes',1,None,'ERmandOne','ERzeroToMany','ставит'),
   ('e3','import_categories',0,'likes',2,None,'ERmandOne','ERzeroToMany','получает')]
# маршруты (ортогональные), рисуем явно
R={
 'e1':[(ux+W,ry('users',0)),(ix-40,ry('users',0)),(ix-40,ry('import_categories',14)),(ix,ry('import_categories',14))],
 'e2':[(ux,ry('users',0)),(ux-20,ry('users',0)),(ux-20,ry('likes',1)),(lx,ry('likes',1))],
 'e3':[(ix+W,ry('import_categories',0)),(ix+W+30,ry('import_categories',0)),(ix+W+30,ry('likes',2)),(lx+W,ry('likes',2))],
}
EXIT={'e1':(1,0),'e2':(0,1),'e3':(1,1)}  # (сторона выхода, сторона входа): 1 = справа, 0 = слева
# users.id -> creator_id: 0..1 со стороны users (creator_id может быть NULL)
E[0]=('e1','users',0,'import_categories',14,None,'ERzeroToOne','ERzeroToMany','создаёт')

# ---------- draw.io ----------
c=['<mxfile host="app.diagrams.net"><diagram name="ER" id="er"><mxGraphModel dx="1200" dy="800" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="1169" pageHeight="827" math="0" shadow="0"><root><mxCell id="0"/><mxCell id="1" parent="0"/>']
for n,(x,y,rows) in T.items():
    c.append(f'<mxCell id="{n}" value="{n}" style="swimlane;fontStyle=0;childLayout=stackLayout;horizontal=1;startSize={RH};horizontalStack=0;resizeParent=1;resizeParentMax=0;resizeLast=0;collapsible=1;marginBottom=0;whiteSpace=wrap;html=1;" vertex="1" parent="1"><mxGeometry x="{x}" y="{y}" width="{W}" height="{H(n)}" as="geometry"/></mxCell>')
    for i,f in enumerate(rows):
        c.append(f'<mxCell id="{n}_{i}" value="{f}" style="text;strokeColor=none;fillColor=none;align=left;verticalAlign=middle;spacingLeft=4;spacingRight=4;overflow=hidden;points=[[0,0.5],[1,0.5]];portConstraint=eastwest;rotatable=0;whiteSpace=wrap;html=1;" vertex="1" parent="{n}"><mxGeometry y="{RH+i*RH}" width="{W}" height="{RH}" as="geometry"/></mxCell>')
for eid,s,si,t,ti,_,sa,ea,lab in E:
    pts=R[eid]
    mids=''.join(f'<mxPoint x="{a}" y="{b}"/>' for a,b in pts[1:-1])
    arr='<Array as="points">'+mids+'</Array>'
    c.append(f'<mxCell id="{eid}" value="{lab}" style="edgeStyle=orthogonalEdgeStyle;rounded=0;html=1;exitX={EXIT[eid][0]};exitY=0.5;exitDx=0;exitDy=0;entryX={EXIT[eid][1]};entryY=0.5;entryDx=0;entryDy=0;labelBackgroundColor=#ffffff;startArrow={sa};endArrow={ea};startFill=0;endFill=0;startSize=10;endSize=10;" edge="1" parent="1" source="{s}_{si}" target="{t}_{ti}"><mxGeometry relative="1" as="geometry">{arr}</mxGeometry></mxCell>')
c.append('</root></mxGraphModel></diagram></mxfile>')
open('docs/er_list.drawio','w').write('\n'.join(c))

# ---------- SVG, как экспорт из draw.io ----------
CW,CH=880,810
o=[f'<svg xmlns="http://www.w3.org/2000/svg" width="{CW}" height="{CH}" viewBox="0 0 {CW} {CH}" font-family="Helvetica, Arial, sans-serif">',f'<rect width="{CW}" height="{CH}" fill="#fff"/>']
for n,(x,y,rows) in T.items():
    o.append(f'<rect x="{x}" y="{y}" width="{W}" height="{H(n)}" fill="#fff" stroke="#000"/>')
    o.append(f'<line x1="{x}" y1="{y+RH}" x2="{x+W}" y2="{y+RH}" stroke="#000"/>')
    o.append(f'<rect x="{x+4}" y="{y+4}" width="9" height="9" fill="#fff" stroke="#666" stroke-width="0.8"/><line x1="{x+6}" y1="{y+8.5}" x2="{x+11}" y2="{y+8.5}" stroke="#333"/>')
    o.append(f'<text x="{x+W/2}" y="{y+19}" text-anchor="middle" font-size="12">{n}</text>')
    for i,f in enumerate(rows):
        o.append(f'<text x="{x+6}" y="{y+RH+i*RH+19}" font-size="12">{f}</text>')
S='stroke="#000" stroke-width="1" fill="none"'
def mark(kind,x,y,d):
    bar=lambda off: f'<line x1="{x+d*off}" y1="{y-7}" x2="{x+d*off}" y2="{y+7}" {S}/>'
    circ=lambda off: f'<circle cx="{x+d*off}" cy="{y}" r="5" fill="#fff" stroke="#000"/>'
    if kind=='ERmandOne': return [bar(7),bar(12)]
    if kind=='ERzeroToOne': return [bar(7),circ(17)]
    return [f'<line x1="{x}" y1="{y-7}" x2="{x+d*10}" y2="{y}" {S}/>',f'<line x1="{x}" y1="{y+7}" x2="{x+d*10}" y2="{y}" {S}/>',
            f'<line x1="{x}" y1="{y}" x2="{x+d*10}" y2="{y}" {S}/>',circ(15)]
LP={'e1':(ix-40,(ry('users',0)+ry('import_categories',14))/2),'e2':(ux-20,(ry('users',0)+ry('likes',1))/2),
    'e3':(ix+W+30,(ry('import_categories',0)+ry('likes',2))/2)}
for eid,s,si,t,ti,_,sa,ea,lab in E:
    pts=R[eid]
    o.append(f'<polyline points="{" ".join(f"{a},{b}" for a,b in pts)}" {S}/>')
    d0=1 if pts[1][0]>pts[0][0] else -1; d1=1 if pts[-2][0]>pts[-1][0] else -1
    o+=mark(sa,*pts[0],d0); o+=mark(ea,*pts[-1],d1)
    x_,y_=LP[eid]; w=len(lab)*7+6
    o.append(f'<rect x="{x_-w/2}" y="{y_-9}" width="{w}" height="16" fill="#fff"/><text x="{x_}" y="{y_+3}" text-anchor="middle" font-size="11">{lab}</text>')
o.append('</svg>')
open('docs/er_list.svg','w').write('\n'.join(o))
