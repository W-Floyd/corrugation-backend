package backend

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

type infinityEmbeddingsRequest struct {
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format"`
	Input          []string `json:"input"`
	Modality       string   `json:"modality"`
}

type infinityEmbeddingsReponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     uint      `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens uint `json:"prompt_tokens"`
		TotalTokens  uint `json:"total_tokens"`
	} `json:"usage"`
	ID      string `json:"id"`
	Created int64  `json:"created"`
}

type Embeddings []float64

var embeddingsCache map[string]Embeddings

func init() {
	embeddingsCache = make(map[string]Embeddings)
}

func (i *infinityEmbeddingsRequest) GenerateEmbeddings() (e Embeddings, err error) {

	b, err := json.Marshal(*i)
	if err != nil {
		return
	}

	c := http.Client{}
	resp, err := c.Post(infinityAddress+"/embeddings", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.Join(errors.New(string(respBody)), errors.New("http error "+strconv.Itoa(resp.StatusCode)+" when submitting data to Infinity backend"))
		return
	}

	var infinityResponse infinityEmbeddingsReponse

	err = json.Unmarshal(respBody, &infinityResponse)
	if err != nil {
		return
	}

	e = infinityResponse.Data[0].Embedding

	return

}

func GenerateTextDocumentEmbeddingsCtx(ctx context.Context, input string) (Embeddings, error) {
	uc, _ := loadUserConfig(UsernameFromContext(ctx))
	textModel, _, _, docPrefix := effectiveInfinityConfig(uc)
	return generateTextEmbeddings(docPrefix+input, textModel)
}

func GenerateTextQueryEmbeddingsCtx(ctx context.Context, input string) (Embeddings, error) {
	uc, _ := loadUserConfig(UsernameFromContext(ctx))
	textModel, _, queryPrefix, _ := effectiveInfinityConfig(uc)
	return generateTextEmbeddings(queryPrefix+input, textModel)
}

func GenerateImageQueryEmbeddingsCtx(ctx context.Context, input string) (Embeddings, error) {
	uc, _ := loadUserConfig(UsernameFromContext(ctx))
	_, imageModel, _, _ := effectiveInfinityConfig(uc)
	return generateTextEmbeddings(input, imageModel)
}

func generateTextEmbeddings(input string, model string) (e Embeddings, err error) {
	infinityRequest := infinityEmbeddingsRequest{
		Model:          model,
		EncodingFormat: "float",
		Input:          []string{input},
		Modality:       "text",
	}
	e, err = infinityRequest.GenerateEmbeddings()
	return
}

func (i *Image) GenerateEmbeddings(ctx context.Context) (err error) {
	if i.ID == 0 {
		err = errors.New("artifact must be persisted before generating embeddings")
		return
	}
	if i.Data == nil || len(*i.Data) == 0 {
		err = errors.New("no data in image")
		return
	}

	uc, _ := loadUserConfig(UsernameFromContext(ctx))
	_, imageModel, _, _ := effectiveInfinityConfig(uc)

	base64Image := base64.StdEncoding.EncodeToString(*i.Data)
	base64Image = "data:" + http.DetectContentType(*i.Data) + ";base64," + base64Image

	infinityRequest := infinityEmbeddingsRequest{
		Model:          imageModel,
		EncodingFormat: "float",
		Input:          []string{base64Image},
		Modality:       "image",
	}

	e, err := infinityRequest.GenerateEmbeddings()
	if err != nil {
		return
	}

	id := i.ID
	err = saveEmbedding(nil, &id, e, imageModel)
	return
}

func AverageEmbeddings(vecs []Embeddings) (Embeddings, error) {
	if len(vecs) == 0 {
		return nil, errors.New("no embeddings to average")
	}
	dim := len(vecs[0])
	for _, v := range vecs {
		if len(v) != dim {
			return nil, errors.New("embedding dimension mismatch")
		}
	}
	avg := make(Embeddings, dim)
	n := float64(len(vecs))
	for _, v := range vecs {
		for i, x := range v {
			avg[i] += x / n
		}
	}
	return avg, nil
}

func (e *Embeddings) MarshalEmbeddings() (hash string, jsonData []byte, err error) {

	jsonData, err = json.Marshal(*e)
	if err != nil {
		return
	}

	h := sha256.New()
	_, err = h.Write(jsonData)
	if err != nil {
		return
	}

	hash = string(h.Sum(nil))

	if _, ok := embeddingsCache[hash]; !ok {
		embeddingsCache[hash] = *e
	}

	return
}
