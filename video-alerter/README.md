# video-alerter

Watches SecuritySpy clips on `/Volumes/aaron`, samples 3 frames per clip, classifies them with a local
vision LLM, and raises a macOS notification for anything rated HIGH threat. Runs on kahuna.local or the
dev Mac (Python 3.15, stdlib only). See `CLAUDE.md` for layout and constraints.

## 1. LLM server (on kahuna)

**Primary: uvx `mlx-serve` on :8095, model `qwen3-vl-30b-a3b`** — `Qwen3-VL-30B-A3B-Instruct`
(vision-first MoE, DeepStack) from `/Users/aaron/Desktop/mlx-llms/mlx-community/Qwen3-VL-30B-A3B-Instruct-bf16`.
Config `~/mlx-serve/models.yaml`; start:

```sh
ssh aaron@kahuna.local
cd ~/mlx-serve
/opt/homebrew/bin/uvx --python 3.13 --from mlx-serve --with mlx-vlm mlx-serve start --host 0.0.0.0 --port 8095
```

Check: `curl http://kahuna.local:8095/v1/models` (both `qwen3-vl-30b-a3b` and fallback `qwen2.5-vl-3b` listed).
Model loads on first request (~60 GB bf16 — allow a minute); unloads after 10 min idle.

Fallbacks: same server, `ALERTER_LLM_MODEL=qwen2.5-vl-3b`; or MLX-Serve.app on :11234
(`ddalcu/Qwen3.8-Flash-Next-MLX-Serve-iQ-MLX-4.7bpw` — general LLM with bolted-on vision, weaker at
small/night objects, shares inference with the coding session).

## 2. Run the watcher

```sh
# on kahuna (LLM local) — defaults already point here:
/opt/homebrew/bin/python3.15 alerter.py        # watch forever, 60 s polls

# on the dev Mac (LLM remote on kahuna):
export ALERTER_LLM_BASE=http://kahuna.local:8095/v1
python3 alerter.py
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

Per-day browse: date chips, thumbnail per clip (click to play the clip inline), status badge
(alert/ok/skip/pending/unanalyzed), verdict text, the camera's learned scene map, and
**false positive / real alert** buttons with an optional note — those become prompt examples
immediately (stored in `feedback.json`, which the watcher re-reads every cycle, so no restart needed).
CLI equivalent: `alerter.py --mark SUB --label fp --note …`.

## 5. Output

- macOS notification (sound) for HIGH threat, and for candidates the LLM could not analyze.
- `alerts.log` — every decision, one line per clip.
- `state.json` — per-clip status (`pending|skip|ok|alert|unanalyzed|error`) + per-camera thumbnail
  baselines. Delete or `--reset` to reprocess.

## Env vars

| Var | Default | Meaning |
|---|---|---|
| `ALERTER_ROOT` | `/Volumes/aaron/SecuritySpy_secondary` | capture root |
| `ALERTER_LLM_BASE` | `http://127.0.0.1:8095/v1` | OpenAI-compatible base (uvx mlx-serve on kahuna) |
| `ALERTER_LLM_MODEL` | `qwen3-vl-30b-a3b` | model id the server accepts |
| `ALERTER_LLM_TIMEOUT` | `180` | per-call seconds |
| `ALERTER_DIFF_THRESHOLD` | `12.0` | thumbnail-diff below this = repeat clip, skipped (dedupe only, see CLAUDE.md) |
| `ALERTER_REVIEW_PORT` / `ALERTER_REVIEW_HOST` | `8765` / `127.0.0.1` | review server bind |
| `ALERTER_STATE` / `ALERTER_LOG` / `ALERTER_SCENE` / `ALERTER_FEEDBACK` | files in project dir | override data paths |

## Autostart (later)

`launchd` plists for both the mlx-serve server and `alerter.py` once the flow is trusted; until then
start the two commands manually.
