#!/usr/bin/env python3
"""Web review UI for video-alerter: browse a day's clips, verdicts, thumbnails; mark false positives.

stdlib only. Run on the machine holding state.json (usually the watcher's machine).
    ALERTER_LLM_* not needed. Port: ALERTER_REVIEW_PORT (default 8765), host: ALERTER_REVIEW_HOST (default 127.0.0.1).
"""

import json
import os
import re
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlparse, parse_qs

PROJECT_DIR = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, PROJECT_DIR)
import alerter  # noqa: E402

ROOT = alerter.ROOT
STATE_PATH = alerter.STATE_PATH
FEEDBACK_PATH = alerter.FEEDBACK_PATH
SCENE_PATH = alerter.SCENE_PATH
PORT = int(os.environ.get("ALERTER_REVIEW_PORT", "8765"))
HOST = os.environ.get("ALERTER_REVIEW_HOST", "127.0.0.1")

CLIP_KEY_RE = re.compile(r"^[^./][^/]*/\d{4}-\d{2}-\d{2}/[^./][^/]*\.m4v$")

PAGE = """<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>video-alerter review</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
:root{color-scheme:dark}
body{font:14px/1.45 -apple-system,system-ui,sans-serif;background:#111418;color:#e6e6e6;margin:0}
header{position:sticky;top:0;background:#171b21;padding:10px 16px;display:flex;gap:10px;align-items:center;flex-wrap:wrap;border-bottom:1px solid #2a2f37}
header h1{font-size:15px;margin:0 12px 0 0}
.chip{padding:4px 10px;border-radius:14px;background:#232932;cursor:pointer;border:1px solid #2f3742}
.chip.sel{background:#2f6fed;border-color:#2f6fed}
label{color:#9aa4b0;user-select:none}
main{padding:16px;max-width:1280px;margin:0 auto}
.cam{margin:18px 0 8px;font-size:13px;color:#9aa4b0;text-transform:uppercase;letter-spacing:.06em}
.scene{font-size:12px;color:#7d8794;margin:-4px 0 12px;max-width:900px}
.card{display:flex;gap:14px;background:#171b21;border:1px solid #2a2f37;border-radius:10px;padding:12px;margin-bottom:12px}
.frame{position:relative;width:480px;height:270px;flex:none;border-radius:6px;overflow:hidden;background:#000}
.frame img,.frame video{width:100%;height:100%;object-fit:cover;display:block}
.frame img{cursor:pointer}
.cap{position:absolute;left:0;right:0;bottom:0;background:rgba(0,0,0,.72);color:#fff;font-size:12.5px;line-height:1.35;padding:6px 9px;max-height:50%;overflow:hidden}
.cap.hot{background:rgba(140,30,30,.88)}
.cap .threat{font-weight:700;color:#ffb3b3}
.muted{color:#9aa4b0;font-style:italic}
.body{min-width:0;flex:1}
.when{font-weight:600}
.badge{display:inline-block;padding:1px 8px;border-radius:10px;font-size:11px;margin-left:8px;vertical-align:1px}
.b-alert{background:#7a1f1f;color:#ffd9d9}.b-ok{background:#1f3a24;color:#b8e6c3}
.b-skip{background:#333;color:#aaa}.b-pending,.b-unanalyzed{background:#4a3c10;color:#f0d98c}
.b-marked{background:#2f6fed;color:#fff;margin-left:6px}
.markbar{display:flex;gap:8px;align-items:center;flex-wrap:wrap;margin-top:10px}
.markbar input{background:#0d1014;border:1px solid #2f3742;color:#e6e6e6;border-radius:6px;padding:4px 8px;flex:1;min-width:160px}
button{background:#232932;color:#e6e6e6;border:1px solid #2f3742;border-radius:6px;padding:4px 12px;cursor:pointer}
button.fp:hover{border-color:#c0392b}button.tp:hover{border-color:#27ae60}
.empty{color:#7d8794;padding:40px;text-align:center}
</style></head><body>
<header><h1>video-alerter</h1><span id="days"></span>
<label><input type="checkbox" id="alertsOnly"> alerts only</label>
<label><input type="checkbox" id="showMarked"> show reviewed</label>
<span id="count" style="color:#9aa4b0"></span></header>
<main id="out"><div class="empty">loading…</div></main>
<script>
let day=null, data=null;
async function j(u,opt){const r=await fetch(u,opt);return r.json()}
function esc(s){return (s||'').replace(/[&<>"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]))}
function fmtMsg(m){return esc(m).replace(/^\[(HIGH|LOW)\]\s*/,m2=>'<span class="threat">'+m2.trim()+'</span> ')}
async function loadDays(){
  const d=await j('/api/days'); const el=document.getElementById('days'); el.innerHTML='';
  d.days.slice(0,21).reverse().forEach(dt=>{
    const c=document.createElement('span');c.className='chip';c.textContent=dt;
    c.onclick=()=>{day=dt;[...el.children].forEach(x=>x.classList.toggle('sel',x.textContent===dt));loadDay()};
    el.appendChild(c)});
  if(d.days.length){day=d.days[d.days.length-1];el.lastChild.classList.add('sel');loadDay()}
  else document.getElementById('out').innerHTML='<div class="empty">no clips in state yet</div>';
}
async function loadDay(){
  if(!day)return; data=await j('/api/day?date='+day); render();
}
function render(){
  const out=document.getElementById('out');
  const only=document.getElementById('alertsOnly').checked;
  const showMarked=document.getElementById('showMarked').checked;
  let alerts=0, todo=0;
  data.cameras.forEach(cam=>cam.clips.forEach(c=>{if(c.status==='alert'){alerts++;if(!c.marked)todo++}}));
  document.getElementById('count').textContent=alerts?`${alerts} alert${alerts>1?'s':''}, ${todo} awaiting review`:'no alerts';
  let html='';
  data.cameras.forEach(cam=>{
    const clips=cam.clips.filter(c=>(!only||c.status==='alert')&&(showMarked||!c.marked));
    if(!clips.length)return;
    html+=`<div class="cam">${esc(cam.name)} — ${clips.length} clip${clips.length>1?'s':''}</div>`;
    if(cam.scene)html+=`<div class="scene">${esc(cam.scene)}</div>`;
    clips.forEach(c=>{
      const marked=c.marked?`<span class="badge b-marked">${esc(c.marked.label)}${c.marked.note?': '+esc(c.marked.note):''}</span>`:'';
      const m=c.msg?fmtMsg(c.msg):'<span class="muted">not analyzed yet — waiting in the watcher queue</span>';
      html+=`<div class="card"><div class="frame"><img src="/thumb?clip=${encodeURIComponent(c.clip)}" onclick="play(this,'${esc(c.clip)}')">
      <div class="cap ${c.status==='alert'?'hot':''}">${m}</div></div>
      <div class="body"><span class="when">${esc(c.when)}</span>
      <span class="badge b-${c.status}">${c.status}</span>${marked}
      <div class="markbar"><input placeholder="why not alarming? (optional note)" id="n-${esc(c.clip)}">
      <button class="fp" onclick="mark('${esc(c.clip)}','fp')">false positive</button>
      <button class="tp" onclick="mark('${esc(c.clip)}','tp')">real alert</button></div></div></div>`;
    });
  });
  out.innerHTML=html||'<div class="empty">nothing this day</div>';
}
function play(img,clip){const v=document.createElement('video');v.src='/video?clip='+encodeURIComponent(clip);v.controls=true;v.autoplay=true;img.replaceWith(v)}
async function mark(clip,label){
  const note=document.getElementById('n-'+clip).value;
  const r=await j('/api/mark',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({clip,label,note})});
  await loadDay();
  if(r.ok){
    const el=document.getElementById('count'), t=el.textContent;
    el.textContent=`✓ pushed to learner — ${r.examples} active correction${r.examples===1?'':'s'}`;
    setTimeout(()=>el.textContent=t,4000);
  }
}
document.getElementById('alertsOnly').onchange=render;
document.getElementById('showMarked').onchange=render;
loadDays();
</script></body></html>"""


