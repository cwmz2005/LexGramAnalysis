// 词法分析模块，一次只读一个ASCII字符
package Lexer

import (
	"io"
	"os"
)

type TokenType int

type Token struct {
	Sign  string    // 单词符号
	TType TokenType // 种别
}

const (
	_          TokenType = iota
	BEGIN                // 1
	END                  // 2
	INTEGER              // 3
	IF                   // 4
	THEN                 // 5
	ELSE                 // 6
	FUNCTION             // 7
	READ                 // 8
	WRITE                // 9
	IDENTIFIER           // 10
	CONSTANT             // 11
	EQ                   // 12  =
	NE                   // 13  <>
	LE                   // 14  <=
	LT                   // 15  <
	GE                   // 16  >=
	GT                   // 17  >
	MINUS                // 18  -
	MUL                  // 19  *
	ASSIGN               // 20  :=
	LPAREN               // 21  (
	RPAREN               // 22  )
	SEMICOLON            // 23  ;
	EOLN                 // 24
	EOF                  // 25
	ERR                  // 出错
)

var stMap = map[string]TokenType{
	"begin":    1,
	"end":      2,
	"integer":  3,
	"if":       4,
	"then":     5,
	"else":     6,
	"function": 7,
	"read":     8,
	"write":    9,
	// "标识符":10
	// "常数":11
	"=":  12,
	"<>": 13,
	"<=": 14,
	"<":  15,
	">=": 16,
	">":  17,
	"-":  18,
	"*":  19,
	":=": 20,
	"(":  21,
	")":  22,
	";":  23,
	// "\n": 24,
	// 程序每行后加：“EOLN[空格]24”
	// 测试程序结尾加：“EOF[空格]25”
}

//	type Lexer interface {
//		GetFileName() string
//		NextToken() Token
//		Close()
//	}
//
// 词法分析器
type Lexer struct {
	f *os.File
	// 从0行开始
	nowLineIndex int
}

func isDigit[T rune | byte](r T) bool {
	return r >= '0' && r <= '9'
}
func isLetter[T rune | byte](r T) bool {
	// 仅支持小写字母
	return r >= 'a' && r <= 'z' /*|| r >= 'A' && r <= 'Z'*/
	// 不支持大写字母就一个正确例程都没了？？？但是实验指导书确实是这么说的

}

// 是否为可以忽略的空白符
func isEmpty[T rune | byte](r T) bool {
	return r == ' ' || r == '\t' || r == '\r'
}

// 是否为单列一个Token的特殊字符，是否为可能两个连起来作为一个Token的第一个特殊字符，是否为...的第二个特殊字符
func isSpecial[T rune | byte](r T) (special bool, connectable bool, connectEnd bool) {
	switch r {
	case '<', ':':
		return true, true, false
	case '=':
		return true, false, true
	case '>':
		return true, true, true
	case '(', ')', '*', '-', ';':
		return true, false, false
	}
	return false, false, false
}

// 判断是否为合法的，两个字符组成的运算符，如果是则输出Token
func isLegalDS(word string) (bool, TokenType) {
	if len(word) != 2 {
		return false, ERR
	}
	switch word {
	case "<>":
		return true, NE
	case ">=":
		return true, GE
	case "<=":
		return true, LE
	case ":=":
		return true, EQ
	}
	return false, ERR
}

// 判断是否该停止，不判断正确性
func shouldStop(bf string, nb byte) (stop bool, rewind bool) {
	// 下一个是可以无视类型的空字符（直接空字符时不会走这个分支）
	if isEmpty(nb) {
		return true, false
	}
	if nb == '\n' {
		if len(bf) == 0 {
			return false, false
		}
		return true, true
	}

	// 下一个字符是特殊字符
	s, _, cb := isSpecial(nb)
	if s {
		if len(bf) == 0 {
			return false, false
		}
		if cb && len(bf) == 1 {
			// 可能的二元运算符则不停，读下一个
			t := bf + string(nb)
			if b, _ := isLegalDS(t); b {
				return false, false
			}
		}
		return true, true
	}

	// 下一个字符是数字或者字母
	if isDigit(nb) || isLetter(nb) {
		// 只有上一个还是数字或字母时或者开头时继续
		li := len(bf) - 1
		if li < 0 || isDigit(bf[li]) || isLetter(bf[li]) {
			return false, false
		}
		return true, true
	}
	// 兜底，非法字符
	if len(bf) > 0 {
		return true, true // 回退，留给下一次处理
	}
	// 兜底，非法字符停止
	return true, false
}

// 查是否为合法标识符，不查是否为保留字
func isLegalIdentifier(word string) bool {
	if len(word) == 0 {
		return false
	}
	if !isLetter(word[0]) {
		// 字母开头
		return false
	}
	for _, l := range word {
		if !isDigit(l) && !isLetter(l) {
			// 必须是数字或小写字母组成
			return false
		}
	}
	return true
}

// 是否为合法常数（只支持无符号整数）
func isConstant(word string) bool {
	if len(word) == 0 {
		return false
	}
	for _, w := range word {
		if !isDigit(w) {
			return false
		}
	}
	return true
}

// 关闭
func (l Lexer) Close() {
	l.f.Close()
}

//	func handleWord(word string) Token {
//		// 处理词
//		if m := stMap[word]; m != 0 {
//			return Token{word, m}
//		}
//		if
//		return Token{word, ERR}
//	}
//
// 获取正在分析的文件名
func (l Lexer) GetFileName() (n string) {
	return l.f.Name()
}

// 获取行号，从第一行开始
func (l Lexer) GetLineNumber() int {
	return l.nowLineIndex + 1
}
func (l *Lexer) NextToken() (t Token) { // 当心！值接收者更改行号后原对象不变！！！默认全传拷贝
	buffer := ""
	nextByte := make([]byte, 1)
	for {
		_, rerr := l.f.Read(nextByte) // 一次只读一个byte，没读到必是出错
		if rerr != nil && rerr != io.EOF {
			return Token{"读取出错：" + rerr.Error(), ERR}
		}
		var stop, rewind bool
		if rerr == io.EOF {
			// 合法的出错
			if len(buffer) == 0 {
				return Token{"EOF", EOF}
			}
			stop, rewind = true, false // 下一个是最后一个字符，可以无限读下去还是EOF错误，不用回溯一位
		} else {
			if len(buffer) == 0 && isEmpty(nextByte[0]) {
				// 跳过开头所有空字符
				continue
			}
			stop, rewind = shouldStop(buffer, nextByte[0])
		}

		if stop {
			if rewind {
				l.f.Seek(-1, io.SeekCurrent)
			}
			if len(buffer) == 0 {
				// 下一个字符完全非法
				return Token{string(nextByte[0]), ERR}
			}
			break
		}

		buffer += string(nextByte)
	}
	// buffer现在是一个词
	if ti := stMap[buffer]; ti != 0 {
		// 保留字
		return Token{buffer, ti}
	}
	if isConstant(buffer) {
		return Token{buffer, CONSTANT}
	}
	if isLegalIdentifier(buffer) {
		return Token{buffer, IDENTIFIER}
	}
	if b, t := isLegalDS(buffer); b {
		return Token{buffer, t}
	}
	if buffer == "\n" {
		l.nowLineIndex += 1
		return Token{"EOLN", EOLN}
	}
	return Token{"不合法的符号" + buffer + string(nextByte[0]), ERR}
}

func MakeLexer(fn string) (l *Lexer, err error) {
	f, err := os.Open(fn)
	return &Lexer{f, 0}, err
}
