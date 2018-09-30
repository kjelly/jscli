package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/kjelly/jscli/lib/libvm"
	"github.com/kjelly/jscli/lib/utils"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
)

func readAll() string {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err == nil {
		return string(bytes)
	}
	panic("Failed to read stdin")
}

// Matrix matrix
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
		CodeList  []string `arg:"positional" help:"javascript code"`
		LineSeq   string   `arg:"-l"`
		ColumnSeq string   `arg:"-c"`
		FuncAlias string   `arg:"-a" help:"The string used for being replaced with function"`
		FuncList  []string `arg:"-f" help:"import external command as buildin function."`
		JSList    []string `arg:"-j" help:"the js files to import"`
		PathList  []string `arg:"-p" help:"the path for searching execute."`
		Reverse   bool     `arg:"-r" help:"execute code from right to left."`
		NoStdin   bool     `arg:"--nostdin,-s" help:"Don't read from stdin."`
	}
	args.LineSeq = "\n"
	args.FuncAlias = "func"
	args.ColumnSeq = " +"
	arg.MustParse(&args)

	var stdin string

	if args.NoStdin {
		stdin = ""
	} else {
		stdin = readAll()
	}

	stdin = strings.TrimSpace(stdin)

	lines, matrixPtr := utils.ParseStdin(stdin, args.ColumnSeq, args.LineSeq)

	vm := otto.New()

	vm.Set("stdin", stdin)
	vm.Set("stdout", "")
	vm.Set("lines", lines)
	vm.Set("matrix", *matrixPtr)
	vm.Run("println = console.log")
	libvm.SetBuiltinFunc(vm)

	for i := 0; i < len(args.FuncList); i++ {
		libvm.InitExternelFunc(vm, (args.FuncList)[i])
		fmt.Printf("%s\n", args.FuncList[i])
	}

	for i := 0; i < len(args.PathList); i++ {
		addPATHEnv((args.PathList)[i])
	}

	for i := 0; i < len(args.JSList); i++ {
		libvm.ReadJSFile(vm, (args.JSList)[i])
	}

	codeList := sort.StringSlice(args.CodeList)

	if args.Reverse {
		for i, j := 0, len(codeList)-1; i < j; i, j = i+1, j-1 {
			codeList[i], codeList[j] = codeList[j], codeList[i]
		}
	}

	re1 := regexp.MustCompile(args.FuncAlias + " ")
	re2 := regexp.MustCompile(args.FuncAlias + "\\(")

	var err error
	var out otto.Value
	for i := 0; i < len(codeList); i++ {
		code := re1.ReplaceAllString(codeList[i], "function ")
		code = re2.ReplaceAllString(code, "function(")
		out, err = vm.Run(code)
		if err != nil {
			panic(err)
		}
	}
	if !out.IsUndefined() && !out.IsNull() {
		outString, _ := out.ToString()
		fmt.Printf("%s\n", outString)
	}

}
