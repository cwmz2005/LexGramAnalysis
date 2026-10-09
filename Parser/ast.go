package parser

import (
	"errors"
)

// 抽象语法树节点类型
type NodeType int

const (
	_     NodeType = iota
	BLOCK          // 分程序/函数体/程序
	// DECLIST           // 说明语句表，注意：与直接上级共享SymbolList？还是直接不要List？好吧有点多余了
	VARDEC   // 变量说明
	FUNCDEC  // 函数说明，函数节点Vars
	EXECLIST // 执行语句表
	READ     // 读
	WRITE    // 写
	ASSIGN   // 赋值
	COND     // 条件语句
	BINOP    // 二元运算符
	VA       // 变量
	CONS     // 常数
	FUNCCALL // 函数调用
	ERR      // 错误节点，方便继续解析
)

// 变量类型，mini语言只有整数，所以是变量/形参
type SymbolType int

const (
	VAR  SymbolType = iota //变量
	PARA                   //形参
	FUNC                   // 函数名
)

// 二元运算符类型，与Token一致便于比较
type OpType int

const (
	EQ    OpType = 12 // 12  =
	NE    OpType = 13 // 13  <>
	LE    OpType = 14 // 14  <=
	LT    OpType = 15 // 15  <
	GE    OpType = 16 // 16  >=
	GT    OpType = 17 // 17  >
	MINUS OpType = 18 // 18  -
	MUL   OpType = 19 // 19  *
)

// 变量表项
type SymbolEntry struct {
	SName string
	SType SymbolType
	SProc *ASTNode // 所属过程
	SNode *ASTNode // 对应的声明节点
}

// 抽象语法树节点
type ASTNode struct {
	NType       NodeType
	Parent      *ASTNode
	Name        string       // 有name的才有
	SymbolTable *SymbolList  // 本级符号表
	Op          OpType       // (仅二元运算)运算符类型
	SubNodes    []*ASTNode   // slice本来就是引用类型
	Val         int          // CONS使用
	Sym         *SymbolEntry // （仅声明、变量、函数调用语句）指向它绑定的符号
}

// 变量表，因为又要map快速匹配，又要有序，所以单独开一类型，注意这是同作用域的变量表
type SymbolList struct {
	sym       []*SymbolEntry
	symIndMap map[string]int
}

// 定义新符号，会检查同级重名，不会自动插入ASTNode
func (l *SymbolList) NewSymbol(name string, proc *ASTNode, stype SymbolType, snode *ASTNode) (*SymbolEntry, error) {
	if _, e := l.symIndMap[name]; e {
		return nil, errors.New("符号" + name + "在过程" + proc.Name + "重复定义")
	}
	s := SymbolEntry{name, stype, proc, snode}
	l.sym = append(l.sym, &s)
	l.symIndMap[name] = len(l.sym) - 1
	return &s, nil
}

func MakeSymbolList() (s *SymbolList) {
	s = &SymbolList{
		make([]*SymbolEntry, 0),
		make(map[string]int, 0),
	}
	return
}

// 获取符号对应的entry
func (l *SymbolList) GetSymbol(name string) *SymbolEntry {
	if l == nil {
		return nil
	}
	ind, e := l.symIndMap[name]
	if !e {
		return nil
	}
	return l.sym[ind]
}

// ASTNode逐级向上查找符号
func (n *ASTNode) GetSymbol(name string) (se *SymbolEntry) {
	for se == nil && n != nil {
		se = n.SymbolTable.GetSymbol(name)
		n = n.Parent
	}
	return se
}

// 创建参数较多语法树节点的函数，防止漏掉参数
// 基础，必备参数，不会插入双亲
func makeBase(ntype NodeType, parent *ASTNode) (n *ASTNode) {
	n = &ASTNode{}
	n.NType = ntype
	n.Parent = parent
	n.SubNodes = make([]*ASTNode, 0)
	return
}

// 创建变量节点，不会插入双亲
func makeVar(parent *ASTNode, name string, sym *SymbolEntry) (n *ASTNode) {
	n = makeBase(VA, parent)
	n.Name = name
	n.Sym = sym
	return n
}
