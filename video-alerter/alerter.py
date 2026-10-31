#!/usr/bin/env python3
"""Watch SecuritySpy capture dirs, pre-filter motion clips, alert via vision LLM.

Runs on kahuna.local or the dev Mac. stdlib only, Python 3.9 compatible.
See CLAUDE.md for layout, constraints and env vars.
"""

import argparse
import base64
import glob
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
import threading
import time
import urllib.error
import urllib.request
from typing import Any, Callable, Dict, List, Optional, Tuple

HOME = os.path.expanduser("~")
PROJECT_DIR = os.path.dirname(os.path.abspath(__file__))
_ROOT_CANDIDATES = [
    os.environ.get("ALERTER_ROOT"),
    "/Volumes/SecuritySpy",                      # dev Mac mount of the live share
    "/Volumes/securityspy",
    os.path.expanduser("~/Volumes/securityspy"),  # kahuna mount (\/Volumes is protected there)
    "/Volumes/aaron/SecuritySpy_secondary",       # stale former target, last resort
]
ROOT = next((p for p in _ROOT_CANDIDATES if p and os.path.isdir(p)), "")
STATE_PATH = os.environ.get("ALERTER_STATE", os.path.join(PROJECT_DIR, "state.json"))
LOG_PATH = os.environ.get("ALERTER_LOG", os.path.join(PROJECT_DIR, "alerts.log"))
SCENE_PATH = os.environ.get("ALERTER_SCENE", os.path.join(PROJECT_DIR, "scene.json"))
FEEDBACK_PATH = os.environ.get("ALERTER_FEEDBACK", os.path.join(PROJECT_DIR, "feedback.json"))
KB_DIR = os.environ.get("ALERTER_KB", os.path.join(PROJECT_DIR, "knowledge"))
KB_MAX_MB = int(os.environ.get("ALERTER_KB_MAX_MB", "20000"))  # 20 GB ceiling; frames are tiny
TMP_DIR = os.path.join(PROJECT_DIR, "tmp")
LLM_BASE = os.environ.get("ALERTER_LLM_BASE", "http://127.0.0.1:8095/v1").rstrip("/")
LLM_MODEL = os.environ.get("ALERTER_LLM_MODEL", "qwen3-vl-30b-a3b")
LLM_TIMEOUT = float(os.environ.get("ALERTER_LLM_TIMEOUT", "180"))
DIFF_THRESHOLD = float(os.environ.get("ALERTER_DIFF_THRESHOLD", "12.0"))
EVENT_WINDOW_S = float(os.environ.get("ALERTER_EVENT_WINDOW", "180"))
EVENT_DIFF_THRESHOLD = float(os.environ.get("ALERTER_EVENT_DIFF_THRESHOLD", "20.0"))
FP_THUMB_DIFF = float(os.environ.get("ALERTER_FP_THUMB_DIFF", "5.0"))
FP_TEXT_SIM = float(os.environ.get("ALERTER_FP_TEXT_SIM", "0.5"))

CLIP_RE = re.compile(r"^(\d{2})-(\d{2})-(\d{4}) (\d{2})-(\d{2})-(\d{2}) (.+)\.m4v$")

PROMPT = (
    "You are a security-camera analyst. These frames are sampled from one motion "
    "clip (camera: {camera}, local time {when}). Describe what is happening and "
    "rate the threat. Reply with ONLY a JSON object, no prose, exactly: "
    '{{"what_seen": "<one sentence>", "threat": "none|low|high", "why": "<short reason>"}}. '
    "threat=high means a person, vehicle, or animal near the property that the owner "
    "should know about, or anything that looks like a break-in, damage, fire or danger. "
    "threat=low means ordinary activity (passers-by, distant traffic). "
    "threat=none means environmental motion only: swaying trees/branches/grass in wind, "
    "moving or fading shadows, cloud-driven light changes, rain or snow, camera noise, "
    "small creatures, or no clear subject. Environmental motion is threat=none no matter "
    "how much of the frame is moving."
)

