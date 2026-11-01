# video-alerter

Goal: watch motion-capture videos produced by SecuritySpy and alert me when something alarming happens. Software runs from this project on **either** analysis machine below — keep it portable: same code paths, no machine-specific logic except what is read from config/env.

| Machine | What it is | Notes |
|---|---|---|
| `tnas.local` (192.168.4.130) | **Current analysis host** — Ubuntu 22.04 server, 4 cores, 8 GB, x86_64, root=uid0. SSH: `ssh aaron@tnas.local` (key auth set up). Runs watcher + review + eval. | Home of the capture tree on **local disk** (`/Volume1/SecuritySpy`) — no SMB. Repo at `/Volume1/alerter`. systemd units `video-alerter` + `video-alerter-review`, env in `/etc/alerter.env`. Review UI: `http://192.168.4.130:8765` (use the IP; `tnas.local` resolves IPv6 first and the server binds IPv4). Modest box + it's the NAS — keep it nice (nice/lock/throttle). `notify()` is macOS-only, so alerts are log+UI until a Linux channel is chosen. |
| `kahuna.local` | Apple M5 Ultra, 36 cores, 256 GB RAM, macOS 27.2, 1.8 TB disk | The **LLM box** — mlx-serve :8095 (qwen3-vl-30b-a3b). Reached over the network from tnas. SSH: `ssh aaron@kahuna.local`. |
| `intel-macmini.local` | Intel i5-8500B, 32 GB, macOS 15.8.2 | Runs **SecuritySpy** (the capture host). SSH: `ssh aaron@intel-macmini.local`. Do not load this box with analysis. |
| `TDI-MBPro40` (this dev Mac) | Apple Silicon, 36 GB RAM, macOS 27.2 | Where the code is written and debugged; same `/Volumes/aaron` mount. No longer runs the watcher (state.json now lives on tnas). |

**Compute moved to the NAS (2026-11-01)** at owner request: watcher + review + eval run on `tnas.local`, reading clips from local `/Volume1/SecuritySpy` (one sequential read per clip via `stage_local`, then all ffmpeg/frame work is local; cross-process `ffmpeg.lock` keeps concurrency at 1; `nice 15`). Only the VLM call crosses the network (to kahuna). TNAS is 4-core and just came off a ZFS scrub — keep throttles gentle (watcher `--interval 240`, learn 60s, eval 60s) until the scrub is done, then dial back.

## Where the video files live

SecuritySpy (running on `intel-macmini.local`) records to the NAS share `//aaron@TNAS(SMB)/securityspy` — **this is the live capture root** (has every day, incl. the newest). Mounted as: `/Volumes/SecuritySpy` on the dev Mac, `~/Volumes/securityspy` on kahuna (its `/Volumes` is SIP-protected; new mounts need a home-dir mount point). The old `SecuritySpy_secondary` folder on the `aaron` share is **stale** (stopped 10-27) — kept as last-resort fallback only. `alerter.py` auto-detects the first existing root (`ALERTER_ROOT` overrides); relative clip keys (`camera/date/file.m4v`) are identical across mounts, so state carries over.

Layout:

```
<Vol>/SecuritySpy/                                # live capture root (SMB share "securityspy")
└── amcrest outdoor bullet right/                # one dir per camera
    ├── 2026-10-29/                              # one dir per calendar day (YYYY-MM-DD)
    │   ├── 10-29-2026 08-56-48 M amcrest outdoor bullet right.m4v   # the recording
    │   ├── .10-29-2026 08-56-48 M ... .thm      # hidden JPEG-ish thumbnail (~110 KB)
    │   └── .FF_Index                            # hidden per-day index, ignore
    └── ...
```

Key facts:

- **Only one camera today**: `amcrest outdoor bullet right` (outdoor). New cameras appear as sibling directories under `SecuritySpy_secondary/` — glob for them, do not hardcode.
- **Video files**: `.m4v`, H.264, sizes ~10–55 MB each, one file per motion event. Filenames encode the start time: `MM-DD-YYYY HH-MM-SS <mode> <camera name>.m4v`.
- **Thumbnails**: hidden `.<same name>.thm` (~110 KB) accompany every recording. Cheap first-pass signal for an LLM/vision model — read the `.thm` before decoding any video. Format: 132-byte proprietary header, then a plain baseline JPEG (~640x361) — skip to the first `FF D8 FF` marker and decode from there.
- **`.FF_Index`**: SecuritySpy bookkeeping. Never parse or touch it.
- Capture is event-driven, not continuous. Measured load (Oct 2026): **median ~38 clips/day, max 85, peak hour 63 clips/h (~1/min, 09:00–17:00)**; near-zero overnight. September's 1–6/day is stale — the camera got much busier. Design for ~100 clips/day with minute-scale bursts, not trickle.
- Total captures so far ≈ 6 GB. Do not bulk-copy or checksum the tree; it is on SMB and slow.

