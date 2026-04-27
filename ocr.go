package assetmai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strconv"

	am "github.com/manageai-inet/agentic-assets"
)

type OcrLayoutApiRequestConfig struct {
	ProjectId       string `json:"project_id,omitzero"`
	Location        string `json:"location,omitzero"`
	ProcessorId     string `json:"processor_id,omitzero"`
	CredentialsPath string `json:"credentials_path,omitzero"`
}

type OcrLayoutApiResponseDataPage struct {
	Page string `json:"page"` // start from 1
	Text string `json:"text"`
}

type OcrLayoutApiResponseData struct {
	DocumentName string 					`json:"document_name"`
	TotalPage int 							`json:"total_page"`
	Pages []OcrLayoutApiResponseDataPage 	`json:"pages"`
}

type OcrLayoutApiResponse struct {
	Status  bool                   `json:"status"`
	Message string                 `json:"message"`
	Data    *OcrLayoutApiResponseData `json:"data"`
}

type OcrLayoutApiRepo struct {
	apiPath        string
	apiKey         string
	defaultOptions OcrLayoutApiRequestConfig
	am.LoggingCapacity
}

func NewOcrLayoutApiRepo(apiPath string, apiKey string, config *OcrLayoutApiRequestConfig) *OcrLayoutApiRepo {
	cfg := &OcrLayoutApiRequestConfig{}
	if config != nil {
		cfg = config
	}
	return &OcrLayoutApiRepo{
		apiPath:        apiPath,
		apiKey:         apiKey,
		defaultOptions: *cfg,
	}
}

func NewOcrLayoutApiRepoFromEnv() *OcrLayoutApiRepo {
	apiPath := os.Getenv("OCR_API_PATH")
	if apiPath == "" {
		panic("Please set `OCR_API_PATH` environment variable")
	}
	apiKey := os.Getenv("OCR_API_KEY")
	if apiKey == "" {
		panic("Please set `OCR_API_KEY` environment variable")
	}
	return &OcrLayoutApiRepo{
		apiPath:        apiPath,
		apiKey:         apiKey,
	}
}

func (r *OcrLayoutApiRepo) Process(ctx context.Context, filename string, fileData []byte, fileType string) (*OcrLayoutApiResponse, error) {
	cfg := r.defaultOptions
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("project_id", cfg.ProjectId)
	writer.WriteField("location", cfg.Location)
	writer.WriteField("processor_id", cfg.ProcessorId)
	writer.WriteField("credentials_path", cfg.CredentialsPath)
	writer.WriteField("file-type", fileType)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, bytes.NewReader(fileData))
	if err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.apiPath, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Basic "+r.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	result := &OcrLayoutApiResponse{}
	json.NewDecoder(resp.Body).Decode(result)
	return result, nil
}

type OcrLayoutApiConverter struct {
	fileType string // support pdf, doc, docx, ppt, pptx
	ocrRepo *OcrLayoutApiRepo
	am.LoggingCapacity
}

func NewOcrLayoutApiConverter(fileType string, ocrRepo *OcrLayoutApiRepo) *OcrLayoutApiConverter {
	return &OcrLayoutApiConverter{
		fileType: fileType,
		ocrRepo: ocrRepo,
		LoggingCapacity: *am.GetDefaultLoggingCapacity(),
	}
}

func (c *OcrLayoutApiConverter) String() string {
	return "OcrLayoutApiConverter"
}

func (c *OcrLayoutApiConverter) Convert(ctx context.Context, kbId, sourceName, sourceUrl string, sourceData []byte, metadata *map[string]any) ([]am.KnowledgeSource, error) {
	logger := c.GetLogger()
	logger.InfoContext(ctx, "Processing OCR to Convert Knowledge", slog.String("kbId", kbId), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl), slog.String("fileType", c.fileType))
	result, err := c.ocrRepo.Process(ctx, sourceName, sourceData, c.fileType)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to process OCR: "+err.Error(), slog.String("kbId", kbId), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl), slog.String("fileType", c.fileType))
		return nil, err
	}
	if result == nil || !result.Status {
		logger.ErrorContext(ctx, "Failed to process OCR: "+result.Message, slog.String("kbId", kbId), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl), slog.String("fileType", c.fileType))
		return nil, fmt.Errorf("failed to process ocr: %s", result.Message)
	}
	logger.InfoContext(ctx, "Successfully processed OCR", slog.String("kbId", kbId), slog.Int("totalPage", result.Data.TotalPage), slog.String("fileType", c.fileType))
	ocrPages := result.Data.Pages
	sort.Slice(ocrPages, func(i, j int) bool {
		pageI, _ := strconv.Atoi(ocrPages[i].Page)
		pageJ, _ := strconv.Atoi(ocrPages[j].Page)
		return pageI < pageJ
	})
	var sources []am.KnowledgeSource
	for i, page := range ocrPages {
		// Create a new map for this specific page
		pageMetadata := make(map[string]any)
		// Copy existing metadata into the new map
		if metadata != nil {
			for k, v := range *metadata {
				pageMetadata[k] = v
			}
		}
		pageSource := am.KnowledgeSource{
			SourceType: am.AssetTypePage,
			SourceName: kbId + ":" + sourceName + ":" + strconv.Itoa(i+1),
			SourceUrl:  &sourceUrl,
			SourceData: &sourceData,
			SourceContents: &page.Text,
			Metadata: &pageMetadata,
		}
		sources = append(sources, pageSource)
	}
	logger.InfoContext(ctx, "Successfully Convert Knowledge", slog.String("kbId", kbId), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl), slog.Int("totalPage", result.Data.TotalPage), slog.String("fileType", c.fileType))
	return sources, nil
}