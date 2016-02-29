package main

import (
	"github.com/robertkrimen/otto"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"strings"
)

var (
	codePtr      = kingpin.Arg("code", "code").Required().String()
	lineSeqPtr   = kingpin.Flag("line-seq", "code").Short('l').Default("\n").String()
	columnSeqPtr = kingpin.Flag("column-seq", "code").Short('c').Default(" ").String()
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

func setPrintArrayFunc(vm *otto.Otto) {
	vm.Run(`
	printArray = function(arr) {
		for(i=0;i<arr.length;i+=1){
			console.log(i)
		}
	}
`)

}

func main() {
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

	vm.Run(*codePtr)
}
