package main

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var (
	codeListPtr  = ArgsStrList(kingpin.Arg("code", "the js code you want to run").Required())
	lineSeqPtr   = kingpin.Flag("line-seq", "the char used for split line").Short('l').Default("\n").String()
	columnSeqPtr = kingpin.Flag("column-seq", "the char used for split column").Short('c').Default(" ").String()
	funcListPtr  = ArgsStrList(kingpin.Flag("funcion", "function").Short('f'))
	pathListPtr  = ArgsStrList(kingpin.Flag("path", "command search path").Short('p'))
	jsListPtr    = ArgsStrList(kingpin.Flag("js", "Javascript file").Short('j'))
	reversePtr   = kingpin.Flag("reverse", "execute code with reverse order").Short('r').Bool()
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

func readJSFile(vm *otto.Otto, path string) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	vm.Run(string(bytes))
}

type Matrix [][]string

func setPrintArrayFunc(vm *otto.Otto) {
	vm.Run(`
	printArray = function(arr) {
		for(i=0;i<arr.length;i+=1){
			console.log(i)
		}
	}`)

}

func callExternalFunc(cmd string, args []string) string {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		panic(err)
	}
	return string(out)

}

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

func initExternelFunc(vm *otto.Otto, cmdPath string) {
	funcName := cmdPath[:len(cmdPath)-len(filepath.Ext(cmdPath))]
	vm.Set(funcName, func(call otto.FunctionCall) otto.Value {
		args := make([]string, len(call.ArgumentList))
		for i := 0; i < len(call.ArgumentList); i += 1 {
			ret, err := call.Argument(i).ToString()
			if err == nil {
				args[i] = ret
			} else {
				panic(err)
			}

		}
		output := callExternalFunc(cmdPath, args)
		result, err := vm.ToValue(output)
		if err != nil {
			panic(err)
		}

		return result
	})

}

func main() {
	addPATHEnv(getWd())
	fmt.Printf("")
	kingpin.Parse()

	stdin := readAll()
	lines := strings.Split(stdin, *lineSeqPtr)
	matrixPtr := new(Matrix)

	for i := 0; i < len(lines); i += 1 {
		*matrixPtr = append(*matrixPtr, strings.Split(lines[i], *columnSeqPtr))
	}

	vm := otto.New()

	vm.Set("stdin", stdin)
	vm.Set("stdout", "")
	vm.Set("lines", lines)
	vm.Set("matrix", *matrixPtr)
	vm.Run("print = console.log")
	setPrintArrayFunc(vm)

	for i := 0; i < len(*funcListPtr); i += 1 {
		initExternelFunc(vm, (*funcListPtr)[i])
	}

	for i := 0; i < len(*pathListPtr); i += 1 {
		addPATHEnv((*pathListPtr)[i])
	}

	for i := 0; i < len(*jsListPtr); i += 1 {
		fmt.Printf("%v", i)
		readJSFile(vm, (*jsListPtr)[i])
	}

	codeList := sort.StringSlice(*codeListPtr)

	if *reversePtr {
		for i, j := 0, len(codeList)-1; i < j; i, j = i+1, j-1 {
			codeList[i], codeList[j] = codeList[j], codeList[i]
		}
	}

	fmt.Printf("%v", codeList)

	for i := 0; i < len(codeList); i += 1 {
		vm.Run(codeList[i])
	}

}