SCENE_PROMPT = (
    "These are frames from the SAME fixed security camera, one by day and one by "
    "night. Describe the PERMANENT layout only: structures, driveway, garage, gates, "
    "road/sidewalk, fences, vegetation; say which areas are the private property and "
    "which are public (street/sidewalk), and note day vs night appearance differences. "
    "3-5 sentences, plain text, no JSON."
)


def load_json(path, default):
    # type: (str, Dict) -> Dict
    try:
        with open(path) as f:
            return json.load(f)
    except (OSError, ValueError):
        return default


def save_json(path, obj):
    # type: (str, Dict) -> None
    tmp = path + ".tmp"
    with open(tmp, "w") as f:
        json.dump(obj, f, indent=1)
    os.replace(tmp, path)


def load_scenes():
    # type: () -> Dict
    return load_json(SCENE_PATH, {})


def save_scenes(scenes):
    # type: (Dict) -> None
    save_json(SCENE_PATH, scenes)


def load_feedback():
    # type: () -> Dict
    fb = load_json(FEEDBACK_PATH, {})
    fb.setdefault("marks", {})
    fb.setdefault("examples", [])
    fb.setdefault("overrides", {})
    fb.setdefault("learn_mode", False)
    return fb


def save_feedback(fb):
    # type: (Dict) -> None
    save_json(FEEDBACK_PATH, fb)


FB_LOCK = threading.Lock()


def update_feedback(mutate):
    # type: (Callable[[Dict], Any]) -> Any
    """Serialized load -> mutate -> save of feedback.json for threaded callers (review server)."""
    with FB_LOCK:
        fb = load_feedback()
        out = mutate(fb)
        save_feedback(fb)
    return out


def rid(s):
    # type: (str) -> str
    return hashlib.sha1(s.encode("utf-8")).hexdigest()[:16]


def kb_dir_for(kind, ident):
    # type: (str, str) -> str
    return os.path.join(KB_DIR, kind, rid(ident))


def kb_free_bytes():
    # type: () -> int
    total = 0
    for root, _dirs, files in os.walk(KB_DIR):
        for fn in files:
            try:
                total += os.path.getsize(os.path.join(root, fn))
            except OSError:
                pass
    return max(0, KB_MAX_MB * 1024 * 1024 - total)


def kb_store_clip(rel, when, camera, status, msg, thm_jpeg, frames):
    # type: (str, str, str, str, str, Optional[bytes], List[bytes]) -> str
    """Persist everything the system saw and concluded for one clip."""
    d = kb_dir_for("clips", rel)
    try:
        if kb_free_bytes() < 50 * 1024 * 1024:
            return d
        os.makedirs(d, exist_ok=True)
        if thm_jpeg:
            with open(os.path.join(d, "thumb.jpg"), "wb") as f:
                f.write(thm_jpeg)
        for i, img in enumerate(frames):
            with open(os.path.join(d, "frame%d.jpg" % i), "wb") as f:
                f.write(img)
        meta = {"clip": rel, "when": when, "camera": camera, "status": status,
                "msg": msg, "model": LLM_MODEL, "at": time.time()}
        with open(os.path.join(d, "meta.json"), "w") as f:
            json.dump(meta, f)
        with open(os.path.join(KB_DIR, "clips.jsonl"), "a") as f:
            f.write(json.dumps(meta) + "\n")
    except OSError:
        pass
    return d


def kb_promote_example(clip_rel, label, note):
    # type: (str, str, str) -> bool
    """Copy a reviewed clip's evidence into the few-shot example library."""
    src = kb_dir_for("clips", clip_rel)
    dst = kb_dir_for("examples", clip_rel)
    try:
        if not os.path.isdir(src):
            video = os.path.join(ROOT, clip_rel)
            os.makedirs(dst, exist_ok=True)
            thm = decode_thm_jpeg(thm_path(video))
            if thm:
                with open(os.path.join(dst, "thumb.jpg"), "wb") as f:
                    f.write(thm)
        else:
            if os.path.isdir(dst):
                shutil.rmtree(dst)
            shutil.copytree(src, dst)
        meta_path = os.path.join(dst, "meta.json")
        meta = {}
        try:
            with open(meta_path) as f:
                meta = json.load(f)
        except (OSError, ValueError):
            meta = {"clip": clip_rel}
        meta.update({"label": label, "note": note, "marked_at": time.time()})
        with open(meta_path, "w") as f:
            json.dump(meta, f)
        return True
    except OSError:
        return False


