package main

import (
	"fmt"
	"os"

	"github.com/otiai10/gosseract/v2"
)

func main() {
	client := gosseract.NewClient()
	defer client.Close()
	client.SetImage(os.Args[1])
	text, _ := client.Text()
	fmt.Println(text)
	// Hello, World!
}
