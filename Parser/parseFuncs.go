// 递归下降分析，各个元的解析函数
package parser

import (
	"errors"
	"strconv"

	"github.com/cwmz2005/LexGramAnalysis/Lexer"
)

// 当前位置报错
func (p *Parser) errNow(errDetail string) {
	p.Errors = append(p.Errors, ParseError{p.nowLine, errors.New(errDetail)})
}

// 没有可选epsilon时，不回溯读取下一个Token，自动处理错误（此时指针不前进），tdesc为指代该token本应是什么的词汇
func (p *Parser) expectToken(ttype Lexer.TokenType, tdesc string) (t Lexer.Token) {
	for {
		t = p.nextToken
		var errt string
		switch t.TType {
		case ttype:
			p.next()
			return
		case Lexer.ERR:
			errt = t.Sign
		case Lexer.EOLN:
			// 语法分析跳过换行符
			p.next()
			continue
		default:
			errt = "错误的单词“" + t.Sign + "”" + "需要“" + tdesc + "”"
		}
		p.errNow(errt)
		return Lexer.Token{Sign: errt, TType: Lexer.ERR}
	}
}

// 读取下一个已定义变量，返回相应符号表条目
func (p *Parser) getVar(name string, node *ASTNode) (ve *SymbolEntry) {
	ve = node.GetSymbol(name)
	if ve == nil {
		p.errNow("标识符“" + name + "”未定义")
		return nil
	} else if ve.SType == FUNC {
		p.errNow(name + "是函数，此处需要变量")
		return nil
	}
	return ve
}

// 程序/函数体，需要手动插入双亲（程序->分程序省略，没啥意义，直接把程序当作分程序）
func (p *Parser) program(parent *ASTNode, name string) (n *ASTNode) {
	p.expectToken(Lexer.BEGIN, "begin")
	// n = &ASTNode{
	// 	NType:       BLOCK,
	// 	SymbolTable: MakeSymbolList(),
	// 	SubNodes:    make([]*ASTNode, 0),
	// 	Parent:      parent,
	// }
	n = makeBase(BLOCK, parent)
	n.Name = name
	n.SymbolTable = MakeSymbolList()
	p.declList(n)
	p.execList(n) // 执行语句表
	p.expectToken(Lexer.END, "end")
	return n
}

// 说明语句表，需要操作符号表所以直接传入
func (p *Parser) declList(n *ASTNode) {
	p.decl(n)
	p.expectToken(Lexer.SEMICOLON, ";")
	for p.nextToken.TType == Lexer.INTEGER {
		// LL1，为integer选择继续声明语句
		p.decl(n)
		p.expectToken(Lexer.SEMICOLON, ";")
	}
	// p.expectToken(Lexer.BEGIN, "begin")
}

// 说明语句，需要排查操作SymbolList所以传入n
func (p *Parser) decl(n *ASTNode) {
	p.expectToken(Lexer.INTEGER, "integer")
	switch p.nextToken.TType {
	case Lexer.IDENTIFIER:
		// 变量声明
		t := p.next()
		ent, err := n.SymbolTable.NewSymbol(t.Sign, n, VAR, nil)
		if err != nil {
			// 大概率是重名
			p.Errors = append(p.Errors, ParseError{p.nowLine, err})
		} else {
			vn := &ASTNode{
				Name:   t.Sign,
				Sym:    ent,
				NType:  VARDEC,
				Parent: n,
			}
			ent.SNode = vn
			n.SubNodes = append(n.SubNodes, vn)
		}
	case Lexer.FUNCTION:
		// 函数声明
		p.next()
		t := p.expectToken(Lexer.IDENTIFIER, "函数名称标识符")
		if t.TType == Lexer.IDENTIFIER {
			ent, err := n.SymbolTable.NewSymbol(t.Sign, n, FUNC, nil)
			var fn *ASTNode
			if err != nil {
				// 大概率是重名
				p.Errors = append(p.Errors, ParseError{p.nowLine, err})
			} else {
				fn = &ASTNode{
					Name:        t.Sign,
					Sym:         ent,
					NType:       FUNCDEC,
					Parent:      n,
					SubNodes:    make([]*ASTNode, 0),
					SymbolTable: MakeSymbolList(),
				}
				ent.SNode = fn
				n.SubNodes = append(n.SubNodes, fn)
				p.expectToken(Lexer.LPAREN, "(")
				tpara := p.expectToken(Lexer.IDENTIFIER, "参数（变量）")
				if tpara.TType == Lexer.IDENTIFIER {
					// 处理参数
					par, pare := fn.SymbolTable.NewSymbol(tpara.Sign, fn, PARA, nil)
					if pare != nil {
						panic("不可能的变量重名：" + pare.Error())
					}
					paran := &ASTNode{
						Name:  tpara.Sign,
						Sym:   par,
						NType: VARDEC,
					}
					fn.SubNodes = append(fn.SubNodes, paran)
					par.SNode = paran
				}
				p.expectToken(Lexer.RPAREN, ")")
				p.expectToken(Lexer.SEMICOLON, ";")
				fn.SubNodes = append(fn.SubNodes, p.program(fn, t.Sign))
			}
		}
	default:
		// 错误
		p.errNow("声明语句不正确，需要跟标识符或者function")
	}
	// p.expectToken(Lexer.BEGIN, "begin")
}

