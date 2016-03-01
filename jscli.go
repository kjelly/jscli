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

var builtinFuncCode = []string{
	`
printArray = function(arr) {
	for(i=0;i<arr.length;i+=1){
		console.log(i)
	}
}

`,
	`
function isFunc(val) {
	if(typeof val === 'function') {
		return true;
	}else{
		return false;
	}
}
`,
	`
mapIf = function(arr, func, preFunc, postFunc) {
	ret = []
	for(i=0;i<arr.length;i+=1) {
		val = arr[i];
		if(isFunc(preFunc) && !preFunc(val)){
			continue;
		}
		out = func(val);
		if(isFunc(postFunc) && !postFunc(out)){
			continue;
		}
		ret.push(val);
	}
	return ret;
}

`,
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
	_, err = vm.Run(string(bytes))
	if err != nil {
		panic(err)
	}
}

type Matrix [][]string

func setBuiltinFunc(vm *otto.Otto) {
	for _, ele := range builtinFuncCode {
		_, err := vm.Run(ele)
		if err != nil {
			panic(err)
		}
	}

	vm.Set("exec", func(call otto.FunctionCall) otto.Value {
		args := make([]string, len(call.ArgumentList))
		cmdPath, err := call.Argument(0).ToString()
		if err != nil {
			panic(err)
		}
		for i := 1; i < len(call.ArgumentList); i += 1 {
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

	vm.Set("execStdin", func(call otto.FunctionCall) otto.Value {
		args := make([]string, len(call.ArgumentList))
		cmdPath, err := call.Argument(0).ToString()
		if err != nil {
			panic(err)
		}
		stdin, err := call.Argument(1).ToString()
		if err != nil {
			panic(err)
		}
		for i := 2; i < len(call.ArgumentList); i += 1 {
			ret, err := call.Argument(i).ToString()
			if err == nil {
				args[i] = ret
			} else {
				panic(err)
			}

		}
		output := callExternalFuncWithStdin(cmdPath, stdin, args)
		result, err := vm.ToValue(output)
		if err != nil {
			panic(err)
		}

		return result
	})
	vm.Set("print", func(call otto.FunctionCall) otto.Value {
		for i := 0; i < len(call.ArgumentList); i += 1 {
			str, err := call.Argument(i).ToString()
			if err == nil {
				fmt.Print(str)
			} else {
				panic(err)
			}
		}
		val, _ := vm.ToValue("")
		return val
	})

	sprintf := func(call otto.FunctionCall) string {
		args := make([]interface{}, len(call.ArgumentList)-1)
		format, err := call.Argument(0).ToString()
		if err != nil {
			panic(err)
		}
		for i := 1; i < len(call.ArgumentList); i += 1 {
			str, err := call.Argument(i).ToString()
			if err == nil {
				args[i-1] = str
			} else {
				panic(err)
			}
		}
		return fmt.Sprintf(format, args...)
	}

	vm.Set("sprint", func(call otto.FunctionCall) otto.Value {
		val := sprintf(call)
		ret, _ := vm.ToValue(val)
		return ret
	})
	vm.Set("printf", func(call otto.FunctionCall) otto.Value {
		val := sprintf(call)
		fmt.Print(val)
		ret, _ := vm.ToValue(val)
		return ret
	})
}

func callExternalFunc(cmd string, args []string) string {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		panic(err)
	}
	return string(out)
}

func callExternalFuncWithStdin(cmd string, stdin string, args []string) string {
	p := exec.Command(cmd, args...)
	p.Stdin = strings.NewReader(stdin)
	out, err := p.Output()
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
	vm.Run("println = console.log")
	setBuiltinFunc(vm)

	for i := 0; i < len(*funcListPtr); i += 1 {
		initExternelFunc(vm, (*funcListPtr)[i])
	}

	for i := 0; i < len(*pathListPtr); i += 1 {
		addPATHEnv((*pathListPtr)[i])
	}

	for i := 0; i < len(*jsListPtr); i += 1 {
		readJSFile(vm, (*jsListPtr)[i])
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
