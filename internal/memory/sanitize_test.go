package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeFTS5Query(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple query",
			input: "hello world",
			want:  `"hello" "world"`,
		},
		{
			name:  "single word",
			input: "search",
			want:  `"search"`,
		},
		{
			name:  "empty string",
			input: "",
			want:  `""`,
		},
		{
			name:  "only special characters",
			input: `" * ( ) :`,
			want:  `""`,
		},
		{
			name:  "double quotes around words",
			input: `"hello" "world"`,
			want:  `"hello" "world"`,
		},
		{
			name:  "asterisk wildcard",
			input: `prefix*`,
			want:  `"prefix"`,
		},
		{
			name:  "parentheses",
			input: `(group1 OR group2)`,
			want:  `"group1" "OR" "group2"`,
		},
		{
			name:  "mixed special and normal",
			input: `find "this" AND *that* :stuff`,
			want:  `"find" "this" "AND" "that" "stuff"`,
		},
		{
			name:  "multiple spaces",
			input: "hello    world",
			want:  `"hello" "world"`,
		},
		{
			name:  "AND OR NOT keywords",
			input: "memory AND context OR search NOT old",
			want:  `"memory" "AND" "context" "OR" "search" "NOT" "old"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFTS5Query(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
