# mistral-batch

Batch inference against the [Mistral AI](https://mistral.ai/) API. Supports both **inline** (< 10k requests) and **file-based** (up to 1M requests) modes.

## Requirements

- Python 3.9+
- A [Mistral API key](https://console.mistral.ai/)

## Setup

```bash
cd projects/mistral-batch
pip install mistralai
export MISTRAL_API_KEY="your-key-here"
```

## Usage

### File mode (default) — recommended for large workloads

```bash
python batch_inference.py -n 1000
```

This writes a `batch_input.jsonl`, uploads it, submits a batch job, polls for completion, and saves results to `batch_results_file.jsonl`.

### Inline mode — convenient for small batches (< 10k requests)

```bash
python batch_inference.py --inline -n 50
```

No file upload step; the request list is sent directly in the job-creation call. Results land in `batch_results_inline.jsonl`.

## CLI reference

| Flag | Default | Description |
|---|---|---|
| `--inline` | off | Use inline batching instead of file upload |
| `-n / --num-requests` | 25 | Number of requests to generate (1 – 1,000,000) |

## Output format

Each line of the result `.jsonl` is a Mistral batch response object:

```json
{
  "custom_id": "req_000042",
  "response": {
    "choices": [{"message": {"role": "assistant", "content": "..."}}]
  }
}
```

## Architecture: Linguistic Logic Gate

The default prompt set demonstrates a **V_t / V_i (Transitive / Intransitive)** gate concept:

| Agent | Verb constraint | Role |
|---|---|---|
| `Agent_Vt` | Action → Object (external energy transfer) | Engine — drives narrative forward |
| `Agent_Vi` | Subject ↔ Action (no direct object) | Exhaust — describes resulting state |

A **Volta** at line 9 flips the weighting from transitive to intransitive, producing a "Thermodynamic Sonnet" structure. Swap the prompt list in `requests_list` with your own V_t / V_i constrained prompts and a Grammar-Schema validation step to enforce the constraint programmatically.
