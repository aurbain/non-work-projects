#!/usr/bin/env python3
import argparse
import subprocess
import tempfile
import os
import glob
import sys
import math

def get_video_info(video_path):
    """Get the duration of the video using ffprobe."""
    cmd = [
        "ffprobe", "-v", "error", "-show_entries", "format=duration",
        "-of", "default=noprint_wrappers=1:nokey=1", video_path
    ]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        return float(result.stdout.strip())
    except Exception:
        return None

def main():
    parser = argparse.ArgumentParser(description="Analyze a whole video by extracting sequential frames.")
    parser.add_argument("video", help="Path to the video file (.m4v, .mp4, etc.)")
    parser.add_argument("--frames", type=int, default=10, help="Number of frames to extract (to avoid memory crashes, default: 10)")
    parser.add_argument("--prompt", type=str, 
                        default="What is happening in this part of the video? Note any specific actions or changes.", 
                        help="Per-frame analysis prompt")
    args = parser.parse_args()

    if not os.path.exists(args.video):
        print(f"Error: Video file '{args.video}' not found.")
        sys.exit(1)

    try:
        import ollama
    except ImportError:
        print("Error: 'ollama' Python library not found. Install it with: pip install ollama")
        sys.exit(1)

    # Calculate interval
    duration = get_video_info(args.video)
    if duration:
        interval = max(1, math.floor(duration / args.frames)) if duration > args.frames else 1
        print(f"Video Duration: {duration:.2f}s | Analyzing {args.frames} points across the timeline...")
    else:
        interval = 5
        print(f"Could not detect duration. Defaulting to 1 frame every {interval}s")

    with tempfile.TemporaryDirectory() as temp_dir:
        print(f"Step 1: Extracting frames with ffmpeg...")
        output_pattern = os.path.join(temp_dir, "frame-%04d.jpg")
        
        ffmpeg_cmd = [
            "ffmpeg", "-hide_banner", "-loglevel", "error", "-y",
            "-i", args.video,
            "-vf", f"fps=1/{interval}",
            output_pattern
        ]
        
        try:
            subprocess.run(ffmpeg_cmd, check=True)
        except Exception as e:
            print(f"Error running ffmpeg: {e}")
            sys.exit(1)

        frames = sorted(glob.glob(os.path.join(temp_dir, "frame-*.jpg")))[:args.frames]
        if not frames:
            print("Error: No frames extracted.")
            sys.exit(1)

        print(f"Step 2: Processing {len(frames)} frames sequentially (Timeline mode)...")
        
        chat_history = []
        full_timeline = []

        for i, frame_path in enumerate(frames):
            timestamp = i * interval
            print(f"  [~{timestamp}s] Analyzing frame {i+1}/{len(frames)}...")
            
            try:
                msg = {
                    'role': 'user',
                    'content': f"Frame at roughly {timestamp} seconds: {args.prompt}",
                    'images': [frame_path]
                }
                
                # To keep it stable, we don't send the entire history of images back every time.
                # We just ask the model to describe this frame in the context of the convo.
                response = ollama.chat(
                    model='qwen2.5vl:7b',
                    messages=chat_history + [msg]
                )
                
                result = response['message']['content']
                full_timeline.append(f"At {timestamp}s: {result}")
                
                # Add a text-only summary of the frame to history to keep context without bloating memory
                chat_history.append(msg)
                chat_history.append({'role': 'assistant', 'content': result})
                
            except Exception as e:
                print(f"    Error on frame {i}: {e}")
                continue

        print("\n" + "="*50)
        print("VIDEO ANALYSIS SUMMARY")
        print("="*50)
        
        # Final pass: ask for a summary of the whole thing based on the textual timeline
        try:
            summary_request = {
                'role': 'user',
                'content': "Based on all the frames you just analyzed, give me a concise summary of the whole video's events."
            }
            final_response = ollama.chat(model='qwen2.5vl:7b', messages=chat_history + [summary_request])
            print(final_response['message']['content'])
        except Exception as e:
            print("Could not generate final summary. Showing raw timeline instead.")
            for entry in full_timeline:
                print(entry)
        
        print("="*50 + "\n")

if __name__ == "__main__":
    main()


if __name__ == "__main__":
    main()