// 执行语句表
func (p *Parser) execList(n *ASTNode) {
	en := makeBase(EXECLIST, n)
	p.exec(en)
	for p.nextToken.TType == Lexer.SEMICOLON {
		// LL1，执行语句表不是分号是结尾
		p.expectToken(Lexer.SEMICOLON, ";")
		p.exec(en)
	}
	n.SubNodes = append(n.SubNodes, en)

}
func (p *Parser) exec(n *ASTNode) {
	switch p.nextToken.TType {
	case Lexer.READ:
		// 读语句
		rn := makeBase(READ, n)
		p.expectToken(Lexer.READ, "read")
		p.expectToken(Lexer.LPAREN, "(")
		vt := p.expectToken(Lexer.IDENTIFIER, "读的变量名")
		if vt.TType == Lexer.IDENTIFIER {
			ve := p.getVar(vt.Sign, n)
			if ve != nil {
				rn.SubNodes = append(rn.SubNodes, makeVar(rn, vt.Sign, ve))
			}
		}
		n.SubNodes = append(n.SubNodes, rn)
		p.expectToken(Lexer.RPAREN, ")")
	case Lexer.WRITE:
		// 写语句
		rn := makeBase(WRITE, n)
		p.expectToken(Lexer.WRITE, "write")
		p.expectToken(Lexer.LPAREN, "(")
		vt := p.expectToken(Lexer.IDENTIFIER, "写的变量名")
		if vt.TType == Lexer.IDENTIFIER {
			ve := p.getVar(vt.Sign, n)
			if ve != nil {
				rn.SubNodes = append(rn.SubNodes, makeVar(rn, vt.Sign, ve))
			}
		}
		n.SubNodes = append(n.SubNodes, rn)
		p.expectToken(Lexer.RPAREN, ")")
	case Lexer.IDENTIFIER:
		// 赋值语句
		vt := p.expectToken(Lexer.IDENTIFIER, "赋值变量名")
		if vt.TType == Lexer.IDENTIFIER {
			ve := p.getVar(vt.Sign, n)
			if ve != nil {
				// 标识符定义了才视作赋值处理
				in := makeBase(ASSIGN, n)
				vn := makeVar(in, vt.Sign, ve)
				in.SubNodes = append(in.SubNodes, vn)
				p.expectToken(Lexer.ASSIGN, ":=")
				p.algExp(in)
				n.SubNodes = append(n.SubNodes, in)
			}
		}
	case Lexer.IF:
		// 条件语句
		p.expectToken(Lexer.IF, "if")
		in := makeBase(COND, n)
		n.SubNodes = append(n.SubNodes, in)
		p.condExp(in)
		p.expectToken(Lexer.THEN, "then")
		p.exec(in)
		p.expectToken(Lexer.ELSE, "else")
		p.exec(in)
	default:
		// 出错
		n.SubNodes = append(n.SubNodes, makeBase(ERR, n))
		p.errNow("错误的Token" + p.nextToken.Sign + "需要执行语句")
	}
}

// 条件表达式
func (p *Parser) condExp(n *ASTNode) {
	cn := makeBase(BINOP, n)
	n.SubNodes = append(n.SubNodes, cn)
	p.algExp(cn)
	switch p.nextToken.TType {
	case isCondOp(p.nextToken.TType):
		op := p.next()
		cn.Op = OpType(op.TType)
	default:
		p.errNow("需要条件运算符")
	}
	p.algExp(cn)
}

