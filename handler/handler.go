package handler

import (
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/tlanfer/SpoolToTag/openai"
)

//go:embed static
var staticFiles embed.FS

const maxImageSize = 20 << 20 // 20 MB

func New(analyzer openai.Analyzer) http.Handler {
	mux := http.NewServeMux()

	staticFS, _ := fs.Sub(staticFiles, "static")
	mux.Handle("/", http.FileServer(http.FS(staticFS)))
	mux.HandleFunc("POST /api/analyze", analyzeHandler(analyzer))

	return mux
}

func analyzeHandler(analyzer openai.Analyzer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(maxImageSize); err != nil {
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "missing image field", http.StatusBadRequest)
			return
		}
		defer file.Close()

		imageData, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "failed to read image", http.StatusInternalServerError)
			return
		}

		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}

		spool, err := analyzer.Analyze(r.Context(), imageData, contentType)
		if err != nil {
			log.Printf("analyze error: %v", err)
			http.Error(w, "analysis failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(spool)
	}
}
