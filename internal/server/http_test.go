package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	storymetadata_v1 "github.com/koneal2013/storymetadatagenerator/api/v1"
	"github.com/koneal2013/storymetadatagenerator/internal/middleware/adaptor"
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
	req := httptest.NewRequest("GET", "/v1/story_metadata", strings.NewReader("1"))
	rr := httptest.NewRecorder()
	handler := adaptor.GenericHttpAdaptor(handleGetStoryMetadata)
	handler.ServeHTTP(rr, req)
	metadata := &storymetadata_v1.StoryMetadataResult{}
	err := json.NewDecoder(rr.Body).Decode(metadata)
	require.NoError(t, err)
	expectedStoryLength := 10
	require.Equal(t, expectedStoryLength, len(metadata.Stories))
	require.Nil(t, metadata.Errs)
	require.Equal(t, rr.Code, http.StatusOK)
}

func testGetStatus(t *testing.T) {
	req := httptest.NewRequest("GET", "/status", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleStatus)
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusOK)
}
