package errors

import (
	"fmt"
	"os"
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
		fmt.Fprintln(os.Stderr, "❌ Error:", se.Message)
		fmt.Fprintln(os.Stderr, "   Code:", se.Code)
		fmt.Fprintln(os.Stderr, "\n💡 Help:")
		fmt.Fprintln(os.Stderr, "  ", se.Help)
		fmt.Fprintln(os.Stderr, "\n📚 Documentation: https://github.com/rodascaar/synkro#troubleshooting")
	} else {
		fmt.Fprintln(os.Stderr, "❌ Unexpected error:", err.Error())
		fmt.Fprintln(os.Stderr, "\n💡 Please report this issue:")
		fmt.Fprintln(os.Stderr, "  https://github.com/rodascaar/synkro/issues")
	}
}

// Predefined errors
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
)
