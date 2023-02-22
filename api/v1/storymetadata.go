package storymetadata_v1

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/neurosnap/sentences"
	"github.com/neurosnap/sentences/english"
	"github.com/pkg/errors"

	"github.com/koneal2013/storymetadatagenerator/api/util"
)

const (
	AverageAdultWpm = 238
	StoryStreamUrl  = "https://api.axios.com/api/render/stream/content/"
	StoryUrlBase    = "https://api.axios.com/api/render/content/"
)

type StoryStream struct {
	Count    int      `json:"count"`
	Next     *string  `json:"next"`
	Previous *string  `json:"previous"`
	Results  []string `json:"results"`
}

type StoryMetadataResult struct {
	Stories              map[string]StoryMetadata `json:"stories"`
	StoryCount           int                      `json:"storyCount"`
	Errs                 []string                 `json:"errors,omitempty"`
	errsChan             chan error
	storyMetadataResults chan StoryMetadata
	storyStreamResults   chan StoryStream
	numOfStreamPages     int
}

type BlockInternal struct {
	Text string `json:"text"`
}
type Block struct {
	Blocks *[]BlockInternal `json:"blocks"`
}

type StoryMetadata struct {
	id                 string
	WordCount          int `json:"word_count"`
	sentenceCount      int
	difficultWordCount int
	ReadabilityScore   string `json:"readability_score"`
	ReadingTime        string `json:"reading_time"`
	Headline           string `json:"headline"`
	Permalink          string `json:"permalink"`
	Blocks             *Block `json:"blocks,omitempty"`
}

// calculateReadabilityScore sets StoryMetadata.ReadabilityScore
func (s *StoryMetadata) calculateReadabilityScore() error {
	if s.sentenceCount == 0 || s.WordCount == 0 {
		return fmt.Errorf("unable to calculate readability sscore for story with id [%s]: word count or sentence count is zero", s.id)
	}
	difficultWordPercentage := (float32(s.difficultWordCount) / float32(s.WordCount)) * 100
	averageSentenceLength := float32(s.WordCount) / float32(s.sentenceCount)
	if difficultWordPercentage > 5 {
		// if difficultWordPercentage is above 5% add 3.6365 to readability score
		s.ReadabilityScore = fmt.Sprintf("%.1f", ((0.1579*difficultWordPercentage)+(0.0496*averageSentenceLength))+3.6365)
		return nil
	}
	s.ReadabilityScore = fmt.Sprintf("%.1f", (0.1579*difficultWordPercentage)+(0.0496*averageSentenceLength))
	return nil
}

// calculateReadingTime calculates the average adult reading time based on StoryMetadata.WordCount and AverageAdultWpm then sets StoryMetadata.ReadingTime
func (s *StoryMetadata) calculateReadingTime() error {
	if s.WordCount == 0 {
		return fmt.Errorf("unable to calculate read time for story with id [%s]: word count is zero", s.id)
	}
	wpm := s.WordCount / AverageAdultWpm
	if wpm == 0 {
		s.ReadingTime = "<1"
		return nil
	}
	s.ReadingTime = fmt.Sprintf("%d", wpm)
	return nil
}

// calculateCount counts the words and sentences in each block and sets the WordCount & sentenceCount fields on the StoryMetadata struct
func (s *StoryMetadata) calculateCount(tokenizer *sentences.DefaultSentenceTokenizer) error {
	nonAlphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	easyWords, err := util.GetEasyWords()
	if err != nil {
		return err
	}

	for _, block := range *s.Blocks.Blocks {
		// check for Sentences
		s.sentenceCount += len(tokenizer.Tokenize(strings.TrimSpace(block.Text)))

		text := bytes.NewBufferString(block.Text)
		scanner := bufio.NewScanner(text)
		scanner.Split(bufio.ScanWords)

		for scanner.Scan() {
			word := scanner.Text()
			// check for difficult word
			word = nonAlphanumericRegex.ReplaceAllString(word, "")
			if _, ok := easyWords[word]; !ok {
				s.difficultWordCount++
			}
			if err := scanner.Err(); err != nil {
				// reset word count to zero if there is an error counting the words
				s.WordCount = 0
				return fmt.Errorf("unable to calculate word count for story with id [%s]", s.id)
			}
			s.WordCount++
		}
	}
	return nil
}

// New takes the number of steam pages to retrieve as a parameter and returns a pointer to a StoryMetadataResult struct
func New(numOfStreamPages int) *StoryMetadataResult {
	return &StoryMetadataResult{
		Stories:              make(map[string]StoryMetadata),
		storyMetadataResults: make(chan StoryMetadata),
		storyStreamResults:   make(chan StoryStream, 20),
		errsChan:             make(chan error),
		numOfStreamPages:     numOfStreamPages,
	}
}

