package errors

import (
	"fmt"
	"os"

	stderrors "errors"
)

type SynkroError struct {
	Code    string
	Message string
	Help    string
	Err     error
}

func (e *SynkroError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (code: %s)", e.Message, e.Err.Error(), e.Code)
	}
	return fmt.Sprintf("%s (code: %s)", e.Message, e.Code)
}

func (e *SynkroError) Unwrap() error {
	return e.Err
}

func DisplayError(err error) {
	if se, ok := err.(*SynkroError); ok {
		fmt.Fprintln(os.Stderr, "Error:", se.Message)
		fmt.Fprintln(os.Stderr, "  Code:", se.Code)
		fmt.Fprintln(os.Stderr, "  Help:", se.Help)
	} else {
		fmt.Fprintln(os.Stderr, "Unexpected error:", err.Error())
		fmt.Fprintln(os.Stderr, "  Please report: https://github.com/rodascaar/synkro/issues")
	}
}

func Is(err error, code string) bool {
	var se *SynkroError
	if stderrors.As(err, &se) {
		return se.Code == code
	}
	return false
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
	ErrDatabaseNotFound = &SynkroError{
		Code:    "DB_NOT_FOUND",
		Message: "Database not found. Please run 'synkro init' first.",
		Help:    "Run: synkro init",
	}
	ErrDatabaseLocked = &SynkroError{
		Code:    "DB_LOCKED",
		Message: "Database is locked by another process.",
		Help:    "Close other Synkro instances or delete .db-wal files.",
	}
	ErrMCPNotConfigured = &SynkroError{
		Code:    "MCP_NOT_CONFIGURED",
		Message: "MCP server is not configured in your IDE.",
		Help:    "Add synkro to MCP servers configuration in your IDE settings.",
	}
	ErrFTS5NotAvailable = &SynkroError{
		Code:    "FTS5_NOT_AVAILABLE",
		Message: "FTS5 not available. SQLite was compiled without FTS5 support.",
		Help:    "Rebuild Synkro with: CGO_ENABLED=1 go build -tags sqlite_fts5",
	}
	ErrTerminalTooSmall = &SynkroError{
		Code:    "TERMINAL_TOO_SMALL",
		Message: "Terminal is too small for TUI (minimum: 120x40).",
		Help:    "Resize your terminal or use: resize -s 120 40",
	}
	ErrMemoryNotFound = &SynkroError{
		Code:    "MEM_NOT_FOUND",
		Message: "Memory not found",
		Help:    "Check the ID and try again",
	}
	ErrInvalidInput = &SynkroError{
		Code:    "INVALID_INPUT",
		Message: "Invalid input",
		Help:    "Check the required fields and try again",
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
