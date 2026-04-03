# recap

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A CLI agent that analyses meeting transcripts — extracts decisions, ownership, deadlines, and contradictions across multiple documents.

Built as an exploration of [Recursive Language Models](https://arxiv.org/abs/2512.24601) (Zhang, Kraska & Khattab, 2025). The core idea from the paper is that long prompts should not be fed into the LLM directly but instead treated as an external environment the model can programmatically interact with. recap applies this by loading documents into a JavaScript sandbox and letting the agent write code to examine, slice, and recursively query sub-LLMs over relevant sections.

## How it works

1. Documents are loaded into a persistent [goja](https://github.com/dop251/goja) JavaScript sandbox as a `documents` array.
2. The agent never sees the raw document text in its context window. Instead, it writes JavaScript to search, filter, and extract relevant sections.
3. An `llm_query(context, query)` function lets the agent spawn sub-LLM calls to reason over specific text chunks — the recursive part.
4. When the context window approaches its budget (90% of token limit), the conversation is summarized and the agent resumes on a fresh context while the sandbox state persists.
5. Results are returned as structured data: decisions (with status), owners, deadlines, and cross-document contradictions — all with source citations.

## Install

```bash
go install github.com/kujtimiihoxha/recap@latest
```

Or build from source:

```bash
git clone https://github.com/kujtimiihoxha/recap.git
cd recap
go build -o recap .
```

## Usage

```
export OPENAI_API_KEY=<your-key>
```

### Analyse documents

```bash
# Analyse all documents in a directory
recap analyse ./examples/related/city-council

# Override model (default: gpt-5.4)
recap analyse --model gpt-4.1 ./examples/standalone/single-meeting

# Disable context summarization
recap analyse --context-limit 0 ./path/to/docs
```

### Ask questions

```bash
recap ask ./examples/related/city-council "What was decided about the park renovation budget?"
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--model` | `gpt-5.4` (or `MODEL` env) | LLM model to use |
| `--context-limit` | `250000` | Token budget; summarization at 90%. `0` disables summarization |

## Architecture

```
internal/
  agent/      # Agent loop, prompts, tool definitions
  sandbox/    # Goja JS runtime with llm_query bridge
  render/     # TUI, streaming output, analysis rendering
```

Key design choices:

- **Code-as-reasoning**: The agent writes JavaScript to interact with documents rather than having them injected into the prompt. This lets it handle documents far larger than the context window.
- **Persistent sandbox**: Variables and functions survive across multiple tool calls and even across context summarization boundaries.
- **Sub-LLM calls**: `llm_query` spawns a separate LLM invocation focused on a specific chunk of text, keeping the outer agent's context clean.
- **Structured output**: The `submit_analysis` tool enforces a typed schema for results, so downstream consumers get reliable JSON.

## Examples

The `examples/` directory contains sample meeting transcripts sourced from the [Meeting Transcripts](https://www.kaggle.com/datasets/abhishekunnam/meeting-transcripts) dataset on Kaggle:

- `examples/standalone/` — single document analysis
- `examples/related/` — multi-document analysis (e.g. a series of city council meetings)

## Inspired by

- [Recursive Language Models](https://arxiv.org/abs/2512.24601) — the paper that introduced the idea of treating long inputs as an environment the LLM interacts with programmatically rather than consuming directly
