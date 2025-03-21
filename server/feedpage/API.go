package feedpage

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogchi "github.com/samber/slog-chi"
)

type API struct {
	Log           *slog.Logger
	ContentEngine *ContentEngine
}

func (a *API) Start(ctx context.Context) error {
	r := chi.NewRouter()

	logConfig := slogchi.Config{
		DefaultLevel:       slog.LevelInfo,
		ClientErrorLevel:   slog.LevelError,
		ServerErrorLevel:   slog.LevelError,
		WithUserAgent:      true,
		WithRequestID:      true,
		WithRequestBody:    false,
		WithRequestHeader:  false,
		WithResponseBody:   false,
		WithResponseHeader: false,
		WithSpanID:         false,
		WithTraceID:        false,
		Filters:            nil,
	}

	r.Use(
		middleware.RealIP,
		slogchi.NewWithConfig(a.Log, logConfig),
		middleware.RequestID,
		middleware.Recoverer,
		middleware.Compress(5),
	)

	options := ChiServerOptions{
		BaseRouter: r,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			a.Log.Error("handler error", slog.Any("error", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
	}
	HandlerWithOptions(a, options)

	server := http.Server{
		Addr:    os.Getenv("HTTP_HOST"),
		Handler: r,
	}
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			a.Log.Error("error running server", slog.Any("error", err))
			cancel()
		}
	}()

	<-ctx.Done()
	server.Shutdown(ctx)

	return nil
}

func (a *API) GetApiPosts(w http.ResponseWriter, r *http.Request, params GetApiPostsParams) {
	posts := a.ContentEngine.GetPosts()
	pageSize := 100
	start := int(params.Page) * pageSize
	out := make([]Post, 0, len(posts))
	if len(posts) >= start {
	loop:
		for _, p := range posts[start:] {
			out = append(out, Post{
				Title:       p.title,
				Source:      p.source,
				Timestamp:   p.timestamp,
				Description: p.description,
				Url:         p.url,
				Thumbnail:   p.thumbnail,
			})
			if len(out) >= pageSize {
				break loop
			}
		}
	}

	err := json.NewEncoder(w).Encode(Posts{
		Items: out,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		a.Log.Error("error sending posts", slog.Any("error", err))
	}
}
