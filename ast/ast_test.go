package ast

import (
	_ "embed"
	"os"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/google/go-cmp/cmp"
)

//go:embed "testdata/bifilar-electromagnet.bjk"
var testFile string

const header = `// BLACKJACK_VERSION_HEADER 0 0 0
`

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *BJK
	}{
		{
			name:  "no graph",
			input: header + "( )",
			want:  &BJK{},
		},
		{
			name:  "no nodes",
			input: header + "( nodes: [ ] )",
			want:  &BJK{Graph: &Graph{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			simpleLexer := lexer.MustSimple([]lexer.SimpleRule{
				{"Header", `(?:#|// BLACKJACK_VERSION_HEADER)[ ]*`},
				{"Ident", `[a-zA-Z]\w*`},
				{"Int", `(?:\d*)?\d+`},
				{"Number", `(?:\d*\.)?\d+`},
				{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
				{"Whitespace", `[ \t\n\r]+`},
			})
			parser := participle.MustBuild[BJK](
				participle.Lexer(simpleLexer),
				participle.Elide("Whitespace"),
				// participle.UseLookahead(200),
			)

			got, err := parser.ParseString("", tt.input, participle.Trace(os.Stderr))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Logf("%v\n", tt.input)
				t.Errorf("ParseString mismatch (-want +got):\n%v", diff)
			}
		})
	}
}
