package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"

	storymetadata_v1 "github.com/koneal2013/storymetadatagenerator/api/v1"
	"github.com/koneal2013/storymetadatagenerator/internal/middleware/adaptor"
)

type HttpConfig struct {
	Port            int
	MiddlewareFuncs []mux.MiddlewareFunc
}

func NewHTTPServer(cfg *HttpConfig) *http.Server {
	r := mux.NewRouter()
	r.HandleFunc("/v1/story_metadata", adaptor.GenericHttpAdaptor(handleGetStoryMetadata)).Methods(http.MethodGet)
	r.HandleFunc("/status", handleStatus).Methods(http.MethodGet)
	r.Use(cfg.MiddlewareFuncs...)
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}
}

// Status godoc
//
//	@Description	Return 200 OK if server is ready to accept requests
//	@Success		200	{object}	string
//
//	@Router			/status [get]
func handleStatus(w http.ResponseWriter, r *http.Request) {

}

// Calculate godoc
//
//	@Description	Returns metadata about a story.
//	@Accept			json
//	@Produce		json
//	@Param			input	body		int 	true	"Number of story pages to retrieve. Defaults to 1 if '0' is provided."
//	@Success		200		{object}	string
//	@Failure		400		{object}	string
//	@Failure		404		{object}	string
//	@Router			/v1/story_metadata [get]
func handleGetStoryMetadata(ctx context.Context, in int) (out *storymetadata_v1.StoryMetadataResult, err error) {
	_, span := otel.GetTracerProvider().Tracer("httpTracer").Start(ctx, "/v1/story_metadata")
	defer span.End()
	resp := storymetadata_v1.New(in)
	resp.LoadStories()
	out = resp
	return
}
