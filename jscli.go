package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/kjelly/jscli/lib/libvm"
	"github.com/kjelly/jscli/lib/utils"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"os"
	"sort"
)

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

	var args struct {
		CodeList  []string `arg:"positional"`
		LineSeq   string   `arg:"-l"`
		ColumnSeq string   `arg:"-c"`
		FuncList  []string `arg:"-f"`
		JSList    []string `arg:"-j"`
		PathList  []string `arg:"-p"`
		Reverse   bool     `arg:"-r"`
		NoStdin   bool     `arg:"--nostdin"`
	}
	args.LineSeq = "\n"
	args.ColumnSeq = " +"
	arg.MustParse(&args)

	var stdin string

	if args.NoStdin {
		stdin = ""
	} else {
		stdin = readAll()
	}

	lines, matrixPtr := utils.ParseStdin(stdin, args.ColumnSeq, args.LineSeq)

	vm := otto.New()

	vm.Set("stdin", stdin)
	vm.Set("stdout", "")
	vm.Set("lines", lines)
	vm.Set("matrix", *matrixPtr)
	vm.Run("println = console.log")
	libvm.SetBuiltinFunc(vm)

	for i := 0; i < len(args.FuncList); i += 1 {
		libvm.InitExternelFunc(vm, (args.FuncList)[i])
	}

	for i := 0; i < len(args.PathList); i += 1 {
		addPATHEnv((args.PathList)[i])
	}

	for i := 0; i < len(args.JSList); i += 1 {
		libvm.ReadJSFile(vm, (args.JSList)[i])
	}

	codeList := sort.StringSlice(args.CodeList)

	if args.Reverse {
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
