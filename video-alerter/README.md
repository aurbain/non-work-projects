# video-alerter

Watches SecuritySpy clips on the live `securityspy` capture root (local disk on the NAS; auto-detected mount elsewhere — see CLAUDE.md),
samples 8 motion-peak frames per clip (plus a magnified motion-region crop, IR boost at night), classifies them with a local vision LLM
via 5-vote self-consistency, and raises an alert for anything rated HIGH. Runs on **tnas.local** (Ubuntu 22.04, Python 3.10, stdlib only)
as systemd services; the VLM runs on kahuna over the network. See `CLAUDE.md` for layout and constraints.

## 1. LLM server (on kahuna)

**Primary: MLX-Serve.app on :11234, model `lmstudio-community/Qwen3-VL-30B-A3B-Instruct-MLX-8bit`** — GUI-managed, host 0.0.0.0, LAN-shared; lazy-loads models by ID from its model dirs, up to 4 resident (idle-evict 4 h). Nothing to start if the app is running.

Check: `curl http://kahuna.local:11234/v1/chat/completions -d '{"model":"lmstudio-community/Qwen3-VL-30B-A3B-Instruct-MLX-8bit","messages":[{"role":"user","content":"ping"}],"max_tokens":4}'` — first request may cold-load the model for a minute.

Legacy fallback (uninstalled 2026-11-01): uvx `mlx-serve` on :8095, model `qwen3-vl-30b-a3b` / `qwen2.5-vl-3b` (config `~/mlx-serve/models.yaml`). Reinstall + start:

```sh
ssh aaron@kahuna.local
/opt/homebrew/bin/uvx --python 3.13 --from mlx-serve --with mlx-vlm mlx-serve start --host 0.0.0.0 --port 8095
```

## 2. Run the watcher

**Home is `tnas.local`** (Ubuntu 22.04, `/Volume1/SecuritySpy/alerter`) as two systemd units — env in `/etc/alerter.env`:

```sh
ssh aaron@tnas.local
systemctl status video-alerter video-alerter-review
journalctl -u video-alerter -f                 # live watcher log
systemctl restart video-alerter                # after editing code (rsync from dev Mac first)
# review UI: http://192.168.4.130:8765        (use the IP — tnas.local resolves IPv6 first)
```

Dev-loop from the Mac: edit → `rsync -a --exclude tmp/ --exclude __pycache__/ --exclude .git/ --exclude state.json --exclude feedback.json --exclude 'alerter.db*' --exclude scene.json --exclude knowledge/ --exclude alerts.log --exclude eval_history.jsonl ./ aaron@tnas.local:/Volume1/SecuritySpy/alerter/` → restart the unit. **Never rsync data files (`state.json`, `alerter.db`, `knowledge/`, …) toward the NAS** — the TNAS copies are authoritative; the Mac copies are stale mirrors (an un-excluded rsync silently rolled back `feedback.json` on 2026-10-31).

Backup of learned knowledge (tracked in git: `alerter.db`, `scene.json`, `knowledge/clips.jsonl`, example thumbs): periodically `rsync -a aaron@tnas.local:/Volume1/SecuritySpy/alerter/{alerter.db,scene.json,alerts.log} . && rsync -a aaron@tnas.local:/Volume1/SecuritySpy/alerter/knowledge/clips.jsonl knowledge/ && rsync -a aaron@tnas.local:/Volume1/SecuritySpy/alerter/knowledge/examples/ knowledge/examples/ && git add -A && git commit`.

Manual/ad-hoc runs (any machine that can see the capture root + kahuna):

```sh
ALERTER_ROOT=/Volume1/SecuritySpy python3 alerter.py
```

Useful flags: `--once` (single pass), `--limit N` (max clips per cycle, default 5), `--dry-run`
(no notifications), `--no-llm` (prefilter only), `--reset` (forget state, re-scan everything).

New clips need two stable polls before analysis (guards against reading a mid-write file).
First run backfills history at `--limit` clips/cycle; raise `--limit` for a faster backfill.

## 3. Learning (the watcher gets better over time)

- **Scene map**: on first run (or `--learn-scene`) the watcher sends one day + one night frame per
  camera to the VLM and stores the permanent layout in `scene.json` (garage left, public road along
  the bottom, day/night appearance…). Every later prompt includes it, so "person on the road" vs
  "person at your garage" is judged with spatial knowledge. Re-run `--learn-scene` after landscaping,
  seasons, or camera moves.
- **Owner feedback**: when an alert was a false positive, label it:
  ```sh
  python3.15 alerter.py --mark "09-30-2026 18-44" --label fp --note "neighbor walking home, daily"
  ```
  Up to 8 corrections (newest first) are injected into future prompts, so the model stops repeating
  the mistake. `--label tp` records confirmations (kept in state, not injected).

## 4. Review UI (after-the-fact browsing)

```sh
python3.15 review.py     # http://127.0.0.1:8765 — set ALERTER_REVIEW_HOST=0.0.0.0 to share on LAN
```

