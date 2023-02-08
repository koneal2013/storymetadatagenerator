package storymetadata_v1

import (
	"os"
	"testing"

	"github.com/neurosnap/sentences/english"
	"github.com/stretchr/testify/require"
)

func TestCalculateCount(t *testing.T) {
	metadata := StoryMetadata{Blocks: &Block{Blocks: &[]BlockInternal{{
		Text: "Hi there. Does this really work? The U.S. housing market is at a record high. A gallon of gas averages $3.28."},
	}}}
	// change working dir to project root
	err := os.Chdir("./../..")
	require.NoError(t, err)

	tokenizer, err := english.NewSentenceTokenizer(nil)
	require.NoError(t, err)

	err = metadata.calculateCount(tokenizer)
	require.NoError(t, err)

	expectedSentenceCount := 4
	expectedWordCount := 21

	require.Equal(t, expectedSentenceCount, metadata.sentenceCount)
	require.Equal(t, expectedWordCount, metadata.WordCount)
}

func TestCalculateReadabilityScore(t *testing.T) {
	metadata := StoryMetadata{
		WordCount:          242,
		difficultWordCount: 115,
		sentenceCount:      29,
	}
	err := metadata.calculateReadabilityScore()
	require.NoError(t, err)

	expectedReadabilityScore := "11.6"

	require.Equal(t, expectedReadabilityScore, metadata.ReadabilityScore)

}
