package parser

import (
	"github.com/cwmz2005/LexGramAnalysis/Lexer"
)

// 语法分析器（也许也做了些语义分析的内容）
type Parser struct {
	lexer     *Lexer.Lexer
	nextToken Lexer.Token
	nowLine   int
	ASTRoot   *ASTNode
	Errors    []ParseError
}

type ParseError struct {
	Line  int // -1代表创建词法分析器时出错
	Error error
}

func Parse(fileName string) (p *Parser) {
	p = &Parser{}
	var e error
	p.lexer, e = Lexer.MakeLexer(fileName)
	defer p.lexer.Close()
	if e != nil {
		p.Errors[0] = ParseError{-1, e}
	}
	p.next()
	p.ASTRoot = p.program(nil, "")
	if p.nextToken.TType != Lexer.EOF {
		p.errNow("多余的内容：" + p.nextToken.Sign)
	}
	return p
}

func (p *Parser) next() Lexer.Token {
	// 需要预取一个Token，所以lexer用自己的特殊方法
	// if p.nextToken.TType == 0 {
	// 	p.nextToken = p.lexer.NextToken()
	// 	p.nowLine = p.lexer.GetLineNumber()
	// }
	pt := p.nextToken
	if p.nextToken.TType != Lexer.EOF {
		p.nextToken = p.lexer.NextToken()
		p.nowLine = p.lexer.GetLineNumber()
		for {
			switch p.nextToken.TType {
			case Lexer.EOLN:
				// 语法分析跳过换行符
				p.nextToken = p.lexer.NextToken()
				p.nowLine = p.lexer.GetLineNumber()
				pt = p.nextToken
				continue
			}
			break
		}
	}
	return pt
}
