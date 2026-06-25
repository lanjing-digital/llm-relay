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
	"unicode/utf8"

	"llm_relay/internal/model"
	"llm_relay/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var upstreamHTTPClient = &http.Client{
	Timeout: 120 * time.Second,
}

func ChatCompletions(c *gin.Context) {
	start := time.Now()

	logEntry := model.Log{
		Method:   c.Request.Method,
		Path:     c.Request.URL.Path,
		ClientIP: c.ClientIP(),
	}

	var requestSnippet string
	var responseSnippet string
	var logErr string
	var externalModel string
	var targetModel string
	var upstreamURL string

	defer func() {
		logEntry.ExternalModel = externalModel
		logEntry.TargetModel = targetModel
		logEntry.UpstreamURL = upstreamURL
		logEntry.RequestSnippet = requestSnippet
		logEntry.ResponseSnippet = responseSnippet
		if logErr == "" && len(c.Errors) > 0 {
			logErr = c.Errors.String()
		}
		logEntry.Error = logErr
		logEntry.StatusCode = c.Writer.Status()
		logEntry.DurationMs = time.Since(start).Milliseconds()
		_ = repository.CreateLog(&logEntry)
	}()

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logErr = "failed to read body"
		responseSnippet = writeJSONWithSnippet(c, http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}
	requestSnippet = truncateUTF8Bytes(body, 2048)

	var requestBody map[string]interface{}
	if err := json.Unmarshal(body, &requestBody); err != nil {
		logErr = "invalid JSON"
		responseSnippet = writeJSONWithSnippet(c, http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	modelName, ok := requestBody["model"].(string)
	if !ok || strings.TrimSpace(modelName) == "" {
		logErr = "model is required"
		responseSnippet = writeJSONWithSnippet(c, http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}
	externalModel = modelName

	config, err := repository.GetConfigByExternalModel(modelName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logErr = "model not configured"
			responseSnippet = writeJSONWithSnippet(c, http.StatusBadRequest, gin.H{"error": "model not configured"})
			return
		}
		logErr = err.Error()
		responseSnippet = writeJSONWithSnippet(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	targetModel = config.TargetModel

	upstreamURL, err = buildChatCompletionsURL(config.TargetBaseURL)
	if err != nil {
		logErr = err.Error()
		responseSnippet = writeJSONWithSnippet(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !validateBearerToken(c.GetHeader("Authorization"), config.TargetAPIKey) {
		logErr = "unauthorized"
		responseSnippet = writeJSONWithSnippet(c, http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	requestBody["model"] = config.TargetModel

	payload, err := json.Marshal(requestBody)
	if err != nil {
		logErr = "failed to encode upstream request"
		responseSnippet = writeJSONWithSnippet(c, http.StatusInternalServerError, gin.H{"error": "failed to encode upstream request"})
		return
	}

	upstreamReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, upstreamURL, bytes.NewReader(payload))
	if err != nil {
		logErr = "failed to create upstream request"
		responseSnippet = writeJSONWithSnippet(c, http.StatusInternalServerError, gin.H{"error": "failed to create upstream request"})
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
		logErr = "upstream request failed"
		responseSnippet = writeJSONWithSnippet(c, http.StatusBadGateway, gin.H{"error": "upstream request failed", "detail": err.Error()})
		return
	}
	defer upstreamResp.Body.Close()

	copyUpstreamResponseHeaders(c.Writer.Header(), upstreamResp.Header)
	c.Status(upstreamResp.StatusCode)

	stream, _ := requestBody["stream"].(bool)
	if stream {
		collector := newSSEContentCollector()
		streamUpstreamResponse(c, io.TeeReader(upstreamResp.Body, collector))
		responseSnippet = collector.result()
		if upstreamResp.StatusCode >= 400 {
			logErr = upstreamResp.Status
		}
		return
	}

	snippetWriter := &limitedBufferWriter{Limit: 2048}
	reader := io.TeeReader(upstreamResp.Body, snippetWriter)
	if _, err := io.Copy(c.Writer, reader); err != nil {
		_ = c.Error(err)
	}
	responseSnippet = snippetWriter.String()
	if upstreamResp.StatusCode >= 400 {
		logErr = upstreamResp.Status
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

func validateBearerToken(authHeader string, expectedToken string) bool {
	parts := strings.Fields(authHeader)
	if len(parts) != 2 {
		return false
	}
	if parts[0] != "Bearer" {
		return false
	}
	return parts[1] == expectedToken
}

type limitedBufferWriter struct {
	Buffer bytes.Buffer
	Limit  int
}

func (w *limitedBufferWriter) Write(p []byte) (int, error) {
	remaining := w.Limit - w.Buffer.Len()
	if remaining > 0 {
		if len(p) > remaining {
			_, _ = w.Buffer.Write(p[:remaining])
		} else {
			_, _ = w.Buffer.Write(p)
		}
	}
	return len(p), nil
}

func (w *limitedBufferWriter) String() string {
	return w.Buffer.String()
}

func truncateUTF8Bytes(b []byte, limit int) string {
	if limit <= 0 {
		return ""
	}
	if len(b) <= limit {
		return string(b)
	}

	truncated := b[:limit]
	for len(truncated) > 0 && !utf8.Valid(truncated) {
		truncated = truncated[:len(truncated)-1]
	}
	return string(truncated)
}

func writeJSONWithSnippet(c *gin.Context, statusCode int, body any) string {
	encoded, err := json.Marshal(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode response"})
		return ""
	}
	c.Data(statusCode, "application/json; charset=utf-8", encoded)
	return truncateUTF8Bytes(encoded, 2048)
}

type sseContentCollector struct {
	buf              bytes.Buffer
	reasoningContent strings.Builder
	content          strings.Builder
}

func newSSEContentCollector() *sseContentCollector {
	return &sseContentCollector{}
}

func (s *sseContentCollector) Write(p []byte) (int, error) {
	s.buf.Write(p)
	s.parseLines()
	return len(p), nil
}

func (s *sseContentCollector) parseLines() {
	for {
		line, err := s.buf.ReadString('\n')
		if err != nil {
			s.buf.WriteString(line)
			return
		}
		s.processLine(line)
	}
}

func (s *sseContentCollector) processLine(line string) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || !strings.HasPrefix(trimmed, "data:") {
		return
	}
	data := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
	if data == "[DONE]" {
		return
	}
	var chunk map[string]interface{}
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return
	}
	choices, ok := chunk["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return
	}
	delta, ok := choices[0].(map[string]interface{})
	if !ok {
		return
	}
	if v, ok := delta["reasoning_content"].(string); ok {
		s.reasoningContent.WriteString(v)
	}
	if v, ok := delta["content"].(string); ok {
		s.content.WriteString(v)
	}
}

func (s *sseContentCollector) result() string {
	var sb strings.Builder
	if s.reasoningContent.Len() > 0 {
		sb.WriteString("=== reasoning_content ===\n")
		sb.WriteString(s.reasoningContent.String())
		sb.WriteString("\n\n")
	}
	if s.content.Len() > 0 {
		sb.WriteString("=== content ===\n")
		sb.WriteString(s.content.String())
	}
	if sb.Len() == 0 {
		return "[stream] (no content captured)"
	}
	return sb.String()
}
