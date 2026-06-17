package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"llm_relay/internal/model"
	"llm_relay/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListConfigs(c *gin.Context) {
	configs, err := repository.GetAllConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, configs)
}

func CreateConfig(c *gin.Context) {
	var config model.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := repository.CreateConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, config)
}

func UpdateConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	existing, err := repository.GetConfigByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var input model.Config
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing.Name = input.Name
	existing.ExternalModel = input.ExternalModel
	existing.TargetModel = input.TargetModel
	existing.TargetBaseURL = input.TargetBaseURL
	existing.TargetAPIKey = input.TargetAPIKey

	if err := repository.UpdateConfig(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}

func DeleteConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if _, err := repository.GetConfigByID(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := repository.DeleteConfig(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func TestConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	config, err := repository.GetConfigByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	upstreamURL, err := buildChatCompletionsURL(config.TargetBaseURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	testPayload := map[string]interface{}{
		"model": config.TargetModel,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "ping",
			},
		},
		"stream":     false,
		"max_tokens": 1,
	}

	body, err := json.Marshal(testPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode test request"})
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, upstreamURL, bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upstream request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+config.TargetAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := upstreamHTTPClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error":   "upstream request failed",
			"detail":  err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"success":     false,
			"status_code": resp.StatusCode,
			"error":       "failed to read upstream response",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          resp.StatusCode >= 200 && resp.StatusCode < 300,
		"status_code":      resp.StatusCode,
		"status":           resp.Status,
		"upstream_url":     upstreamURL,
		"response_snippet": string(respBody),
	})
}
