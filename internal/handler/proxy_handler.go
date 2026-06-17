package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"llm_relay/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var upstreamHTTPClient = &http.Client{
	Timeout: 120 * time.Second,
}

func ChatCompletions(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var requestBody map[string]interface{}
	if err := json.Unmarshal(body, &requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	modelName, ok := requestBody["model"].(string)
	if !ok || strings.TrimSpace(modelName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	config, err := repository.GetConfigByExternalModel(modelName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "model not configured"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	requestBody["model"] = config.TargetModel

	payload, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode upstream request"})
		return
	}

	upstreamURL, err := buildChatCompletionsURL(config.TargetBaseURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	upstreamReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, upstreamURL, bytes.NewReader(payload))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upstream request"})
		return
	}

	copyAllowedRequestHeaders(c.Request.Header, upstreamReq.Header)
	upstreamReq.Header.Set("Authorization", "Bearer "+config.TargetAPIKey)
	upstreamReq.Header.Set("Content-Type", "application/json")
	if upstreamReq.Header.Get("Accept") == "" {
		upstreamReq.Header.Set("Accept", "application/json")
	}

	upstreamResp, err := upstreamHTTPClient.Do(upstreamReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream request failed", "detail": err.Error()})
		return
	}
	defer upstreamResp.Body.Close()

	copyUpstreamResponseHeaders(c.Writer.Header(), upstreamResp.Header)
	c.Status(upstreamResp.StatusCode)

	stream, _ := requestBody["stream"].(bool)
	if stream {
		streamUpstreamResponse(c, upstreamResp.Body)
		return
	}

	if _, err := io.Copy(c.Writer, upstreamResp.Body); err != nil {
		_ = c.Error(err)
	}
}

var versionPathRegex = regexp.MustCompile(`/v\d+$`)

func buildChatCompletionsURL(base string) (string, error) {
	trimmed := strings.TrimSpace(base)
	if trimmed == "" {
		return "", errInvalidBaseURL("target_base_url is required")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", errInvalidBaseURL("invalid target_base_url")
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", errInvalidBaseURL("invalid target_base_url")
	}

	switch {
	case strings.HasSuffix(parsed.Path, "/chat/completions"):
		return parsed.String(), nil
	case versionPathRegex.MatchString(parsed.Path):
		parsed.Path = path.Join(parsed.Path, "chat/completions")
	default:
		parsed.Path = path.Join(parsed.Path, "/v1/chat/completions")
	}

	return parsed.String(), nil
}

func copyAllowedRequestHeaders(src http.Header, dst http.Header) {
	for header, values := range src {
		switch http.CanonicalHeaderKey(header) {
		case "Authorization", "Connection", "Content-Length", "Host":
			continue
		default:
			for _, value := range values {
				dst.Add(header, value)
			}
		}
	}
}

func copyUpstreamResponseHeaders(dst http.Header, src http.Header) {
	for header, values := range src {
		switch http.CanonicalHeaderKey(header) {
		case "Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade":
			continue
		default:
			for _, value := range values {
				dst.Add(header, value)
			}
		}
	}
}

func streamUpstreamResponse(c *gin.Context, body io.Reader) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming is not supported"})
		return
	}

	buffer := make([]byte, 32*1024)
	for {
		n, err := body.Read(buffer)
		if n > 0 {
			if _, writeErr := c.Writer.Write(buffer[:n]); writeErr != nil {
				_ = c.Error(writeErr)
				return
			}
			flusher.Flush()
		}

		if err == nil {
			continue
		}
		if err == io.EOF {
			return
		}

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			_ = c.Error(netErr)
			return
		}

		_ = c.Error(err)
		return
	}
}

type invalidBaseURLError string

func (e invalidBaseURLError) Error() string {
	return string(e)
}

func errInvalidBaseURL(message string) error {
	return invalidBaseURLError(message)
}
