package main

import (
	"io"
	"io/ioutil"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	SlowSearch(out)
}

func main() {
	//buf := bytes.Buffer{}
	SlowSearch(ioutil.Discard)
	//str := string(buf.Bytes())
	//log.Println(str)
}
