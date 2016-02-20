package main

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Reads all .json files in the current folder
// and encodes them as strings literals in textfiles.go
func main() {
	fs, _ := ioutil.ReadDir(".")
	out, _ := os.Create("swagger.pb.go")
	out.Write([]byte("package echopb \n\nconst (\n"))
	for _, f := range fs {
		if strings.HasSuffix(f.Name(), ".json") {
			name := strings.TrimPrefix(f.Name(), "service.")
			out.Write([]byte(strings.TrimSuffix(name, ".json") + " = `"))
			f, _ := os.Open(f.Name())
			io.Copy(out, f)
			out.Write([]byte("`\n"))
		}
	}
	out.Write([]byte(")\n"))
}
