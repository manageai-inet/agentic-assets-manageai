package assetmai

import (
	"context"
	"os"

	mai "github.com/manageai-inet/Eino-ManageAI-Extension/components/core/manageai"
)

type Reranker struct {
	client *mai.Client
	model string
}

func NewReranker(client *mai.Client, model string) *Reranker {
	return &Reranker{client: client, model: model}
}

func NewRerankerFromEnv() *Reranker {
	client, err := mai.GetDefaultClient()
	if err != nil {
		panic(err.Error())
	}
	model := os.Getenv("MAI_RERANKING_MODEL")
	if model == "" {
		panic("`MAI_RERANKING_MODEL` environment variable is not set")
	}
	redirectHelper := mai.GetRedirectHelper()
	redirectModel, err := redirectHelper.GetModelId(&model, false)
	if err != nil {
		panic("`MAI_RERANKING_MODEL` environment variable is not set as valid model: " + err.Error())
	}
	model = *redirectModel
	return &Reranker{client: client, model: model}
}

func (r *Reranker) GetRerankingModel() string {
	return r.model
}

func (r *Reranker) Rerank(ctx context.Context, query string, references []string) ([]float32, error) {
	rerankOpts := &mai.RerankOptions{Model: &r.model}
	response, err := r.client.Rerank(query, references, rerankOpts)
	if err != nil {
		return nil, err
	}
	scores := make([]float32, len(response.Data))
	for i, score := range response.Data {
		s := score.RelevanceScore
		scores[i] = float32(s)
	}
	return scores, nil
}