package main

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"github.com/ya790206/jscli/lib/libvm"
	"github.com/ya790206/jscli/lib/utils"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"sort"
)

var (
	codeListPtr  = ArgsStrList(kingpin.Arg("code", "the js code you want to run").Required())
	lineSeqPtr   = kingpin.Flag("line-seq", "the char used for split line").Short('l').Default("\n").String()
	columnSeqPtr = kingpin.Flag("column-seq", "the char used for split column").Short('c').Default(" +").String()
	funcListPtr  = ArgsStrList(kingpin.Flag("funcion", "function").Short('f'))
	pathListPtr  = ArgsStrList(kingpin.Flag("path", "command search path").Short('p'))
	jsListPtr    = ArgsStrList(kingpin.Flag("js", "Javascript file").Short('j'))
	reversePtr   = kingpin.Flag("reverse", "execute code with reverse order").Short('r').Bool()
	nostdinPtr   = kingpin.Flag("nostdin", "Dont't read from stdin").Bool()
)

type argsStrList []string

func (i *argsStrList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *argsStrList) String() string {
	return ""
}

func (i *argsStrList) IsCumulative() bool {
	return true
}

func ArgsStrList(s kingpin.Settings) (target *[]string) {
	target = new([]string)
	s.SetValue((*argsStrList)(target))
	return
}

func readAll() string {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err == nil {
		return string(bytes)
	} else {
		panic("Failed to read stdin")
	}
}

type Matrix [][]string

func addPATHEnv(path string) {
	oldEnv := os.Getenv("PATH")
	newEnv := path + ":" + oldEnv
	os.Setenv("PATH", newEnv)
}

func getWd() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd
}

func main() {
	addPATHEnv(getWd())
	fmt.Printf("")
	kingpin.Parse()

	var stdin string

	if *nostdinPtr {
		stdin = ""
	} else {
		stdin = readAll()
	}

	lines, matrixPtr := utils.ParseStdin(stdin, *columnSeqPtr, *lineSeqPtr)

	vm := otto.New()

	vm.Set("stdin", stdin)
	vm.Set("stdout", "")
	vm.Set("lines", lines)
	vm.Set("matrix", *matrixPtr)
	vm.Run("println = console.log")
	libvm.SetBuiltinFunc(vm)

	for i := 0; i < len(*funcListPtr); i += 1 {
		libvm.InitExternelFunc(vm, (*funcListPtr)[i])
	}

	for i := 0; i < len(*pathListPtr); i += 1 {
		addPATHEnv((*pathListPtr)[i])
	}

	for i := 0; i < len(*jsListPtr); i += 1 {
		libvm.ReadJSFile(vm, (*jsListPtr)[i])
	}

	codeList := sort.StringSlice(*codeListPtr)

	if *reversePtr {
		for i, j := 0, len(codeList)-1; i < j; i, j = i+1, j-1 {
			codeList[i], codeList[j] = codeList[j], codeList[i]
		}
	}

	for i := 0; i < len(codeList); i += 1 {
		_, err := vm.Run(codeList[i])
		if err != nil {
			panic(err)
		}
	}

}
