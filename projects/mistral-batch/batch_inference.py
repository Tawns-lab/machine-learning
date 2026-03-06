import os
import time
import json
import argparse
from mistralai import Mistral

# ========================= CONFIG =========================
api_key = os.environ.get("MISTRAL_API_KEY")
if not api_key:
    raise ValueError("Please set MISTRAL_API_KEY environment variable")

client = Mistral(api_key=api_key)

MODEL = "mistral-small-latest"
ENDPOINT = "/v1/chat/completions"
METADATA = {"job_type": "unified_demo", "user": "alttawan"}

# =========================================================


def main():
    parser = argparse.ArgumentParser(
        description="Mistral Batch Inference – Unified Inline or File mode",
        formatter_class=argparse.ArgumentDefaultsHelpFormatter,
    )
    parser.add_argument(
        "--inline",
        action="store_true",
        help="Use inline batching (no file upload). Only for < 10,000 requests.",
    )
    parser.add_argument(
        "-n",
        "--num-requests",
        type=int,
        default=25,
        help="Number of requests to generate (1 - 1_000_000)",
    )
    args = parser.parse_args()

    use_inline = args.inline
    num_requests = args.num_requests

    # Safety check
    if num_requests < 1 or num_requests > 1_000_000:
        parser.error("--num-requests must be between 1 and 1_000_000")
    if use_inline and num_requests > 9999:
        print(
            "⚠️  Warning: Inline batching officially supports < 10,000 requests. "
            "Proceeding anyway (may fail on server side)."
        )

    mode = "INLINE" if use_inline else "FILE"
    print(f"🚀 Running in {mode} mode | {num_requests:,} requests | Model: {MODEL}")

    # === 1. Build the request list (same for both modes) ===
    requests_list = []
    for i in range(num_requests):
        req = {
            "custom_id": f"req_{i:06d}",
            "body": {
                "max_tokens": 200,
                "temperature": 0.7,
                "messages": [
                    {
                        "role": "user",
                        "content": f"Tell me an interesting fact about French culture – request #{i}",
                    }
                ],
            },
        }
        requests_list.append(req)

    # === 2. Create job – branch on mode ===
    if use_inline:
        # Inline batching (no file upload)
        job = client.batch.jobs.create(
            requests=requests_list,
            model=MODEL,
            endpoint=ENDPOINT,
            metadata=METADATA,
        )
        print(f"✅ Inline job created! Job ID: {job.id}")

    else:
        # File batching (recommended for large workloads)
        input_file_path = "batch_input.jsonl"
        with open(input_file_path, "w", encoding="utf-8") as f:
            for req in requests_list:
                f.write(json.dumps(req) + "\n")

        print(f"✅ Wrote {num_requests:,} requests to {input_file_path}")

        with open(input_file_path, "rb") as f:
            uploaded = client.files.upload(
                file={"file_name": input_file_path, "content": f},
                purpose="batch",
            )
        print(f"✅ Uploaded file ID: {uploaded.id}")

        job = client.batch.jobs.create(
            input_files=[uploaded.id],
            model=MODEL,
            endpoint=ENDPOINT,
            metadata=METADATA,
        )
        print(f"✅ File batch job created! Job ID: {job.id}")

    # === 3. Poll until finished ===
    print("⏳ Waiting for batch to complete...")
    while job.status not in ["SUCCESS", "FAILED", "TIMEOUT_EXCEEDED", "CANCELLED"]:
        time.sleep(8)
        job = client.batch.jobs.get(job_id=job.id)
        print(
            f"   Status: {job.status:20} | "
            f"Progress: {job.succeeded_requests or 0}/{job.total_requests} "
            f"({job.failed_requests or 0} failed)"
        )

    print(f"🏁 Final status: {job.status}")

    # === 4. Download results ===
    if job.status == "SUCCESS" and job.output_file:
        result_stream = client.files.download(file_id=job.output_file)
        output_path = f"batch_results_{'inline' if use_inline else 'file'}.jsonl"

        with open(output_path, "wb") as f:
            f.write(result_stream.read())

        print(f"✅ Results saved to {output_path}")

        # Pretty-print first 3 results
        print("\n--- First 3 results preview ---")
        with open(output_path, "r", encoding="utf-8") as f:
            for i, line in enumerate(f):
                if i >= 3:
                    break
                res = json.loads(line)
                content = (
                    res.get("response", {})
                    .get("choices", [{}])[0]
                    .get("message", {})
                    .get("content", "")
                )
                print(f"[{res['custom_id']}] → {content[:140]}...")
    else:
        print("❌ Job did not succeed or no output file available.")


if __name__ == "__main__":
    main()