Per-day browse: date chips, thumbnail per clip (click for a full-clip lightbox — `Esc` closes), suspected
alerts sorted to the top, status badge (alert/ok/skip/pending/unanalyzed), the alert text overlaid on the
frame, the camera's learned scene map, and five actions per clip:
- **false positive / real alert** — become prompt examples immediately (`feedback.json`, re-read every cycle).
- **this is resident** — pins the clip's thumbnail as an **owner reference card** (max 5, newest first):
  its image + your note (e.g. "Bill, the homeowner") are injected into *every* future prompt as
  *"Owner-approved resident — NEVER a threat"*, exempt from the FP recency rotation. This is how the
  system learns a specific person: unlike FP examples (which age out of the 3-image/5-text windows),
  owner cards never leave the prompt until you un-pin them. CLI: `--mark SUB --label owner --note "Bill"`.
- **ignore for now** — parks the clip (audit-trail mark, hidden from the default view) *without* teaching
  the model anything: no prompt example, no KB promotion. It stays in the learn-mode rotation, so a later
  re-eval with fresh knowledge can still resolve it.
- **re-evaluate** — re-runs the VLM on that clip *with the current knowledge* (scene map + all owner
  corrections + FP example images). Useful after you've marked a batch of FPs: re-eval an old alert and
  watch it downgrade. Result is stored as an override in `feedback.json` (the watcher owns `state.json`,
  so this keeps each file single-writer) and refreshes the clip's `knowledge/` record. The review server
  inherits the LLM endpoint from `state.json["config"]`, so you don't need to export `ALERTER_LLM_BASE` twice.
- **learn mode** (header toggle) — automatic re-evaluation of the whole corpus. When enabled, the review
  server runs a background worker that re-judges one clip every `ALERTER_LEARN_INTERVAL` seconds
  (default 15) with the *current* knowledge, round-robin by oldest re-eval; never-re-evaluated alerts go
  first, and alerts rotate ~2x afterward, so fresh corrections propagate to the clips that matter soonest.
  Results land in the same `feedback.json["overrides"]` as manual re-evals; the day view auto-refreshes
  while the toggle is on. The flag persists in `feedback.json["learn_mode"]` (survives restarts) — turn it
  off when the corpus has converged to stop the VLM churn. `GET /api/learn` reports progress
  (done/fail/current clip); failures on missing files are cooled down for 30 min instead of hot-looping.

CLI equivalent of marking: `alerter.py --mark SUB --label fp --note …`.

## 5. Output

- macOS notification (sound) for HIGH threat, and for candidates the LLM could not analyze.
- `alerts.log` — every decision, one line per clip.
- `state.json` — per-clip status (`pending|skip|ok|alert|unanalyzed|error`) + per-camera thumbnail
  baselines. Delete or `--reset` to reprocess.

## Env vars

| Var | Default | Meaning |
|---|---|---|
| `ALERTER_ROOT` | `/Volumes/aaron/SecuritySpy_secondary` | capture root |
| `ALERTER_LLM_BASE` | `http://kahuna.local:11234/v1` | OpenAI-compatible base (MLX-Serve.app on kahuna) |
| `ALERTER_LLM_MODEL` | `qwen3-vl-30b-a3b` | model id the server accepts |
| `ALERTER_LLM_TIMEOUT` | `180` | per-call seconds |
| `ALERTER_DIFF_THRESHOLD` | `12.0` | thumbnail-diff below this = repeat clip, skipped (dedupe only, see CLAUDE.md) |
| `ALERTER_REVIEW_PORT` / `ALERTER_REVIEW_HOST` | `8765` / `127.0.0.1` | review server bind |
| `ALERTER_STATE` / `ALERTER_LOG` / `ALERTER_SCENE` / `ALERTER_FEEDBACK` | files in project dir | override data paths |
| `ALERTER_FRAMES` | `8` | frames sampled per clip (motion-peak, then uniform fill) |
| `ALERTER_MOTION_PEAKS` | `1` | `0` disables scene-score peak sampling (pure uniform) |
| `ALERTER_ZOOM_CROP` | `1` | `0` disables the magnified motion-region crop |
| `ALERTER_NIGHT_BOOST` | `1` | `0` disables the IR/night contrast-boost on near-monochrome clips |
| `ALERTER_VOTES` | `5` | self-consistency samples per clip (`1` = single shot, no voting) |
| `ALERTER_VOTE_TEMP` | `0.7` | sampling temperature for votes |
| `ALERTER_VOTE_QUORUM` | `1.0` | fraction of votes that must say HIGH to alert (`1.0`=unanimous, `0.6`=3/5) |
| `ALERTER_EVAL` | `eval_history.jsonl` | where `--eval` appends confusion matrices |

## Measuring accuracy — `--eval`

The golden set is whatever you've marked in the review UI (fp/tp/owner). `--eval` re-judges those clips
with the *current* prompt + pipeline and scores against your labels, appending a confusion matrix to
`eval_history.jsonl`:

```sh
python3.15 alerter.py --eval
# EVAL n=118 tp=9 fp=4 tn=99 fn=6 err=0  precision=0.69 recall=0.60  (41.2 min)
#   MISS <clip> [tp] got=ok  ...   <- real threats the pipeline now misses (over-suppression)
#   MISS <clip> [fp] got=alert ... <- false positives it still can't kill
```

`tp/fp` are measured on the *owner's* definition of a real alert (a `tp` mark). Run it before and after any
prompt/threshold/pipeline change — it's the only honest way to know a tweak helped. `fn` (missed real
threats) is the number to guard hardest in a security system; if it rises, lower `ALERTER_VOTE_QUORUM`.

## Autostart (later)

`launchd` plists for both the mlx-serve server and `alerter.py` once the flow is trusted; until then
start the two commands manually.
