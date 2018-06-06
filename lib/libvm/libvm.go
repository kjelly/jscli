package libvm

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var builtinFuncCode = []string{
	`
printArray = function(arr) {
	for(i=0;i<arr.length;i+=1){
		console.log(arr[i])
	}
}
printA=printArray;
printL=printArray;
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
	`
printColumn = function(arr, column, func) {
	if(isFunc(func)) printFunc=func;
	else printFunc=println;
	arr.map(function(line, index, x){
		printFunc(line[column]);
	});
}
printC=printColumn;
`,
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

func SetBuiltinFunc(vm *otto.Otto) {
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
		for i := 1; i < len(call.ArgumentList); i++ {
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
		for i := 1; i < len(call.ArgumentList); i++ {
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

func InitExternelFunc(vm *otto.Otto, cmdPath string) {
	funcName := cmdPath[:len(cmdPath)-len(filepath.Ext(cmdPath))]
	vm.Set(funcName, func(call otto.FunctionCall) otto.Value {
		args := make([]string, len(call.ArgumentList))
		for i := 0; i < len(call.ArgumentList); i++ {
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

func ReadJSFile(vm *otto.Otto, path string) {
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