def kb_example_images(examples, limit=3):
    # type: (List[Dict], int) -> List[Tuple[str, bytes]]
    """Most-recent FP examples that have a stored image, for image few-shotting."""
    out = []
    for e in examples[:limit]:
        p = os.path.join(kb_dir_for("examples", e["clip"]), "thumb.jpg")
        try:
            with open(p, "rb") as f:
                img = f.read()
        except OSError:
            continue
        if img:
            caption = "Owner-labeled FALSE POSITIVE: %s%s" % (
                e.get("msg", "")[:140], " — owner note: " + e["note"] if e.get("note") else "")
            out.append((caption, img))
    return out


def build_prompt(camera, when, scenes, examples):
    # type: (str, str, Dict, List) -> str
    p = PROMPT.format(camera=camera, when=when)
    scene = scenes.get(camera)
    if scene:
        p += " Learned scene map for this camera: %s" % scene["text"]
    if examples:
        p += " Owner corrections on past alerts — respect them when judging similar scenes:"
        for e in examples[:5]:
            note = " (%s)" % e["note"] if e.get("note") else ""
            p += " [%s] %s%s" % (e["label"].upper(), e["msg"][:140], note)
    return p


def find_bin(name):
    # type: (str) -> Optional[str]
    return shutil.which(name) or (
        "/opt/homebrew/bin/" + name if os.path.exists("/opt/homebrew/bin/" + name) else None
    )


FFMPEG = find_bin("ffmpeg")
FFPROBE = find_bin("ffprobe")


def log(msg):
    # type: (str) -> None
    line = "%s %s" % (time.strftime("%Y-%m-%d %H:%M:%S"), msg)
    print(line, flush=True)
    try:
        with open(LOG_PATH, "a") as f:
            f.write(line + "\n")
    except OSError:
        pass


def notify(title, body):
    # type: (str, str) -> None
    script = (
        'display notification %s with title %s sound name "Glass"'
        % (applescript_str(body), applescript_str(title))
    )
    try:
        subprocess.run(["osascript", "-e", script], timeout=10)
    except (OSError, subprocess.TimeoutExpired):
        pass


def applescript_str(s):
    # type: (str) -> str
    esc = s.replace("\\", "\\\\").replace('"', '\\"')
    return '"' + esc + '"'


def load_state():
    # type: () -> Dict
    try:
        with open(STATE_PATH) as f:
            return json.load(f)
    except (OSError, ValueError):
        return {"clips": {}, "baselines": {}}


def save_state(state):
    # type: (Dict) -> None
    tmp = STATE_PATH + ".tmp"
    with open(tmp, "w") as f:
        json.dump(state, f, indent=1)
    os.replace(tmp, STATE_PATH)


def clip_start_time(name):
    # type: (str) -> str
    m = CLIP_RE.match(name)
    if not m:
        return ""
    return "%s-%s-%s %s:%s:%s" % (m.group(3), m.group(1), m.group(2), m.group(4), m.group(5), m.group(6))


def enumerate_clips():
    # type: () -> List[Tuple[str, str, str, str]]
    """Return (rel_path, camera, when, video_abs) for every .m4v, sorted by camera+time."""
    out = []
    try:
        cameras = sorted(
            d for d in glob.glob(os.path.join(ROOT, "*"))
            if os.path.isdir(d) and not os.path.basename(d).startswith(".")
        )
    except OSError:
        return out
    for cam_dir in cameras:
        camera = os.path.basename(cam_dir)
        for day_dir in sorted(glob.glob(os.path.join(cam_dir, "*"))):
            if not os.path.isdir(day_dir):
                continue
            try:
                names = sorted(os.listdir(day_dir))
            except OSError:
                continue
            for name in names:
                if name.startswith(".") or not name.endswith(".m4v"):
                    continue
                video = os.path.join(day_dir, name)
                rel = os.path.join(camera, os.path.basename(day_dir), name)
                out.append((rel, camera, clip_start_time(name), video))
    return out


