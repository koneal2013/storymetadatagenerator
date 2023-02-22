package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	storymetadata_v1 "github.com/koneal2013/storymetadatagenerator/api/v1"
	"github.com/koneal2013/storymetadatagenerator/internal/middleware/adaptor"
	"github.com/koneal2013/storymetadatagenerator/internal/server"
)

func TestHttpServer(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"get status returns '200 ok'":                                 testGetStatus,
		"get story metadata for 'N' number of pages returns '200 ok'": testGetStoryMetadataSuccess,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

func testGetStoryMetadataSuccess(t *testing.T) {
	// change working dir to project root
	err := os.Chdir("./../..")
	require.NoError(t, err)

	svrDeps := server.HttpSvrDeps{
		HttpTracer: otel.GetTracerProvider().Tracer("Test http trace"),
	}

	handler := adaptor.GenericHttpAdaptor(svrDeps.HandleGetStoryMetadata)
	svr := mockServer(handler)
	defer svr.Close()

	req, err := http.NewRequest("GET", svr.URL, strings.NewReader("2"))
	require.NoError(t, err)

	client := http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)

	metadata := &storymetadata_v1.StoryMetadataResult{}
	err = json.NewDecoder(resp.Body).Decode(metadata)
	require.NoError(t, err)
	expectedStoryLength := 20
	require.Equal(t, expectedStoryLength, len(metadata.Stories))
	require.Nil(t, metadata.Errs)
	require.Equal(t, resp.StatusCode, http.StatusOK)
}

func testGetStatus(t *testing.T) {
	svrDeps := server.HttpSvrDeps{
		HttpTracer: otel.GetTracerProvider().Tracer("Test http trace"),
	}
	handler := http.HandlerFunc(svrDeps.HandleStatus)

	svr := mockServer(handler)
	defer svr.Close()

	resp, err := http.Get(svr.URL)
	require.NoError(t, err)

	require.Equal(t, resp.StatusCode, http.StatusOK)
}

func mockServer(f func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(f))
}
