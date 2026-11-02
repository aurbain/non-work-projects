# video-alerter

Goal: watch motion-capture videos produced by SecuritySpy and alert the owner when something alarming happens. Stdlib-only Python (3.10+ on the NAS, 3.15 on Macs); keep it portable — no machine-specific logic except config/env.

## Machines

| Machine | Role | Notes |
|---|---|---|
| `tnas.local` / **192.168.4.130** | **Analysis host** — TerraMaster F4-425 Plus, TOS 7 (Ubuntu 22.04 base), 4-core, 8 GB, aaron=uid0. SSH key auth. | Runs watcher + review UI + eval as systemd units `video-alerter` / `video-alerter-review`; env `/etc/alerter.env`. Capture tree on **local NVMe** `/Volume1/SecuritySpy` (no SMB here). Repo copy: `/Volume1/SecuritySpy/alerter` (inside the capture root — `enumerate_clips` skips it). Review UI: **http://192.168.4.130:8765** (use the IP; `tnas.local` resolves IPv6, server binds IPv4). Use the IP for ssh too if mDNS is flaky. Quirks: `systemctl restart` sometimes fails "Interactive authentication required" — run `systemctl daemon-reload` first. Ad-hoc shells need `set -a; . /etc/alerter.env; set +a`. |
| `kahuna.local` | **LLM host only** | MLX-Serve.app (GUI-managed, :11234, LAN) serves the alerter VL model **and** coding models. Lazy-loads by model ID, up to 4 resident, idle-evict 4 h. SSH `aaron@kahuna.local`; ffmpeg at `/opt/homebrew/bin`. |
| `intel-macmini.local` | **Capture host** (SecuritySpy) | Records to SMB share `//aaron@TNAS(SMB)/securityspy` → `/Volume1/SecuritySpy` on the NAS. Do not load with analysis. Camera is 192.168.1.123 on its own NIC (en1); RTSP port 554. |
| dev Mac (repo home) | Coding only | No watcher. |

**Dev loop:** edit locally, then rsync **code only** — always exclude the data files; TNAS copies are authoritative (a 2026-10-31 rsync without excludes rolled back the learned feedback):

```
rsync -a --exclude tmp/ --exclude __pycache__/ --exclude .git/ --exclude state.json --exclude feedback.json \
  --exclude 'alerter.db*' --exclude scene.json --exclude knowledge/ --exclude alerts.log --exclude eval_history.jsonl \
  ./ aaron@192.168.4.130:/Volume1/SecuritySpy/alerter/
```

Restart services after deploy (`systemctl daemon-reload; systemctl restart video-alerter video-alerter-review`).

## LLM

- **Endpoint:** `http://kahuna.local:11234/v1` (OpenAI-compatible). **Model:** `lmstudio-community/Qwen3-VL-30B-A3B-Instruct-MLX-8bit` — vision-first MoE, DeepStack; chosen for night-IR/small-object recall. Configured in `/etc/alerter.env` (`ALERTER_LLM_BASE`, `ALERTER_LLM_MODEL`); built-in defaults match.
- The old uvx `mlx-serve` manager (:8095) was **uninstalled 2026-11-01**; reinstall recipe in README §1 if ever needed.
- **pi** (coding agent) on the TNAS also uses kahuna: provider `mlx-serv` in `~/.pi/agent/models.json`, default model `ddalcu/Qwen3.8-Flash-Next-MLX-Serve-iQ-MLX-4.7bpw`, `tuiMode: regular`.
- Watcher degrades gracefully: no LLM → clips logged as `unanalyzed` candidates, never crashes.

## Storage state of the TNAS (2026-11-02)

- `/Volume1` = WD_BLACK SN770 2 TB NVMe (vg1-lv0, ext4): capture tree + alerter + homes. **Healthy. All system data lives here.**
- `/Volume2` = 4× HDD RAID5 (vg0-lv1, btrfs): **DAMAGED**. Hyper Cache (Fanxiang 256 G NVMe) turned out to be **writeback with bad blocks**; the forced cache detach (TOS GUI + `lvconvert --uncache --force`) abandoned unflushed dirty blocks and btrfs metadata went with them. Volume2 mounts only read-only with `rescue=ignorebadroots,nologreplay,...`. **Treat as read-only until rescued, then TOS-GUI delete + recreate** — recreation claims the whole 16.3 T pool (10.9 T currently unallocated), so expansion and reformat are one step. Do NOT mkfs from CLI (TOS expects its subvolume layout). Never mount rw.
- Hyper Cache is gone permanently (drive physically removed). The USB-C port is **20 Gbps (Gen 2×2)** — use it for rescue copies; the Type-A ports and the currently-plugged TerraMaster dock negotiate USB 2 (40 MB/s) — replug into the C port.
- `md1` RAID5 = the HDDs (Volume2); `md1` naming shuffles across boots — match by member disks, not names. 263 clips in state have no video file on the NAS (migrated from the old SMB root; some days are thumbnails-only) — the learn loop blacklists them permanently.

## Where the video files live

Capture root = `/Volume1/SecuritySpy` on the NAS (SMB share "securityspy"; mounted `/Volumes/SecuritySpy` on Macs, `~/Volumes/securityspy` on kahuna). `ALERTER_ROOT` overrides; `alerter.py` auto-detects the first existing root; relative clip keys (`camera/date/file.m4v`) are stable across mounts.