def thm_path(video):
    # type: (str) -> str
    d, name = os.path.split(video)
    return os.path.join(d, "." + name[:-4] + ".thm")


def video_ready(video, entry):
    # type: (str, Dict) -> bool
    """Ready = .thm exists and video size unchanged since the previous poll."""
    if not os.path.exists(thm_path(video)):
        return False
    try:
        size = os.path.getsize(video)
    except OSError:
        return False
    stable = entry.get("size") == size and entry.get("seen_polls", 0) >= 1
    entry["size"] = size
    entry["seen_polls"] = entry.get("seen_polls", 0) + 1
    return stable


def decode_thm_jpeg(path):
    # type: (str) -> Optional[bytes]
    """Return raw JPEG bytes from a .thm (proprietary header, then JPEG at FF D8 FF)."""
    try:
        with open(path, "rb") as f:
            data = f.read()
    except OSError:
        return None
    idx = data.find(b"\xff\xd8\xff")
    return data[idx:] if idx >= 0 else None


def jpeg_to_gray(jpeg_bytes, w=64, h=36):
    # type: (bytes, int, int) -> Optional[bytes]
    """Downscale JPEG to small grayscale via ffmpeg; return raw gray bytes."""
    if not FFMPEG:
        return None
    os.makedirs(TMP_DIR, exist_ok=True)
    src = os.path.join(TMP_DIR, "in-%d.jpg" % os.getpid())
    dst = os.path.join(TMP_DIR, "out-%d.pgm" % os.getpid())
    try:
        with open(src, "wb") as f:
            f.write(jpeg_bytes)
        r = subprocess.run(
            [FFMPEG, "-v", "error", "-y", "-i", src, "-vf", "scale=%d:%d" % (w, h), "-pix_fmt", "gray", dst],
            capture_output=True, timeout=30,
        )
        if r.returncode != 0:
            return None
        raw = parse_pgm(dst)
        return norm_gray(raw) if raw is not None else None
    except (OSError, subprocess.TimeoutExpired):
        return None


def parse_pgm(path):
    # type: (str) -> Optional[bytes]
    with open(path, "rb") as f:
        data = f.read()
    if not data.startswith(b"P5"):
        return None
    parts = data.split(b"\n", 3)
    if len(parts) < 4:
        return None
    return parts[3]


def gray_diff(a, b):
    # type: (bytes, bytes) -> float
    n = min(len(a), len(b))
    if n == 0:
        return 999.0
    total = 0
    for i in range(n):
        total += abs(a[i] - b[i])
    return total / n


def norm_gray(g):
    # type: (bytes) -> bytes
    """Brightness/contrast-normalize a gray thumbnail: cloud shadows and exposure
    ramps shift mean/std but leave structure intact, so normalized diffs ignore them."""
    n = len(g)
    if n == 0:
        return g
    mean = sum(g) / float(n)
    var = sum((v - mean) ** 2 for v in g) / float(n)
    std = var ** 0.5
    if std < 1.0:
        return bytes([128] * n)
    k = 40.0 / std
    out = bytearray(n)
    for i in range(n):
        v = int((g[i] - mean) * k + 128)
        out[i] = 0 if v < 0 else (255 if v > 255 else v)
    return bytes(out)


def token_set(s):
    # type: (str) -> set
    return set(re.findall(r"[a-z0-9]+", s.lower()))


def text_sim(a, b):
    # type: (str, str) -> float
    ta, tb = token_set(a), token_set(b)
    if not ta or not tb:
        return 0.0
    return len(ta & tb) / float(len(ta | tb))


