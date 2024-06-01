package parser_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Astemirdum/si/internal/parser"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/suite"
)

type LexerTestSuite struct {
	suite.Suite
	lexer *lexer.StatefulDefinition

	symbols        map[string]lexer.TokenType
	reverseSymbols map[lexer.TokenType]string
}

func TestLexerTestSuite(t *testing.T) {
	suite.Run(t, new(LexerTestSuite))
}

func (suite *LexerTestSuite) SetupTest() {
	suite.lexer = parser.BuildLexer()
	suite.symbols = suite.lexer.Symbols()
	suite.reverseSymbols = make(map[lexer.TokenType]string)

	for k, v := range suite.symbols {
		suite.reverseSymbols[v] = k
	}
}

func (suite *LexerTestSuite) EqualToken(l lexer.Lexer, token, value string) {
	tok, err := l.Next()
	suite.NoError(err)
	suite.Equal(token, suite.reverseSymbols[tok.Type])
	suite.Equal(value, tok.Value)
}

func (suite *LexerTestSuite) dumpAllTokens(tokens lexer.Lexer) {
	for {
		tok, err := tokens.Next()
		require.NoError(suite.T(), err, fmt.Sprintf("suite.EqualToken(tokens, \"%s\", `%s`)", suite.reverseSymbols[tok.Type], tok.Value))
		if tok.EOF() {
			break
		}
		fmt.Println(suite.reverseSymbols[tok.Type], tok.Value)
	}
}

func (suite *LexerTestSuite) TestSimple() {
	tokens, err := suite.lexer.LexString("main.c", "int main() { return 0; }")
	suite.NoError(err)

	suite.EqualToken(tokens, "Ident", `int`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `main`)
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `{`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Keyword", `return`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Number", `0`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `}`)
}

