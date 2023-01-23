package storymetadata_v1

type ErrorStoryMetadata struct {
	Err error `json:"error"`
}

func (e ErrorStoryMetadata) Error() string {
	return e.Err.Error()
}
