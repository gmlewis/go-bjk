package ast

import (
	_ "embed"
	"testing"

	"github.com/alecthomas/participle/v2"
)

//go:embed "testdata/bifilar-electromagnet.bjk"
var testFile string

func TestParse(t *testing.T) {
	parser, err := participle.Build[BJK]()
	if err != nil {
		t.Fatal(err)
	}

	ast, err := parser.ParseString("", testFile)
	if err != nil {
		t.Fatal(err)
	}

	t.Errorf("ast=%#v", ast)
}
