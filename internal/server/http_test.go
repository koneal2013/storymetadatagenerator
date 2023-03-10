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

const (
	succeed = "\u2714"
	failed  = "\u2718"
)

func TestHttpServer(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"get status returns '200 ok'":                                 testGetStatus,
		"get story metadata for 'N' number of pages returns '200 ok'": testGetStoryMetadataSuccess,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
			if t.Failed() {
				t.Logf("%s %s", scenario, failed)
				return
			}
			t.Logf("%s %s", scenario, succeed)
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
	svr := testServer(handler)
	defer svr.Close()

	req, err := http.NewRequest("GET", svr.URL, strings.NewReader("1"))
	require.NoError(t, err)

	client := http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)

	metadata := &storymetadata_v1.StoryMetadataResult{}
	err = json.NewDecoder(resp.Body).Decode(metadata)
	require.NoError(t, err)
	expectedStoryLength := 10
	require.Equal(t, expectedStoryLength, len(metadata.Stories))
	require.Nil(t, metadata.Errs)
	require.Equal(t, resp.StatusCode, http.StatusOK)
}

func testGetStatus(t *testing.T) {
	svrDeps := server.HttpSvrDeps{
		HttpTracer: otel.GetTracerProvider().Tracer("Test http trace"),
	}
	handler := http.HandlerFunc(svrDeps.HandleStatus)

	svr := testServer(handler)
	defer svr.Close()

	resp, err := http.Get(svr.URL)
	require.NoError(t, err)

	require.Equal(t, resp.StatusCode, http.StatusOK)
}

func testServer(f func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(f))
}

// mock server for instances where no internet connection is available
func mockServer() *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(storymetadata_v1.StoryMetadataResult{
			Stories: map[string]storymetadata_v1.StoryMetadata{
				"testStory": {
					WordCount:        10,
					ReadabilityScore: "12",
					ReadingTime:      "<1",
					Headline:         "testHeadline",
					Permalink:        "testPermalink",
				},
			},
		})
		w.Header().Add("Content-Type", "application/json")
	}
	return httptest.NewServer(http.HandlerFunc(f))
}