```
<root>/<camera name>/<YYYY-MM-DD>/MM-DD-YYYY HH-MM-SS <mode> <camera>.m4v   # one file per motion event (10–55 MB, H.264)
                              .<same name>.thm                              # hidden ~110 KB thumbnail: 132-byte header, then plain JPEG (skip to FF D8 FF)
                              .FF_Index                                     # SecuritySpy bookkeeping — never touch
```

- One camera today: `amcrest outdoor bullet right` — glob for siblings, never hardcode.
- Load: median ~38 clips/day, bursts to ~1/min daytime; design for ~100/day.
- Treat a file as ready only after its `.thm` exists and size is stable one poll. **Never write into the capture tree.**

## Pipeline (alerter.py)

1. `enumerate_clips` (skips the project dir) → per-clip readiness check.
2. Pre-filter: brightness/contrast-**normalized** thumb diff (`norm_gray`, thresholds 12 / 20 within a 3-min event window) dedupes repeat segments; `fp_thumb_match` suppresses clips within `FP_THUMB_DIFF` of a learned FP vector.
3. Evidence (`extract_evidence`, one ffmpeg at a time via cross-process `tmp/ffmpeg.lock`, `nice 15`): stage locally, probe duration, 8 motion-peak-sampled frames (scene-score picks), night/IR contrast boost, magnified crop of the motion region.
4. Verdict: `llm_analyze_vote` — `ALERTER_VOTES` (5) votes at temp 0.7; alert requires `ALERTER_VOTE_QUORUM` (default 1.0 = unanimous HIGH). Prompt: scene map + owner cards + FP corrections.
5. **Identity discipline (2026-11-02):** the prompt forbids naming a person without a visible distinguishing feature, resident cards excuse *presence* but never *dangerous actions*, and "not X" owner notes override cards. When a verdict claims a resident's name, `identity_verify` runs a **neutral** SAME/DIFFERENT/UNSURE check against the reference images; DIFFERENT/UNSURE re-judges with all cards voided and prefixes `[identity unverified: …]`. (Motivation: the model was confabulating "Bill" for the owner's own Subaru.)
6. `owner_label_override`: any clip the owner marked fp/owner is force-downgraded from alert — **the owner's label always beats the model**.
7. KB (`knowledge/`): every clip's thumb+frames+verdict under `clips/<hash>/` with `meta.json` (incl. `ev` = evidence layout + `EVIDENCE_VERSION`, and the thumb gray vector). **`kb_evidence` cache: re-judges reuse cached frames — zero ffmpeg** unless the version bumped. `examples/` holds promoted FP/TP/owner evidence; owner cards pin ≤3 reference images.
8. Alerting: `notify()` — osascript on macOS, **SMTP email** when configured: `ALERTER_SMTP_HOST/PORT/USER`, `ALERTER_MAIL_FROM/TO`, clip attached if ≤ `ALERTER_MAIL_MAX_MB` (20) else thumbnail+path. Gmail setup in `/etc/alerter.env`; **app password must live in `/etc/alerter-smtp.pass` (root-only, created by the owner — never in chat/logs)**. Mail failures never kill the watcher.
9. `--eval` = golden set (every fp/tp/owner mark) re-judged with the current pipeline → confusion matrix + misses appended to `eval_history.jsonl`. **`fn` (missed real threats) is the metric to guard.** Measure before/after every prompt/pipeline change.

## Runtime data — alerter.db (SQLite, WAL)

`state.json` and `feedback.json` were migrated into **`alerter.db`** (2026-11-01; `*.pre-sqlite` backups kept). Tables: `clips` (watcher-owned statuses/baselines), `marks`, `examples`, `overrides`, `owners`, `meta`. Row `data` column holds the full JSON entry. All writes go through `alerter.update_feedback()` / `save_state()` — one `BEGIN IMMEDIATE` txn each, cross-process safe (watcher/review/CLI/mark-path all write). `reset_state()` re-enumerates. Review server inherits the watcher's recorded `config` (llm_base/model) from meta.

**Tracked in git** (this repo, root = `non-work-projects`): `alerter.db`, `scene.json`, `knowledge/clips.jsonl`, `knowledge/examples/*/meta.json|thumb.jpg`. Pull from TNAS + commit periodically — that's the backup:
```
rsync -a aaron@192.168.4.130:/Volume1/SecuritySpy/alerter/{alerter.db,scene.json,knowledge/clips.jsonl} video-alerter/  # + examples
```

## Review UI (review.py)

http://192.168.4.130:8765 — newest-day-first chips (default = today), alerts filter, thumbnail→clip player, verdict text with `judged <time>` (re-eval stamped), buttons: false positive / real alert / **known person** (pins owner card) / ignore / **re-evaluate** (shows progress + timestamped result). Re-eval stores overrides; the UI shows the override in place of the original. Learn worker (`ALERTER_LEARN_INTERVAL`, 60 s in prod): round-robin re-judge, alerts ×2 weight, **clips similar to learned FP corrections jump the queue**, missing-evidence clips blacklisted.

## Throttles & etiquette

TNAS is a 4-core NAS: watcher `--interval 240`, learn 60 s, `ALERTER_EVAL_DELAY` ≥3. One ffmpeg at a time globally. Raise pace only deliberately and revert after. `notify()` email is the alert channel of record; log line always (`alerts.log` + journald).
