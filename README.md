# agentic-assets-manageai

[![GitHub release](https://img.shields.io/github/v/release/manageai-inet/agentic-assets-manageai?label=version)](https://github.com/manageai-inet/agentic-assets-manageai/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/manageai-inet/agentic-assets-manageai.svg)](https://pkg.go.dev/github.com/manageai-inet/agentic-assets-manageai)

Extended Version of [agentic-assets](https://github.com/manageai-inet/agentic-assets) that wires the toolkit to ManageAI services (OCR, LLM, Embedding, Reranking).

## Installation

```bash
go get github.com/manageai-inet/agentic-assets-manageai
```

Requires Go `1.25` or newer.

## Components

| Component | File | Purpose |
| --- | --- | --- |
| `OcrLayoutApiRepo` / `OcrLayoutApiConverter` | `ocr.go` | Calls the ManageAI OCR layout API and converts results into `KnowledgeSource` entries (one per page). |
| `LLM` | `llm.go` | Adapter that bridges OpenAI-style `ChatCompletion` calls onto the ManageAI client. |
| `Embedder` | `embedding.go` | Wraps the ManageAI embedding API (single + batch). |
| `Reranker` | `reranking.go` | Wraps the ManageAI rerank API for query/reference scoring. |

## Quick Start

Every component supports two constructors: an explicit one and a `*FromEnv` variant that reads configuration from environment variables.

### OCR

```go
import (
    "context"
    assetmai "github.com/manageai-inet/agentic-assets-manageai"
)

repo := assetmai.NewOcrLayoutApiRepoFromEnv()
converter := assetmai.NewOcrLayoutApiConverter("pdf", repo)

sources, err := converter.Convert(ctx, kbId, sourceName, sourceUrl, fileBytes, nil)
```

Supported `fileType` values: `pdf`, `doc`, `docx`, `ppt`, `pptx`.

### LLM

```go
llm := assetmai.NewLLMFromEnv()
msg, err := llm.Generate(ctx, messages, tools, toolChoice)
```

### Embedding

```go
embedder := assetmai.NewEmbedderFromEnv()
vector, err := embedder.Embed(ctx, "hello world")
vectors, err := embedder.EmbedBatch(ctx, []string{"a", "b"})
```

### Reranking

```go
reranker := assetmai.NewRerankerFromEnv()
scores, err := reranker.Rerank(ctx, query, references)
```

## Environment Variables

| Variable | Used by | Description |
| --- | --- | --- |
| `OCR_API_PATH` | `NewOcrLayoutApiRepoFromEnv` | Endpoint of the OCR layout API. |
| `OCR_API_KEY` | `NewOcrLayoutApiRepoFromEnv` | Basic auth key for the OCR layout API. |
| `MAI_EMBEDDING_MODEL` | `NewEmbedderFromEnv` | Embedding model id (resolved via the ManageAI redirect helper). |
| `MAI_EMBEDDING_DIM` | `NewEmbedderFromEnv` | Vector dimension as an integer. |
| `MAI_RERANKING_MODEL` | `NewRerankerFromEnv` | Reranking model id (resolved via the ManageAI redirect helper). |

The LLM adapter and any `*FromEnv` constructor that talks to ManageAI also rely on the credentials picked up by `mai.GetDefaultClient()`.

## GitHub Workflows

Workflows live under [`.github/workflows/`](.github/workflows).

### `release.yml`

Triggers on tags matching `v*` and publishes a GitHub Release with auto-generated notes.

- **Trigger:** `push` of a tag matching `v*`
- **Permissions:** `contents: write`
- **Action:** [`softprops/action-gh-release@v2`](https://github.com/softprops/action-gh-release) with `generate_release_notes: true`

To cut a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The workflow will create the corresponding GitHub Release automatically.

## License

See [LICENSE](LICENSE).
