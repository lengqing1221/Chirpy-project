	package main

	import (
		"log"
		"net/http"
		"sync/atomic"
		"github.com/lib/pq"	
		"fmt"
		"encoding/json"
		"strings"
	)

	type apiConfig struct {
		// variable allows for concurrent access
		fileserverHits atomic.Int32
	}

	func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cfg.fileserverHits.Add(1)
			next.ServeHTTP(w, r)
		})
	}

	func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
		htmlContent := fmt.Sprintf(`
				<html>
				<body>
					<h1>Welcome, Chirpy Admin</h1>
					<p>Chirpy has been visited %d times!</p>
				</body>
				</html>`, cfg.fileserverHits.Load())
		w.Write([]byte(htmlContent))
	}

	func dataHandler(w http.ResponseWriter, r *http.Request) {
		type parameter struct {
			Body string `json:"body"`
		}
		
		type response struct {
			CleanedBody string `json:"cleaned_body"`
			Valid bool `json:"valid"`
		}

		decode := json.NewDecoder(r.Body)

		defer r.Body.Close()
		var d parameter
		if err := decode.Decode(&d); err != nil {
			http.Error(w, `{"error: Something went wrong"}`, http.StatusInternalServerError, )
			return
		}
		length := len(d.Body)

		if length > 140 {
			http.Error(w, `{"error":"Chirp is too long"}`, http.StatusBadRequest)
			return
		}
		
		
		cleaned_body := strings.Replace(d.Body, "Fornax", "****", -1)
		cleaned_body = strings.Replace(cleaned_body, "kerfuffle", "****", -1)
		cleaned_body = strings.Replace(cleaned_body, "sharbert", "****", -1)
		res := response {
			CleanedBody: cleaned_body,
			Valid: true,
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}


	func main() {
		const port = "8080"

		// initialize empty config
		apiCfg := apiConfig{
			fileserverHits: atomic.Int32{},
		}

		// Creating server
		mux := http.NewServeMux()
		mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
		mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
		mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
		mux.HandleFunc("POST /api/validate_chirp", dataHandler)
		// Readiness endpoint
		mux.HandleFunc("GET /api/healthz", readinessHandler)
		
		srv := &http.Server{
			Addr: ":" + port,
			Handler: mux,
		}

		log.Println("Listening on port: %s\n", port)
		log.Fatal(srv.ListenAndServe())
	}
