package errors

import (
	"fmt"
)

type SynkroError struct {
	Code    string
	Message string
	Help    string
	Err     error
}

const (
	CodeDBError           = "DB_ERROR"
	CodeGraphError        = "GRAPH_ERROR"
	CodeGraphNotAvailable = "GRAPH_NOT_AVAILABLE"
	CodeMarshalError      = "MARSHAL_ERROR"
)

func (e *SynkroError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (code: %s)", e.Message, e.Err.Error(), e.Code)
	}
	return fmt.Sprintf("%s (code: %s)", e.Message, e.Code)
}

func (e *SynkroError) Unwrap() error {
	return e.Err
}

func Wrap(err error, code, message, help string) *SynkroError {
	if err == nil {
		return nil
	}
	return &SynkroError{
		Code:    code,
		Message: message,
		Help:    help,
		Err:     err,
	}
}

var (
	ErrInvalidInput = &SynkroError{
		Code:    "INVALID_INPUT",
		Message: "Invalid input",
		Help:    "Check the required fields and try again",
	}
	ErrMemoryNotFound = &SynkroError{
		Code:    "MEM_NOT_FOUND",
		Message: "Memory not found",
		Help:    "Check the ID and try again",
	}
	ErrEmbeddingFailed = &SynkroError{
		Code:    "EMBED_FAILED",
		Message: "Failed to generate embedding",
		Help:    "Check the model configuration and try again",
	}
	ErrFTS5Query = &SynkroError{
		Code:    "FTS5_QUERY",
		Message: "Invalid search query",
		Help:    "Avoid special characters: * \" ( ) AND OR NOT",
	}
	ErrVecSearch = &SynkroError{
		Code:    "VEC_SEARCH",
		Message: "Vector search failed",
		Help:    "Ensure embeddings are generated for your memories",
	}
	ErrRelationNotFound = &SynkroError{
		Code:    "RELATION_NOT_FOUND",
		Message: "Relation not found",
		Help:    "Check source_id and target_id",
	}
	ErrInvalidRelationType = &SynkroError{
		Code:    "INVALID_RELATION",
		Message: "Invalid relation type",
		Help:    "Valid types: extends, depends_on, conflicts_with, example_of, part_of, related_to",
	}
)
