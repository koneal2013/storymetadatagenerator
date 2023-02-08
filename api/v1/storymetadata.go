package storymetadata_v1

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/neurosnap/sentences"
	"github.com/neurosnap/sentences/english"
	"github.com/pkg/errors"

	"github.com/koneal2013/storymetadatagenerator/api/util"
)

const (
	AVERAGE_ADULT_WPM = 238
	STORY_STREAM_URL  = "https://api.axios.com/api/render/stream/content/"
	STORY_URL         = "https://api.axios.com/api/render/content/"
)

type StoryStream struct {
	Count    int      `json:"count"`
	Next     *string  `json:"next"`
	Previous *string  `json:"previous"`
	Results  []string `json:"results"`
}

type StoryMetadataResultI interface {
	LoadStories() *StoryMetadataResult
}

type StoryMetadataResult struct {
	Stories              map[string]*StoryMetadata `json:"stories"`
	Errs                 []*string                 `json:"errors,omitempty"`
	errsChan             chan error
	storyMetadataResults chan *StoryMetadata
	storyStreamResults   chan *StoryStream
	resultCount          *atomic.Uint32
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

// calculateReadingTime calculates the average adult reading time based on StoryMetadata.WordCount and AVERAGE_ADULT_WPM then sets StoryMetadata.ReadingTime
func (s *StoryMetadata) calculateReadingTime() error {
	if s.WordCount == 0 {
		return fmt.Errorf("unable to calculate read time for story with id [%s]: word count is zero", s.id)
	}
	wpm := s.WordCount / AVERAGE_ADULT_WPM
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
func New(numOfStreamPages int) (*StoryMetadataResult, error) {
	chanSize := int(float64(numOfStreamPages) * 0.05)
	// load metadata for sentence tokenizer
	return &StoryMetadataResult{
		Stories:              map[string]*StoryMetadata{},
		storyMetadataResults: make(chan *StoryMetadata, chanSize),
		storyStreamResults:   make(chan *StoryStream, chanSize),
		errsChan:             make(chan error, 1),
		resultCount:          &atomic.Uint32{},
		numOfStreamPages:     numOfStreamPages,
	}, nil
}

// LoadStories gets and processes the number of stream pages (set by New).
// LoadStories calls StoryMetadata.CalculateWordCount and StoryMetadata.CalculateReadingTime
func (sr *StoryMetadataResult) LoadStories() *StoryMetadataResult {
	if sr.numOfStreamPages < 0 {
		close(sr.storyMetadataResults)
		close(sr.errsChan)
		return sr
	}
	if sr.numOfStreamPages == 0 {
		sr.numOfStreamPages = 1
	}
	// get story streams with provided number of pages
	go sr.getStoryStream(sr.numOfStreamPages)
	go sr.getStoryMetadata()
	for {
		select {
		case metadata, ok := <-sr.storyMetadataResults:
			if !ok {
				close(sr.errsChan)
				return sr
			}
			sr.Stories[metadata.id] = metadata
		case err := <-sr.errsChan:
			storyErr := ErrorStoryMetadata{
				Err: err,
			}.Error()
			sr.Errs = append(sr.Errs, &storyErr)
			// increment the result count when an error occurs to ensure all goroutines return
			sr.resultCount.Add(1)
		default:
			if sr.resultCount.Load() == uint32(sr.numOfStreamPages*10) {
				close(sr.storyMetadataResults)
				close(sr.errsChan)
				return sr
			}
		}
	}
}

// getStoryMetadata creates StoryMetadata objects for each story in StoryMetadataResult.storyStreamResults and sends them to the StoryMetadataResult.storyMetadataResults channel
func (sr *StoryMetadataResult) getStoryMetadata() {
	// get individual story metadata
	for storyStream := range sr.storyStreamResults {
		for _, storyId := range storyStream.Results {
			go func(id string) {
				metadata := &StoryMetadata{}
				err := getResource(fmt.Sprintf("%s%s/", STORY_URL, id), metadata)
				if err != nil {
					sr.errsChan <- err
					return
				}
				metadata.id = id
				// load and instantiate the sentence tokenizer for each goroutine
				t, err := english.NewSentenceTokenizer(nil)
				if err != nil {
					sr.errsChan <- errors.WithMessagef(err, "could not load sentence tokenizer for story id [%s]", id)
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
				sr.storyMetadataResults <- metadata
				sr.resultCount.Add(1)
				return
			}(storyId)
		}
	}
}

// getStoryStream gets the number of story stream pages based on the value in StoryMetadataResult.numOfStreamPages and creates StoryStream objects for each story. Each StoryStream object is then pushed to the StoryMetadataResult.storyStreamResults channel
func (sr *StoryMetadataResult) getStoryStream(numOfStreamPages int) {
	defer close(sr.storyStreamResults)
	storyStream := &StoryStream{}
	err := getResource(STORY_STREAM_URL, storyStream)
	if err != nil {
		sr.errsChan <- err
		// if the initial story stream retrieve fails, no further stories can be processed, so we close the metadata results channel
		close(sr.storyMetadataResults)
		return
	}
	sr.storyStreamResults <- storyStream
	for storyStream.Next != nil && numOfStreamPages != 1 {
		if err = getResource(*storyStream.Next, storyStream); err != nil {
			sr.errsChan <- err
			numOfStreamPages--
			return
		}
		sr.storyStreamResults <- storyStream
		numOfStreamPages--
	}
}

func getResource[T StoryStream | StoryMetadata](url string, resource *T) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.WithMessagef(err, "error getting resource located at [%s]: ", url)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("resource located at [%s] could not be found", url)
	}
	if err = json.NewDecoder(resp.Body).Decode(resource); err != nil {
		return errors.WithMessagef(err, "error getting resource located at [%s]: ", url)
	}
	return nil
}
