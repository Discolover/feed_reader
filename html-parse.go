package demo

import (
	"fmt"
	"golang.org/x/net/html"
	"os"
	"io"
)

func tokenizeHTML(r io.Reader) error {
	z := html.NewTokenizer(r)

	for {
		if z.Next() == html.ErrorToken {
			// Returning io.EOF indicates success.
			return z.Err()
		}
		fmt.Println(z.Token().Type)
	}
}

func main() {
	f, err := os.Open("sample.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// bytes, err := io.ReadAll(f)
	// if err != nil {
	// 	panic(err)
	// }
	err = tokenizeHTML(f)
	if err != io.EOF {
		panic(err)
	}
}