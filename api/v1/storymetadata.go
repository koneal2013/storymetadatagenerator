package storymetadata_v1

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
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
	id          string
	WordCount   int    `json:"word_count"`
	ReadingTime string `json:"reading_time"`
	Headline    string `json:"headline"`
	Permalink   string `json:"permalink"`
	Blocks      *Block `json:"blocks,omitempty"`
}

// calculateReadingTime calculates the average adult reading time based on StoryMetadata.WordCount and AVERAGE_ADULT_WPM then sets StoryMetadata.ReadingTime
func (s *StoryMetadata) calculateReadingTime() error {
	wpm := s.WordCount / AVERAGE_ADULT_WPM
	if s.WordCount == 0 {
		return fmt.Errorf("unable to calculate read time for story with id [%s]: word count is zero", s.id)
	}
	if wpm == 0 {
		s.ReadingTime = "<1"
		return nil
	}
	s.ReadingTime = fmt.Sprintf("%d", wpm)
	return nil
}

// CalculateWordCount counts the words in each block and sets the WordCount field on the StoryMetadata struct
func (s *StoryMetadata) calculateWordCount() error {
	for _, block := range *s.Blocks.Blocks {
		text := bytes.NewBufferString(block.Text)
		scanner := bufio.NewScanner(text)
		scanner.Split(bufio.ScanWords)

		for scanner.Scan() {
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
	chanSize := int(float64(numOfStreamPages) * 0.05)
	return &StoryMetadataResult{
		Stories:              map[string]*StoryMetadata{},
		storyMetadataResults: make(chan *StoryMetadata, chanSize),
		storyStreamResults:   make(chan *StoryStream, chanSize),
		errsChan:             make(chan error, 1),
		resultCount:          &atomic.Uint32{},
		numOfStreamPages:     numOfStreamPages,
	}
}

// processErrs reads from the StoryMetadataResult.errsChan channel and appends errors to the StoryMetadataResult.Errs slice which is returned as a part of the response
func (sr *StoryMetadataResult) processErrs() {
	for err := range sr.errsChan {
		storyErr := ErrorStoryMetadata{
			Err: err,
		}.Error()
		sr.Errs = append(sr.Errs, &storyErr)
	}
}

// LoadStories gets and processes the number of stream pages (set by New).
// LoadStories calls StoryMetadata.CalculateWordCount and StoryMetadata.CalculateReadingTime
func (sr *StoryMetadataResult) LoadStories() {
	if sr.numOfStreamPages < 0 {
		close(sr.storyMetadataResults)
		close(sr.errsChan)
		return
	}
	if sr.numOfStreamPages == 0 {
		sr.numOfStreamPages = 1
	}
	// get story streams with provided number of pages
	go sr.getStoryStream(sr.numOfStreamPages)
	go sr.getStoryMetadata()
	for {
		select {
		case metadata := <-sr.storyMetadataResults:
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
				return
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
				defer sr.resultCount.Add(1)
				metadata := &StoryMetadata{}
				err := getResource(fmt.Sprintf("%s%s/", STORY_URL, id), metadata)
				if err != nil {
					sr.errsChan <- err
					return
				}
				metadata.id = id
				err = metadata.calculateWordCount()
				if err != nil {
					sr.errsChan <- err
					return
				}
				err = metadata.calculateReadingTime()
				if err != nil {
					sr.errsChan <- err
					return
				}
				metadata.Blocks = nil
				sr.storyMetadataResults <- metadata
				return
			}(storyId)
		}
	}
}

// getStoryStream gets the number of story stream pages based on the value in StoryMetadataResult.numOfStreamPages and creates StoryStream objects for each story. Each StoryStream object is then pushed to the StoryMetadataResult.storyStreamResults channel
func (sr *StoryMetadataResult) getStoryStream(numOfStreamPages int) {
	defer close(sr.storyStreamResults)
	storyStream := &StoryStream{}
	resp, err := http.Get(STORY_STREAM_URL)
	if err != nil {
		sr.errsChan <- err
		return
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(storyStream); err != nil {
		sr.errsChan <- err
	}
	sr.storyStreamResults <- storyStream
	for storyStream.Next != nil && numOfStreamPages != 1 {
		if err = getResource(*storyStream.Next, storyStream); err != nil {
			sr.errsChan <- err
			return
		}
		sr.storyStreamResults <- storyStream
		numOfStreamPages--
	}
}

func getResource[T StoryStream | StoryMetadata](url string, resource *T) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return fmt.Errorf("resource located at [%s] could not be found", url)
	}
	if err = json.NewDecoder(resp.Body).Decode(resource); err != nil {
		return err
	}
	return nil
}
