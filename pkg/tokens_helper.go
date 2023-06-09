package pkg

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type tokens struct {
	hclwrite.Tokens
}

func newTokens() *tokens {
	return &tokens{
		Tokens: hclwrite.Tokens{},
	}
}

func (t *tokens) newLine() *tokens {
	t.Tokens = append(t.Tokens, &hclwrite.Token{
		Type:  hclsyntax.TokenNewline,
		Bytes: []byte("\n"),
	})
	return t
}

func (t *tokens) dot() *tokens {
	t.Tokens = append(t.Tokens, &hclwrite.Token{
		Type:  hclsyntax.TokenDot,
		Bytes: []byte("."),
	})
	return t
}

func (t *tokens) ident(ident string, spaceBefore int) *tokens {
	t.Tokens = append(t.Tokens, &hclwrite.Token{
		Type:         hclsyntax.TokenIdent,
		Bytes:        []byte(ident),
		SpacesBefore: spaceBefore,
	})
	return t
}

func (t *tokens) oHeredoc(content string) *tokens {
	t.Tokens = append(t.Tokens, &hclwrite.Token{
		Type:  hclsyntax.TokenOHeredoc,
		Bytes: []byte(content),
	})
	return t
}

func (t *tokens) cHeredoc(content string) *tokens {
	t.Tokens = append(t.Tokens, &hclwrite.Token{
		Type:  hclsyntax.TokenCHeredoc,
		Bytes: []byte(content),
	})
	return t
}

func (t *tokens) rawTokens(rawTokens hclwrite.Tokens) *tokens {
	t.Tokens = append(t.Tokens, rawTokens...)
	return t
}