func (suite *LexerTestSuite) TestFor() {
	tokens, err := suite.lexer.LexString("main.c", `
	i64 x;
	for (x=0; x<10; x++;){
		printf("%d\n", x);
	}`)
	suite.NoError(err)

	suite.EqualToken(tokens, "Whitespace", "\n\t")
	suite.EqualToken(tokens, "BasicType", `i64`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", "\n\t")
	suite.EqualToken(tokens, "Keyword", `for`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Punct", `=`)
	suite.EqualToken(tokens, "Number", `0`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Punct", `<`)
	suite.EqualToken(tokens, "Number", `10`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Punct", `+`)
	suite.EqualToken(tokens, "Punct", `+`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Punct", `{`)
	suite.EqualToken(tokens, "Whitespace", "\n\t\t")
	suite.EqualToken(tokens, "Ident", `printf`)
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "StringStart", `"`)
	suite.EqualToken(tokens, "Chars", `%d`)
	suite.EqualToken(tokens, "Escaped", `\n`)
	suite.EqualToken(tokens, "StringEnd", `"`)
	suite.EqualToken(tokens, "Punct", `,`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", "\n\t")
	suite.EqualToken(tokens, "Punct", `}`)
}

func (suite *LexerTestSuite) TestMultiLineComment() {
	tokens, err := suite.lexer.LexString("main.c", `
int main() {
	/* a
	aa
	*/
	return 0;
}
`)
	suite.NoError(err)

	suite.EqualToken(tokens, "Whitespace", `
`)
	suite.EqualToken(tokens, "Ident", `int`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `main`)
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `{`)
	suite.EqualToken(tokens, "Whitespace", `
	`)
	suite.EqualToken(tokens, "MultiLineComment", `/* a
	aa
	*/`)
	suite.EqualToken(tokens, "Whitespace", `
	`)
	suite.EqualToken(tokens, "Keyword", `return`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Number", `0`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", `
`)
	suite.EqualToken(tokens, "Punct", `}`)
	suite.EqualToken(tokens, "Whitespace", `
`)
}

func (suite *LexerTestSuite) TestString() {
	tokens, err := suite.lexer.LexString("main.c", `int main() { return "h\"ali" + "hello"; }`)
	suite.NoError(err)

	suite.EqualToken(tokens, "Ident", `int`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `main`)
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `{`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Keyword", `return`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "StringStart", `"`)
	suite.EqualToken(tokens, "Chars", `h`)
	suite.EqualToken(tokens, "Escaped", `\"`)
	suite.EqualToken(tokens, "Chars", `ali`)
	suite.EqualToken(tokens, "StringEnd", `"`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `+`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "StringStart", `"`)
	suite.EqualToken(tokens, "Chars", `hello`)
	suite.EqualToken(tokens, "StringEnd", `"`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `}`)
}

func (suite *LexerTestSuite) TestComment() {
	tokens, err := suite.lexer.LexString("main.c", `
int main() { // huha
	// hello
	return 0; // hallo
}
`)
	suite.NoError(err)

	suite.EqualToken(tokens, "Whitespace", `
`)
	suite.EqualToken(tokens, "Ident", `int`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `main`)
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `{`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Comment", `// huha
`)
	suite.EqualToken(tokens, "Whitespace", `	`)
	suite.EqualToken(tokens, "Comment", `// hello
`)
	suite.EqualToken(tokens, "Whitespace", `	`)
	suite.EqualToken(tokens, "Keyword", `return`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Number", `0`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Comment", `// hallo
`)
	suite.EqualToken(tokens, "Punct", `}`)
	suite.EqualToken(tokens, "Whitespace", `
`)
}

func (suite *LexerTestSuite) TestFnCall() {
	tokens, err := suite.lexer.LexString("main.c", `int main() { abc(); }`)
	suite.NoError(err)

	suite.EqualToken(tokens, "Ident", `int`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `main`)
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `{`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `abc`)
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `}`)
}

func (suite *LexerTestSuite) TestVoidKeyword() {
	tokens, err := suite.lexer.LexString("main.c", `
	void voidfn()
	{
		return;
	}	
	`)
	suite.NoError(err)

	suite.EqualToken(tokens, "Whitespace", `
	`)
	suite.EqualToken(tokens, "BasicType", `void`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `voidfn`)
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Whitespace", `
	`)
	suite.EqualToken(tokens, "Punct", `{`)
	suite.EqualToken(tokens, "Whitespace", `
		`)
	suite.EqualToken(tokens, "Keyword", `return`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", `
	`)
	suite.EqualToken(tokens, "Punct", `}`)
	suite.EqualToken(tokens, "Whitespace", `	
	`)
}

func (suite *LexerTestSuite) TestIndexExpr() {
	tokens, err := suite.lexer.LexString("main.c", `int main() { c[12]; }`)
	suite.NoError(err)

	suite.EqualToken(tokens, "Ident", `int`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `main`)
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `{`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `c`)
	suite.EqualToken(tokens, "Punct", `[`)
	suite.EqualToken(tokens, "Number", `12`)
	suite.EqualToken(tokens, "Punct", `]`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `}`)
}

func (suite *LexerTestSuite) TestNull() {
	tokens, err := suite.lexer.LexString("main.c", `int main() { i8 *a=NULL; }`)
	suite.NoError(err)

	//suite.dumpAllTokens(tokens)

	suite.EqualToken(tokens, "Ident", `int`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `main`)
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `{`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "BasicType", `i8`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `*`)
	suite.EqualToken(tokens, "Ident", `a`)
	suite.EqualToken(tokens, "Punct", `=`)
	suite.EqualToken(tokens, "Null", `NULL`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `}`)
}

func (suite *LexerTestSuite) TestContinue() {
	tokens, err := suite.lexer.LexString("main.c", `
	i64 x;
	for (x=0; x<10; x++;){
		if (x < 5) {continue;};
	}`)
	suite.NoError(err)

	suite.EqualToken(tokens, "Whitespace", "\n\t")
	suite.EqualToken(tokens, "BasicType", `i64`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", "\n\t")
	suite.EqualToken(tokens, "Keyword", `for`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Punct", `=`)
	suite.EqualToken(tokens, "Number", `0`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Punct", `<`)
	suite.EqualToken(tokens, "Number", `10`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Punct", `+`)
	suite.EqualToken(tokens, "Punct", `+`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Punct", `{`)
	suite.EqualToken(tokens, "Whitespace", "\n\t\t")
	suite.EqualToken(tokens, "Keyword", `if`)
	suite.EqualToken(tokens, "Whitespace", " ")
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Whitespace", " ")
	suite.EqualToken(tokens, "Punct", `<`)
	suite.EqualToken(tokens, "Whitespace", " ")
	suite.EqualToken(tokens, "Number", `5`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Whitespace", " ")
	suite.EqualToken(tokens, "Punct", `{`)
	suite.EqualToken(tokens, "Keyword", `continue`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Punct", `}`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", "\n\t")
	suite.EqualToken(tokens, "Punct", `}`)
}

func (suite *LexerTestSuite) TestBreak() {
	tokens, err := suite.lexer.LexString("main.c", `
	i64 x;
	for (x=0; x<10; x++;){
		if (x < 5) {break;};
	}`)
	suite.NoError(err)

	suite.EqualToken(tokens, "Whitespace", "\n\t")
	suite.EqualToken(tokens, "BasicType", `i64`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", "\n\t")
	suite.EqualToken(tokens, "Keyword", `for`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Punct", `=`)
	suite.EqualToken(tokens, "Number", `0`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Punct", `<`)
	suite.EqualToken(tokens, "Number", `10`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", ` `)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Punct", `+`)
	suite.EqualToken(tokens, "Punct", `+`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Punct", `{`)
	suite.EqualToken(tokens, "Whitespace", "\n\t\t")
	suite.EqualToken(tokens, "Keyword", `if`)
	suite.EqualToken(tokens, "Whitespace", " ")
	suite.EqualToken(tokens, "Punct", `(`)
	suite.EqualToken(tokens, "Ident", `x`)
	suite.EqualToken(tokens, "Whitespace", " ")
	suite.EqualToken(tokens, "Punct", `<`)
	suite.EqualToken(tokens, "Whitespace", " ")
	suite.EqualToken(tokens, "Number", `5`)
	suite.EqualToken(tokens, "Punct", `)`)
	suite.EqualToken(tokens, "Whitespace", " ")
	suite.EqualToken(tokens, "Punct", `{`)
	suite.EqualToken(tokens, "Keyword", `break`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Punct", `}`)
	suite.EqualToken(tokens, "Punct", `;`)
	suite.EqualToken(tokens, "Whitespace", "\n\t")
	suite.EqualToken(tokens, "Punct", `}`)
}

func (suite *LexerTestSuite) TestPrefixExpr() {
	tokens, err := suite.lexer.LexString("main.c", `++ha`)
	suite.NoError(err)

	suite.EqualToken(tokens, "Punct", `+`)
	suite.EqualToken(tokens, "Punct", `+`)
	suite.EqualToken(tokens, "Ident", `ha`)

	tokens, err = suite.lexer.LexString("main.c", `[ha]`)
	suite.NoError(err)

	suite.EqualToken(tokens, "Punct", `[`)
	suite.EqualToken(tokens, "Ident", `ha`)
	suite.EqualToken(tokens, "Punct", `]`)

	tokens, err = suite.lexer.LexString("main.c", `---ha---`)
	suite.NoError(err)

	suite.EqualToken(tokens, "Punct", `-`)
	suite.EqualToken(tokens, "Punct", `-`)
	suite.EqualToken(tokens, "Punct", `-`)
	suite.EqualToken(tokens, "Ident", `ha`)
	suite.EqualToken(tokens, "Punct", `-`)
	suite.EqualToken(tokens, "Punct", `-`)
	suite.EqualToken(tokens, "Punct", `-`)
}