def fp_thumb_match(new_gray, examples):
    # type: (bytes, List[Dict]) -> Optional[Tuple[Dict, float]]
    best = None
    for e in examples:
        g = e.get("gray")
        if not g:
            continue
        d = gray_diff(new_gray, bytes(bytearray(g)))
        if d < FP_THUMB_DIFF and (best is None or d < best[1]):
            best = (e, d)
    return best


def fp_text_match(msg, examples):
    # type: (str, List[Dict]) -> Optional[Dict]
    for e in examples:
        if text_sim(msg, e.get("msg", "")) >= FP_TEXT_SIM:
            return e
    return None


def sample_frames(video, count=3):
    # type: (str, int) -> List[bytes]
    """Extract `count` JPEG frames (768px wide) at 25/50/75% of the clip."""
    frames = []  # type: List[bytes]
    if not (FFMPEG and FFPROBE):
        return frames
    os.makedirs(TMP_DIR, exist_ok=True)
    try:
        r = subprocess.run(
            [FFPROBE, "-v", "error", "-show_entries", "format=duration", "-of", "default=nw=1:nk=1", video],
            capture_output=True, text=True, timeout=30,
        )
        duration = float(r.stdout.strip())
    except (OSError, ValueError, subprocess.TimeoutExpired):
        duration = 0.0
    if duration <= 0:
        duration = 10.0
    stamps = [duration * f for f in (0.25, 0.5, 0.75)][:count]
    for i, t in enumerate(stamps):
        out = os.path.join(TMP_DIR, "frame-%d-%d.jpg" % (os.getpid(), i))
        try:
            r = subprocess.run(
                [FFMPEG, "-v", "error", "-y", "-ss", "%.2f" % t, "-i", video,
                 "-frames:v", "1", "-vf", "scale=768:-2", "-q:v", "4", out],
                capture_output=True, timeout=60,
            )
            if r.returncode == 0 and os.path.getsize(out) > 0:
                with open(out, "rb") as f:
                    frames.append(f.read())
        except (OSError, subprocess.TimeoutExpired):
            pass
    return frames