def safe_clip(rel):
    # type: (str) -> bool
    return bool(CLIP_KEY_RE.match(rel)) and ".." not in rel


def clip_abs(rel):
    # type: (str) -> str
    return os.path.join(ROOT, rel)


class Handler(BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        pass

    def _send(self, code, ctype, body, extra=None):
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(body)))
        for k, v in (extra or {}).items():
            self.send_header(k, v)
        self.end_headers()
        if self.command != "HEAD":
            self.wfile.write(body)

    def _send_file(self, path, ctype):
        try:
            size = os.path.getsize(path)
        except OSError:
            self._send(404, "text/plain", b"not found")
            return
        start, end = 0, size - 1
        rng = self.headers.get("Range")
        if rng:
            m = re.match(r"bytes=(\d*)-(\d*)", rng)
            if m:
                if m.group(1):
                    start = int(m.group(1))
                if m.group(2):
                    end = min(int(m.group(2)), size - 1)
                end = max(end, start)
        with open(path, "rb") as f:
            f.seek(start)
            body = f.read(end - start + 1)
        code = 206 if rng else 200
        extra = {"Accept-Ranges": "bytes"}
        if rng:
            extra["Content-Range"] = "bytes %d-%d/%d" % (start, end, size)
        self._send(code, ctype, body, extra)

    def do_GET(self):
        url = urlparse(self.path)
        q = parse_qs(url.query)
        if url.path == "/":
            self._send(200, "text/html; charset=utf-8", PAGE.encode("utf-8"))
        elif url.path == "/api/days":
            state = alerter.load_json(STATE_PATH, {"clips": {}})
            days = sorted({k.split("/")[1] for k in state.get("clips", {}) if "/" in k})
            self._json({"days": days})
        elif url.path == "/api/day":
            self._day(q.get("date", [""])[0])
        elif url.path == "/thumb":
            rel = q.get("clip", [""])[0]
            if not safe_clip(rel):
                self._send(400, "text/plain", b"bad clip")
                return
            thm = alerter.thm_path(clip_abs(rel))
            jpeg = alerter.decode_thm_jpeg(thm) if os.path.exists(thm) else None
            if jpeg:
                self._send(200, "image/jpeg", jpeg)
            else:
                self._send(404, "text/plain", b"no thumbnail")
        elif url.path == "/video":
            rel = q.get("clip", [""])[0]
            p = clip_abs(rel) if safe_clip(rel) else ""
            if p and os.path.isfile(p):
                self._send_file(p, "video/mp4")
            else:
                self._send(404, "text/plain", b"not found")
        else:
            self._send(404, "text/plain", b"not found")

    def do_POST(self):
        if urlparse(self.path).path != "/api/mark":
            self._send(404, "text/plain", b"not found")
            return
        try:
            n = int(self.headers.get("Content-Length", "0"))
            body = json.loads(self.rfile.read(n).decode("utf-8"))
            rel = body["clip"]
            label = body.get("label", "fp")
            assert safe_clip(rel) and label in ("fp", "tp")
        except (ValueError, AssertionError, KeyError):
            self._send(400, "application/json", b'{"ok":false,"error":"bad request"}')
            return
        note = str(body.get("note", ""))[:200]
        state = alerter.load_json(STATE_PATH, {"clips": {}})
        fb = alerter.load_feedback()
        matches = alerter.mark_feedback(fb, state, rel, label, note)
        alerter.save_feedback(fb)
        self._json({"ok": True, "matched": len(matches), "examples": len(fb["examples"])})

    def _json(self, obj):
        self._send(200, "application/json", json.dumps(obj).encode("utf-8"))

    def _day(self, date):
        if not re.match(r"^\d{4}-\d{2}-\d{2}$", date or ""):
            self._json({"date": date, "cameras": []})
            return
        state = alerter.load_json(STATE_PATH, {"clips": {}})
        fb = alerter.load_feedback()
        scenes = alerter.load_json(SCENE_PATH, {})
        cams = {}
        for k, entry in state.get("clips", {}).items():
            parts = k.split("/")
            if len(parts) != 3 or parts[1] != date:
                continue
            when = alerter.clip_start_time(parts[2]) or parts[2]
            cams.setdefault(parts[0], []).append({
                "clip": k,
                "when": when,
                "status": entry.get("status", "?"),
                "msg": entry.get("msg", ""),
                "marked": fb["marks"].get(k),
            })
        cameras = [
            {"name": name, "scene": (scenes.get(name) or {}).get("text", ""),
             "clips": sorted(items, key=lambda c: c["when"])}
            for name, items in sorted(cams.items())
        ]
        self._json({"date": date, "cameras": cameras})


def main():
    if not os.path.exists(STATE_PATH):
        print("warning: %s missing — run alerter.py first" % STATE_PATH)
    print("review UI on http://%s:%d (state=%s)" % (HOST, PORT, STATE_PATH), flush=True)
    ThreadingHTTPServer((HOST, PORT), Handler).serve_forever()


if __name__ == "__main__":
    main()