## Mount gotchas (SMB)

- `/Volumes/aaron` is a network mount: it can be stale or unmounted. Check `mount | grep aaron` / stat the directory before assuming "no new files". If missing, remount: `mount_smbfs '//aaron@TNAS(SMB)._smb._tcp.local/aaron' /Volumes/aaron` (credentials are in the macOS keychain).
- SMB + many small files is slow. Enumerate with one `ls`/`find` per day dir, not per file. Filenames contain spaces — always quote.
- Never write into the capture tree. It is SecuritySpy's; a new file appearing mid-write looks complete but isn't — treat a file as ready only after its `.thm` exists and size is stable for one poll interval.

## Machine notes

- Both analysis machines are Apple Silicon macOS; `/Volumes/aaron` is mounted on both. Detect at startup (mount present?) rather than branching on hostname.
- kahuna.local: SSH is `ssh aaron@kahuna.local` (alias `kahuna.local` in `~/.ssh/config`; user is `aaron`, NOT the local username). Homebrew installed at `/opt/homebrew`; in non-login SSH shells use full paths (`/opt/homebrew/bin/ffmpeg`). Has `ffmpeg`/`ffprobe` 9.x.
- **Python**: run the watcher on the latest stable — **3.15** (`/opt/homebrew/bin/python3.15` on both analysis machines; system `/usr/bin/python3` is 3.9, do not target it). Keep the watcher **stdlib-only** so it needs no venv. **3.13** (`/opt/homebrew/bin/python3.13`) is only for the uvx `mlx-serve` fallback venv, where MLX wheels lag the newer interpreters.
- LLM servers already running on kahuna (installed, do not re-install): **MLX-Serve.app** (v26.10.3) — OpenAI-compatible on **:11234**, host 0.0.0.0, LAN-shared, currently loads `ddalcu/Qwen3.8-Flash-Next-MLX-Serve-iQ-MLX-4.7bpw` (MoE 4-bit, capabilities: chat/vision/video/json_schema, 262k ctx; model dirs `~/Desktop/mlx-llms`, `~/.mlx-serve/models`, `~/Desktop/lmstudio-llms`, `~/Desktop/omlx-llms`). **oMLX.app** (v0.7.3.dev2, `app.omlx`) also runs with its own server. Shared volume `/Volumes/lmstudio-llms/` holds `mlx-community/Qwen2.5-VL-3B-Instruct-8bit` (used by the uvx `mlx-serve` fallback on :8095, config `~/mlx-serve/models.yaml`).
- Dev Mac: Homebrew `ffmpeg`/`ffprobe` 9.x and `python3` 3.15 installed. Frame extraction pattern: `ffprobe` for duration, then `ffmpeg -ss <t> -frames:v 1` at a few sampled timestamps.

## Design constraints for LLM-written code

