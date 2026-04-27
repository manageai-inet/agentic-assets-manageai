package assetmai

import (
	"context"
	"os"
	"strconv"

	mai "github.com/manageai-inet/Eino-ManageAI-Extension/components/core/manageai"
)

type Embedder struct {
	client *mai.Client
	model string
	dim int
}

func NewEmbedder(client *mai.Client, model string, dim int) *Embedder {
	return &Embedder{client: client, model: model, dim: dim}
}

func NewEmbedderFromEnv() *Embedder {
	client, err := mai.GetDefaultClient()
	if err != nil {
		panic(err.Error())
	}
	model := os.Getenv("MAI_EMBEDDING_MODEL")
	if model == "" {
		panic("`MAI_EMBEDDING_MODEL` environment variable is not set")
	}
	redirectHelper := mai.GetRedirectHelper()
	redirectModel, err := redirectHelper.GetModelId(&model, false)
	if err != nil {
		panic("`MAI_EMBEDDING_MODEL` environment variable is not set as valid model: " + err.Error())
	}
	model = *redirectModel

	dimStr := os.Getenv("MAI_EMBEDDING_DIM")
	if dimStr == "" {
		panic("`MAI_EMBEDDING_DIM` environment variable is not set")
	}
	dim, err := strconv.Atoi(dimStr)
	if err != nil {
		panic("`MAI_EMBEDDING_DIM` environment variable is not set as valid integer: " + err.Error())
	}
	return &Embedder{client: client, model: model, dim: dim}
}

func (m *Embedder) GetEmbeddingModel() string {
	return m.model
}

func (m *Embedder) GetEmbeddingDim() int {
	return m.dim
}

func (m *Embedder) Embed(ctx context.Context, content string) ([]float32, error) {
	response, err := m.client.Embedding([]string{content}, &mai.EmbeddingOptions{
		Model: &m.model,
	})
	if err != nil {
		return nil, err
	}
	vector := []float32{}
	for _, v := range response.Data[0].Embedding {
		vector = append(vector, float32(v))
	}
	return vector, nil
}

func (m *Embedder) EmbedBatch(ctx context.Context, contents []string) ([][]float32, error) {
	response, err := m.client.Embedding(contents, &mai.EmbeddingOptions{
		Model: &m.model,
	})
	if err != nil {
		return nil, err
	}
	embeddings := [][]float32{}
	for _, embedding := range response.Data {
		vector := []float32{}
		for _, v := range embedding.Embedding {
			vector = append(vector, float32(v))
		}
		embeddings = append(embeddings, vector)
	}
	return embeddings, nil
}