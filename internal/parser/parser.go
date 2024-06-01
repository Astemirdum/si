package parser

import (
	"errors"

	"github.com/Astemirdum/si/pkg"

	"github.com/alecthomas/participle/v2"
)

func BuildParser[T any]() *participle.Parser[T] {
	return participle.MustBuild[T](
		participle.Lexer(BuildLexer()),
		participle.UseLookahead(1024*4),
		participle.Elide("Whitespace"),
		participle.Elide("MultiLineComment"),
		participle.Elide("Comment"),
	)
}

type Parser struct {
	parser *participle.Parser[Module]
}

func NewParser() *Parser {
	return &Parser{
		parser: BuildParser[Module](),
	}
}

func (p *Parser) ParseFile(file *pkg.File) (*Module, error) {
	program, err := p.parser.ParseString(file.Name, file.Contents)

	if err != nil {
		var parseError participle.Error

		if errors.As(err, &parseError) {
			return nil, pkg.WithPos(err, file, parseError.Position())
		}

		return nil, err
	}

	return program, nil
}

func (p *Parser) String() string {
	return p.parser.String()
}