// LoadStories gets and processes the number of stream pages (set by New).
// LoadStories calls StoryMetadata.CalculateWordCount and StoryMetadata.CalculateReadingTime
func (sr *StoryMetadataResult) LoadStories(ctx context.Context) *StoryMetadataResult {
	if sr.numOfStreamPages < 0 {
		close(sr.storyMetadataResults)
		return sr
	}
	if sr.numOfStreamPages == 0 {
		sr.numOfStreamPages = 1
	}

	workerPoolCtx, workerPoolCancel := context.WithCancel(context.Background())
	// start worker pool
	go sr.createStoryStreamWorkerPool(workerPoolCtx)

	storyStreamCtx, storyStreamCancel := context.WithCancel(context.Background())
	// get story streams with provided number of pages
	go sr.getStoryStream(storyStreamCtx, sr.numOfStreamPages)

	for {
		select {
		case metadata, ok := <-sr.storyMetadataResults:
			if !ok {
				sr.StoryCount = len(sr.Stories)
				return sr
			}
			sr.Stories[metadata.id] = metadata
		case err := <-sr.errsChan:
			storyErr := ErrorStoryMetadata{
				Err: err,
			}.Error()
			sr.Errs = append(sr.Errs, storyErr)
		case <-ctx.Done():
			storyStreamCancel()
			workerPoolCancel()
			sr.Errs = append(sr.Errs, ctx.Err().Error())
			return sr
		}
	}
}

// getStoryMetadata creates StoryMetadata objects for each story in StoryMetadataResult.storyStreamResults
// and sends them to the StoryMetadataResult.storyMetadataResults channel
func (sr *StoryMetadataResult) createStoryStreamWorkerPool(ctx context.Context) {
	var wg sync.WaitGroup
	g := runtime.NumCPU()
	wg.Add(g)

	defer close(sr.storyMetadataResults)
	defer wg.Wait()

	for g > 0 {
		go func(w *sync.WaitGroup) {
			defer w.Done()
			for stream := range sr.storyStreamResults {
				i := 0
				for i < len(stream.Results) {
					storyId := stream.Results[i]
					metadata := StoryMetadata{}
					storyUrl := fmt.Sprintf("%s%s/", StoryUrlBase, storyId)
					err := getResource(ctx, storyUrl, &metadata)
					if err != nil {
						sr.errsChan <- errors.Wrapf(err, "getStoryMetadata->getResource(%v)", storyUrl)
						return
					}
					metadata.id = storyId
					// load and instantiate the sentence tokenizer for each goroutine
					t, err := english.NewSentenceTokenizer(nil)
					if err != nil {
						sr.errsChan <- errors.Wrapf(err, "could not load sentence tokenizer for story id [%s]", storyId)
						return
					}
					err = metadata.calculateCount(t)
					if err != nil {
						sr.errsChan <- err
						return
					}
					err = metadata.calculateReadingTime()
					if err != nil {
						sr.errsChan <- err
						return
					}
					err = metadata.calculateReadabilityScore()
					if err != nil {
						sr.errsChan <- err
						return
					}
					metadata.Blocks = nil
					select {
					case <-ctx.Done():
						return
					default:
						sr.storyMetadataResults <- metadata
					}
					i++
				}
			}
		}(&wg)
		g--
	}
}

// getStoryStream gets the number of story stream pages based on the value in StoryMetadataResult.numOfStreamPages and creates StoryStream objects for each story. Each StoryStream object is then pushed to the StoryMetadataResult.storyStreamResults channel
func (sr *StoryMetadataResult) getStoryStream(ctx context.Context, numOfStreamPages int) {
	defer close(sr.storyStreamResults)
	storyStream := StoryStream{}
	err := getResource(ctx, StoryStreamUrl, &storyStream)
	if err != nil {
		sr.errsChan <- errors.Wrapf(err, "getStoryStream->getResource(%v)", StoryStreamUrl)
		return
	}
	sr.storyStreamResults <- storyStream
	for storyStream.Next != nil && numOfStreamPages != 1 {
		if err = getResource(ctx, *storyStream.Next, &storyStream); err != nil {
			sr.errsChan <- errors.Wrapf(err, "getStoryStream->getResource(%v)", *storyStream.Next)
			numOfStreamPages--
			return
		}
		sr.storyStreamResults <- storyStream
		numOfStreamPages--
	}
}

func getResource[T StoryStream | StoryMetadata](ctx context.Context, url string, resource *T) error {
	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrapf(err, "error getting resource located at [%s]: ", url)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("resource located at [%s] could not be found", url)
	}
	if err = json.NewDecoder(resp.Body).Decode(resource); err != nil {
		return errors.Wrapf(err, "error getting resource located at [%s]: ", url)
	}
	return nil
}