- SecuritySpy recording has priority on kahuna: cap ffmpeg concurrency at 1. Evidence is now richer by owner choice ("we have lots of compute"): 8 motion-peak-sampled frames (ffmpeg scene scores, uniform fill), night/IR contrast boost, and a magnified crop of the motion region per clip.
- Self-consistency voting: each clip judged `ALERTER_VOTES` (default 5) times at temp 0.7; alert requires `ALERTER_VOTE_QUORUM` (default 1.0 = unanimous HIGH). This crushes the stochastic FP tail. Guard: high quorum raises FN (missed real threats) — measure with `--eval` and lower the quorum if `fn` climbs. Pre-filter (dedupe/FP suppression) still runs first so most clips never reach the LLM.
- LLM calls are cheap relative to available compute but still pre-filtered: dedupe and only send clips that pass to a vision model. Thumbnail diffs are **brightness/contrast-normalized** (`norm_gray`) so cloud-shadow/exposure ramps don't masquerade as events (same frame under a 30% darkening: raw diff 65 → normalized 6). The remaining median ~35 between clips is genuine content change, so dedupe (threshold 12; 20 within a 3-min event window) only trims repeat segments, not real motion. Assume most clips reach the LLM.
- `--eval` (golden set = every fp/tp/owner mark) re-judges marked clips with the current pipeline and appends a confusion matrix + misses to `eval_history.jsonl`. Measure before/after any prompt/threshold/pipeline change; `fn` (missed real threats) is the metric to guard in a security system.
- Wind and weather: environmental motion is structural (leaves genuinely move), so it can't be deduped away — the prompt hard rule ("environmental motion is threat=none no matter how much of the frame is moving") + scene map vegetation zones + owner FP feedback are the defense. If windy days still produce HIGH verdicts, that's feedback-loop material.
- Persist state (last-seen file per camera) in this project dir (e.g. `state.json`); the capture tree is read-only ground truth.
- Alerting goes to me — prefer macOS notification + a log line; any push channel gets chosen later.
- Learning is prompt-side, not weights, backed by a disk knowledge base:
  - `scene.json` — per-camera permanent-layout description (auto-learned when missing, `--learn-scene` to refresh).
  - `knowledge/clips/<hash>/` — every analyzed clip keeps its thumb, 3 sampled frames, and verdict (`meta.json`), plus an append-only `clips.jsonl`. `knowledge/examples/` holds the same evidence for every reviewed (FP/TP) clip. Ceiling `ALERTER_KB_MAX_MB` (default 20 GB; ~1 MB/clip → years of history).
  - `feedback.json` — marks (audit trail) + up to 100 FP corrections with normalized-thumbnail vectors for deterministic suppression + `owners` (≤5 pinned resident reference cards). Cards mean **owner-approved people** (household, neighbors, workers — e.g. "Bill the neighbor" + his worker): caption + prompt text say NEVER a threat; UI button is "known person". Clips marked fp/owner whose effective status is `alert` get force-downgraded to `ok` with a `[owner-labeled FP: …]` override — owner label beats the model.
  - Prompt injection per call: scene map + owner cards (text for all ≤5 + up to 3 pinned **images** captioned "Owner-approved resident … NEVER a threat", exempt from recency rotation — this is identity learning) + 5 newest text corrections + up to 3 FP example **images** (captioned "Owner-labeled FALSE POSITIVE …").
- Owner corrections: `alerter.py --mark SUB --label fp|tp|ignore|owner --note …` or the review UI buttons (same code path; fp/tp/owner also promote the clip's evidence into `knowledge/examples/`; `ignore` teaches nothing; `owner` pins the resident card).
- **File ownership invariant**: the watcher is the ONLY writer of `state.json`; the review server is the ONLY writer of `feedback.json`. Review actions (marks, re-evals) therefore write `feedback.json`, never `state.json`. A re-eval (`POST /api/reeval`, re-runs the VLM with current knowledge) stores its result in `feedback.json["overrides"][clip]`; the review UI shows the override in place of the state status. The watcher records its `config` (llm_base, model) in `state.json` at startup so the review server inherits the same endpoint.

## LLM requirements (what must be running on the analysis machine)

- **Primary endpoint: mlx-serve on kahuna :8095, model `qwen3-vl-30b-a3b`** — `Qwen3-VL-30B-A3B-Instruct` (vision-first MoE, DeepStack enabled, ~3B active so fast), served by uvx `mlx-serve` (config `~/mlx-serve/models.yaml`, `hf_path` = `/Users/aaron/Desktop/mlx-llms/mlx-community/Qwen3-VL-30B-A3B-Instruct-bf16`). This is the alerter's default (`ALERTER_LLM_BASE=http://127.0.0.1:8095/v1`, `ALERTER_LLM_MODEL=qwen3-vl-30b-a3b`). Restart mlx-serve after editing models.yaml.
- **Fallbacks:** same mlx-serve, model `qwen2.5-vl-3b` (`Qwen2.5-VL-3B-Instruct-8bit`, `/Volumes/lmstudio-llms/...`); or MLX-Serve.app on :11234 (`ddalcu/Qwen3.8-Flash-Next...` — a reasoning LLM with a bolted-on vision tower, weaker at small/night objects, and it shares inference with the coding session). Set the two env vars to switch.
- **Vision model is required** — must accept base64 `image_url` content parts. Qwen3-VL is chosen specifically for small-object/night-IR recall that the Flash-Next model's non-DeepStack vision tower misses.- Text model (`Qwen3.5-9B`) is optional — only for a later summarizer/deduper of alerts.
- The watcher degrades gracefully: if no LLM endpoint answers, it still does motion pre-filtering, logs candidate clips, and alerts "unanalyzed" rather than crashing.
