package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/Astemirdum/si/internal/compiler"
)

func main() {
	//filePath := flag.String("filepath", "", "Path to the file")
	//flag.Parse()
	if len(os.Args) < 2 {
		log.Fatal("No file path provided")
	}
	filePath := os.Args[1]
	fmt.Println("File path:", filePath)

	if err := run(filePath); err != nil {
		log.Fatal(err)
	}
}

func run(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	sb := &strings.Builder{}
	if _, err = io.Copy(sb, file); err != nil {
		return err
	}
	src := sb.String()

	c := compiler.NewCompiler()
	defer c.Destroy()
	var opts []compiler.Option
	//opts = append(opts, compiler.DeclareMalloc())
	result, err := c.RunProgramSi(src, opts...)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}