def llm_chat(content):
    # type: (List[Dict]) -> str
    body = json.dumps({
        "model": LLM_MODEL,
        "messages": [{"role": "user", "content": content}],
        "temperature": 0.0,
        "max_tokens": 1200,
    }).encode("utf-8")
    req = urllib.request.Request(
        LLM_BASE + "/chat/completions",
        data=body,
        headers={"Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=LLM_TIMEOUT) as resp:
        payload = json.loads(resp.read().decode("utf-8"))
    msg_obj = payload["choices"][0]["message"]
    return msg_obj.get("content") or msg_obj.get("reasoning_content") or ""


def llm_analyze(frames, prompt, no_llm, ex_imgs=None):
    # type: (List[bytes], str, bool, Optional[List[Tuple[str, bytes]]]) -> Tuple[str, str]
    """Return (status, message). status in ok|alert|unanalyzed."""
    if no_llm or not frames:
        return "unanalyzed", "no LLM analysis (frames=%d, llm=%s)" % (len(frames), "off" if no_llm else "on")
    content = [{"type": "text", "text": prompt}]
    for caption, img in (ex_imgs or []):
        content.append({"type": "text", "text": caption})
        content.append({
            "type": "image_url",
            "image_url": {"url": "data:image/jpeg;base64," + base64.b64encode(img).decode("ascii")},
        })
    if ex_imgs:
        content.append({"type": "text", "text": "Now judge THIS clip:"})
    for img in frames:
        content.append({
            "type": "image_url",
            "image_url": {"url": "data:image/jpeg;base64," + base64.b64encode(img).decode("ascii")},
        })
    try:
        text = llm_chat(content)
    except (urllib.error.URLError, OSError, ValueError) as e:
        return "unanalyzed", "LLM endpoint unreachable (%s)" % e
    if not text.strip():
        return "unanalyzed", "LLM returned empty response"
    threat, what, why = parse_verdict(text)
    if threat is None:
        return "unanalyzed", "LLM verdict unparsable: %s" % text[:160]
    msg = "%s | %s" % (what, why) if why else what
    if threat == "high":
        return "alert", "[HIGH] %s" % msg
    if threat == "low":
        return "ok", "[low] %s" % msg
    return "ok", msg


def parse_verdict(text):
    # type: (str) -> Tuple[Optional[str], str, str]
    start, end = text.find("{"), text.rfind("}")
    if start < 0 or end <= start:
        return None, "", ""
    try:
        obj = json.loads(text[start:end + 1])
    except ValueError:
        return None, "", ""
    threat = str(obj.get("threat", "")).strip().lower()
    if threat not in ("none", "low", "high"):
        return None, "", ""
    return threat, str(obj.get("what_seen", "")), str(obj.get("why", ""))


def process_cycle(state, args):
    # type: (Dict, argparse.Namespace) -> int
    clips = enumerate_clips()
    known = state.setdefault("clips", {})
    baselines = state.setdefault("baselines", {})
    scenes = load_scenes()
    examples = load_feedback()["examples"]
    pending = 0
    processed = 0

    for rel, camera, when, video in clips:
        entry = known.get(rel)
        if entry is None:
            known[rel] = {"size": -1, "seen_polls": 0, "status": "pending"}
            continue
        if entry["status"] != "pending":
            continue
        if not video_ready(video, entry):
            continue
        if pending >= args.limit:
            continue
        pending += 1
        examples = load_feedback()["examples"]  # re-read: a review mid-cycle applies to the next clip
        prompt = build_prompt(camera, when, scenes, examples)
        ex_imgs = kb_example_images(examples)
        status, msg, frames, thm = analyze_clip(state, camera, when, video, prompt, examples, args, ex_imgs)
        if status == "unanalyzed" and entry.get("retries", 0) < 2:
            entry["retries"] = entry.get("retries", 0) + 1
            entry["seen_polls"] = 1
            log("RETRY %s (attempt %d): %s" % (rel, entry["retries"], msg))
            continue
        entry["status"] = status
        entry["msg"] = msg
        entry["done_at"] = time.time()
        kb_store_clip(rel, when, camera, status, msg, thm, frames)
        processed += 1
        if status == "alert":
            log("ALERT %s: %s" % (rel, msg))
            if not args.dry_run:
                notify("Security alert (%s)" % camera, msg)
        elif status == "unanalyzed":
            log("CANDIDATE (unanalyzed) %s: %s" % (rel, msg))
            if not args.dry_run:
                notify("Motion candidate, unanalyzed (%s)" % camera, rel)
        else:
            log("%s %s: %s" % (status.upper(), rel, msg))
    return processed


def analyze_clip(state, camera, when, video, prompt, examples, args, ex_imgs=None):
    # type: (Dict, str, str, str, str, List[Dict], argparse.Namespace, Optional[List[Tuple[str, bytes]]]) -> Tuple[str, str, List[bytes], Optional[bytes]]
    epoch = 0.0
    try:
        epoch = time.mktime(time.strptime(when, "%Y-%m-%d %H:%M:%S"))
    except ValueError:
        pass

    def set_baseline():
        if new_gray is not None:
            state.setdefault("baselines", {})[camera] = list(new_gray)
            state.setdefault("baseline_time", {})[camera] = epoch or time.time()

    thm = decode_thm_jpeg(thm_path(video))
    new_gray = jpeg_to_gray(thm) if thm else None
    baseline = state.get("baselines", {}).get(camera)
    bt = state.get("baseline_time", {}).get(camera)
    in_event = bool(bt and epoch and abs(epoch - bt) <= EVENT_WINDOW_S)
    thr = EVENT_DIFF_THRESHOLD if in_event else DIFF_THRESHOLD
    if new_gray is not None and baseline is not None:
        d = gray_diff(new_gray, bytes(bytearray(baseline)))
        if d < thr:
            set_baseline()
            return "skip", "repeat clip (diff %.1f < %.1f%s)" % (
                d, thr, ", within event window" if in_event else ""), [], thm
        hit = fp_thumb_match(new_gray, examples)
        if hit is not None:
            set_baseline()
            return "skip", "matches reviewed FP %s (thumb diff %.1f)%s" % (
                hit[0]["clip"].split("/")[-1][:28], hit[1],
                " — " + hit[0]["note"] if hit[0].get("note") else ""), [], thm
    frames = sample_frames(video, count=3)
    status, msg = llm_analyze(frames, prompt, args.no_llm, ex_imgs)
    if status == "alert":
        dup = fp_text_match(msg, examples)
        if dup is not None:
            status = "ok"
            msg = "[downgraded: matches reviewed FP %s] %s" % (
                dup["clip"].split("/")[-1][:28], msg)
    set_baseline()
    return status, msg, frames, thm


def learn_scene(camera, args):
    # type: (str, argparse.Namespace) -> bool
    """Ask the VLM for a permanent-layout description; one day + one night frame."""
    clips = [c for c in enumerate_clips() if c[1] == camera]
    if not clips:
        log("learn-scene: no clips for %s" % camera)
        return False
    day = night = None
    for rel, cam, when, video in reversed(clips):
        try:
            hour = int(when[11:13])
        except (ValueError, IndexError):
            continue
        if day is None and 9 <= hour <= 17:
            day = video
        if night is None and (hour >= 20 or hour <= 5):
            night = video
        if day and night:
            break
    picked = [v for v in (day or clips[-1][3], night) if v]
    frames = []
    for v in picked:
        frames += sample_frames(v, count=1)[:1]
    if not frames:
        log("learn-scene: no frames for %s" % camera)
        return False
    content = [{"type": "text", "text": SCENE_PROMPT}]
    for img in frames:
        content.append({
            "type": "image_url",
            "image_url": {"url": "data:image/jpeg;base64," + base64.b64encode(img).decode("ascii")},
        })
    try:
        text = llm_chat(content).strip()
    except (urllib.error.URLError, OSError, ValueError) as e:
        log("learn-scene: LLM unreachable (%s)" % e)
        return False
    if not text:
        return False
    scenes = load_scenes()
    scenes[camera] = {"text": text, "learned_at": time.time(), "frames": len(frames)}
    save_scenes(scenes)
    log("learn-scene %s: %s" % (camera, text[:200]))
    return True


def mark_feedback(fb, state, clip_substr, label, note):
    # type: (Dict, Dict, str, str, str) -> List[str]
    """Label matching clips; fp corrections become prompt examples."""
    matches = [k for k in state.get("clips", {}) if clip_substr in k]
    for k in matches:
        fb["marks"][k] = {"label": label, "note": note, "at": time.time()}
        if label != "ignore":  # "ignore for now" is review-side parking only — teaches the model nothing
            kb_promote_example(k, label, note)
        if label == "fp":
            msg = state["clips"][k].get("msg", "")
            ex = {"clip": k, "msg": msg, "label": "fp", "note": note}
            thm = decode_thm_jpeg(thm_path(os.path.join(ROOT, k)))
            if thm:
                gray = jpeg_to_gray(thm)
                if gray is not None:
                    ex["gray"] = list(gray)
            fb["examples"] = [ex] + [e for e in fb.get("examples", []) if e.get("clip") != k][:99]
    fb["examples"] = fb.get("examples", [])[:100]
    return matches


def reeval_clip(rel):
    # type: (str) -> Tuple[str, str]
    """Re-run the VLM on one clip with the CURRENT scene map + owner corrections.
    Bypasses dedupe/thumbnail-skip (a manual re-eval always re-analyzes). Stores the
    result as a feedback override + refreshes the knowledge base. Returns (status, msg)."""
    video = os.path.join(ROOT, rel)
    camera = rel.split("/")[0]
    when = clip_start_time(rel.split("/")[-1])
    if not os.path.isfile(video):
        return "error", "clip file missing on disk"
    examples = load_feedback()["examples"]
    prompt = build_prompt(camera, when, load_scenes(), examples)
    ex_imgs = kb_example_images(examples)
    thm = decode_thm_jpeg(thm_path(video))
    frames = sample_frames(video, count=3)
    status, msg = llm_analyze(frames, prompt, False, ex_imgs)
    kb_store_clip(rel, when, camera, status, msg, thm, frames)
    update_feedback(lambda fb: fb.setdefault("overrides", {}).__setitem__(
        rel, {"status": status, "msg": msg, "at": time.time()}))
    return status, msg


def wait_for_next_cycle(interval):
    # type: (float) -> None
    """Sleep until the next poll — or until a review lands, so learning applies immediately."""
    try:
        last = os.path.getmtime(FEEDBACK_PATH)
    except OSError:
        last = 0.0
    deadline = time.time() + interval
    while time.time() < deadline:
        time.sleep(0.5)
        try:
            m = os.path.getmtime(FEEDBACK_PATH)
        except OSError:
            m = 0.0
        if m != last:
            log("review recorded — refreshing learning for the next analysis now")
            return


def main():
    # type: () -> int
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--once", action="store_true", help="single poll pass, then exit")
    ap.add_argument("--interval", type=float, default=60.0, help="poll interval seconds (default 60)")
    ap.add_argument("--limit", type=int, default=5, help="max clips analyzed per cycle (default 5)")
    ap.add_argument("--no-llm", action="store_true", help="skip LLM calls; only pre-filter and log")
    ap.add_argument("--dry-run", action="store_true", help="no macOS notifications")
    ap.add_argument("--reset", action="store_true", help="clear state.json and exit")
    ap.add_argument("--learn-scene", action="store_true", help="build/refresh scene maps, then exit")
    ap.add_argument("--mark", metavar="SUBSTRING", help="label matching clips in state (fp|tp), then exit")
    ap.add_argument("--label", choices=["fp", "tp"], default="fp", help="label for --mark (default fp)")
    ap.add_argument("--note", default="", help="owner note for --mark, injected into future prompts")
    args = ap.parse_args()

    if args.reset:
        if os.path.exists(STATE_PATH):
            os.remove(STATE_PATH)
        print("state cleared")
        return 0
    if args.mark:
        fb = load_feedback()
        matches = mark_feedback(fb, load_state(), args.mark, args.label, args.note)
        save_feedback(fb)
        if not matches:
            print("no clips match %r" % args.mark)
            return 1
        for k in matches:
            print("%s -> %s" % (k, args.label))
        return 0
    if not FFMPEG or not FFPROBE:
        print("ffmpeg/ffprobe not found (PATH or /opt/homebrew/bin)", file=sys.stderr)
        return 1
    if not os.path.isdir(ROOT):
        print("capture root missing: %s (is /Volumes/aaron mounted?)" % ROOT, file=sys.stderr)
        return 1

    state = load_state()
    state["config"] = {"llm_base": LLM_BASE, "model": LLM_MODEL}  # review server inherits this
    save_state(state)
    log("watching %s (llm=%s model=%s threshold=%.1f)" % (ROOT, "off" if args.no_llm else LLM_BASE, LLM_MODEL, DIFF_THRESHOLD))
    cameras = sorted({c[1] for c in enumerate_clips()})
    if args.learn_scene:
        return 0 if all([learn_scene(c, args) for c in cameras]) else 1
    scenes = load_scenes()
    if not args.no_llm:
        for cam in cameras:
            if cam not in scenes:
                log("no scene map for %s — learning it now" % cam)
                learn_scene(cam, args)
    while True:
        try:
            process_cycle(state, args)
        except Exception as e:  # keep the watcher alive
            log("cycle error: %r" % (e,))
        save_state(state)
        if args.once:
            return 0
        wait_for_next_cycle(args.interval)


if __name__ == "__main__":
    sys.exit(main())
