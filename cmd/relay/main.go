package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	relayassets "llm_relay"
	"llm_relay/internal/db"
	"llm_relay/internal/handler"
	"llm_relay/internal/middleware"

	"github.com/gin-gonic/gin"
)

func serveAdminIndex(c *gin.Context, webDistFS fs.FS) {
	content, err := fs.ReadFile(webDistFS, "index.html")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load admin index"})
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
}

func main() {
	dbPath := "data.db"
	if err := db.Init(dbPath); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	r := gin.Default()

	registerAPIRoutes(r)
	registerAdminRoutes(r)

	port := os.Getenv("LLM_RELAY_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func registerAPIRoutes(r *gin.Engine) {
	api := r.Group("/api")

	api.POST("/auth/login", handler.Login)

	protected := api.Group("")
	protected.Use(middleware.JWTAuth())
	{
		protected.GET("/configs", handler.ListConfigs)
		protected.POST("/configs", handler.CreateConfig)
		protected.PUT("/configs/:id", handler.UpdateConfig)
		protected.DELETE("/configs/:id", handler.DeleteConfig)
		protected.POST("/configs/:id/test", handler.TestConfig)
		protected.GET("/logs", handler.ListLogs)
		protected.GET("/logs/:id", handler.GetLog)
	}

	r.POST("/v1/chat/completions", handler.ChatCompletions)
}

func registerAdminRoutes(r *gin.Engine) {
	webDistFS, err := fs.Sub(relayassets.EmbeddedWebDist, "web/dist")
	if err != nil {
		log.Fatalf("Failed to prepare embedded web assets: %v", err)
	}

	r.GET("/admin", func(c *gin.Context) {
		serveAdminIndex(c, webDistFS)
	})

	r.GET("/admin/*path", func(c *gin.Context) {
		requestPath := strings.TrimPrefix(c.Param("path"), "/")
		if requestPath == "" {
			requestPath = "index.html"
		}

		if file, openErr := webDistFS.Open(requestPath); openErr == nil {
			_ = file.Close()
			c.FileFromFS(requestPath, http.FS(webDistFS))
			return
		}

		if path.Ext(requestPath) != "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		serveAdminIndex(c, webDistFS)
	})

	r.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if strings.HasPrefix(requestPath, "/api/") || strings.HasPrefix(requestPath, "/v1/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		if strings.HasPrefix(requestPath, "/admin/") || requestPath == "/admin" {
			assetPath := strings.TrimPrefix(requestPath, "/admin/")
			if assetPath == "" || strings.HasSuffix(assetPath, "/") {
				serveAdminIndex(c, webDistFS)
				return
			}

			normalized := path.Clean(assetPath)
			if normalized == "." {
				normalized = "index.html"
			}

			if file, err := webDistFS.Open(normalized); err == nil {
				_ = file.Close()
				c.FileFromFS(normalized, http.FS(webDistFS))
				return
			}

			if path.Ext(normalized) != "" {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}

			serveAdminIndex(c, webDistFS)
			return
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
}
