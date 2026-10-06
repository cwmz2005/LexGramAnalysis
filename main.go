package main

import (
	"fmt"
	"os"

	"github.com/cwmz2005/LexGramAnalysis/Lexer"
)

func main() {
	// 词法分析测试
	if len(os.Args) <= 1 {
		fmt.Println("请将mini源代码文件作为第一个参数传入")
		return
	}
	l, e := Lexer.MakeLexer(os.Args[1])
	if e != nil {
		fmt.Println("打开文件出错：", e)
		return
	}
	defer l.Close()
	for {
		nt, err := l.NextToken()
		if err != nil {
			fmt.Println("下一个Token出错：", err)
			return
		}
		if nt.TType == Lexer.ERR {
			fmt.Printf("第%d行：不正确的符号“%s”\n", l.GetLineNumber(), nt.Sign)
			continue
		}
		fmt.Printf("读取Token：“%s”，种别：%d\n", nt.Sign, nt.TType)
		if nt.TType == Lexer.EOF {
			return
		}
	}
}