// 是否为条件运算符
func isCondOp(tp Lexer.TokenType) (r Lexer.TokenType) {
	r = Lexer.ERR
	switch tp {
	case Lexer.LE, Lexer.LT, Lexer.GT, Lexer.GE, Lexer.EQ, Lexer.NE: /*关系运算符*/
		return tp
	default:
		return
	}
}

// 是否为算术表达式 FOLLOW集合成员，如果是则返回本身（方便case比较），不是则返回ERR
func isAlgFollow(tp Lexer.TokenType) (r Lexer.TokenType) {
	r = Lexer.ERR
	switch tp {
	case isCondOp(tp), /*关系运算符*/
		Lexer.RPAREN, /*函数调用*/
		Lexer.THEN,   /*条件表达式Follow*/
		/*赋值语句->执行语句Follow*/
		Lexer.ELSE, Lexer.SEMICOLON, Lexer.END:
		return tp
	default:
		return
	}
}

// 算术表达式
func (p *Parser) algExp(n *ASTNode) {
	exp := p.algItem(n)
	switch p.nextToken.TType {
	case Lexer.MINUS:
		for p.nextToken.TType == Lexer.MINUS {
			p.expectToken(Lexer.MINUS, "-")
			bio := makeBase(BINOP, n)
			bio.Op = MINUS
			exp.Parent = bio
			rightItem := p.algItem(bio)
			bio.SubNodes = append(bio.SubNodes, exp, rightItem) // 原始节点变左孩子
			exp = bio
		}
		n.SubNodes = append(n.SubNodes, exp)
		return
	case isAlgFollow(p.nextToken.TType): // Follow集return
		n.SubNodes = append(n.SubNodes, exp)
		return
	default:
		n.SubNodes = append(n.SubNodes, exp)
		p.errNow("算术表达式后不能跟" + p.nextToken.Sign)
		//出错
	}
}

// 项 上层有可能直接返回或新建二元操作节点并作为操作数，所以不会自动插入n，注意！需要手动更新Parent
// parent变成了非原子操作！ntr危险！指定了孩子的parent但不一定真插入该parent
func (p *Parser) algItem(n *ASTNode) (item *ASTNode) {
	item = p.algFac(n)
	switch p.nextToken.TType {
	case Lexer.MUL:
		for p.nextToken.TType == Lexer.MUL {
			p.expectToken(Lexer.MUL, "*")
			bio := makeBase(BINOP, n)
			bio.Op = MUL
			item.Parent = bio
			rightItem := p.algFac(bio)
			bio.SubNodes = append(bio.SubNodes, item, rightItem) // 原始节点变左孩子
		}
		return
	// FOLLOW集合，表示结束，直接返回常数
	case isAlgFollow(p.nextToken.TType), Lexer.MINUS:
		return
	default:
		p.errNow("项后不能跟" + p.nextToken.Sign)
		return
	}
}

// 因子 同理不自动插入节点 注意！如果不是直接插入n需要手动更新Parent
func (p *Parser) algFac(n *ASTNode) (fac *ASTNode) {
	switch p.nextToken.TType {
	case Lexer.IDENTIFIER:
		vt := p.expectToken(Lexer.IDENTIFIER, "标识符名")
		switch p.nextToken.TType {
		case Lexer.MUL, /*后因子*/ // 变量直接结尾，Follow集即可判定为变量
			Lexer.MINUS, /*项FOLLOW*/
			/*算术表达式follow*/
			isAlgFollow(p.nextToken.TType):
			// 变量
			ve := p.getVar(vt.Sign, n)
			if ve != nil {
				return makeVar(n, ve.SName, ve)
			}
		case Lexer.LPAREN:
			// 函数
			p.expectToken(Lexer.LPAREN, "(")
			se := n.GetSymbol(vt.Sign)
			if se != nil {
				if se.SType != FUNC {
					p.errNow("标识符" + se.SName + "不是函数")
				} else {
					fac = makeBase(FUNCCALL, n)
					fac.Name = vt.Sign
					fac.Sym = se
					return
				}
			} else {
				p.errNow("标识符" + vt.Sign + "未定义")
			}
		default:
			// 出错
			p.errNow(p.nextToken.Sign + "不是合法的因子！")
		}
	case Lexer.CONSTANT:
		// 常数
		v := p.expectToken(Lexer.CONSTANT, "常数")
		fac = makeBase(CONS, n)
		fac.Val, _ = strconv.Atoi(v.Sign) // 不可能出错，CONSTANT确保是数字
		return fac
	default:
		p.errNow(p.nextToken.Sign + "不是合法的因子！")
	}
	return makeBase(ERR, n)
}
