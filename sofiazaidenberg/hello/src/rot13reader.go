package main

import (
		"fmt"
	"io"
	//	"math"
	"os"
	"strings"
	"bytes"
)

type rot13Reader struct {
	r io.Reader
}

func (z rot13Reader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	n, err = z.r.Read(p)

	if err != nil {
		return n, err
	}

	for l := range p[0:n] {
		if p[l] >= 'a' && p[l] <= 'z' {
			if p[l]+13 > 'z' {
				p[l] = 12 + 'a' - ('z' - p[l])
			} else {
				p[l] += 13
			}
		} else if p[l] >= 'A' && p[l] <= 'Z' {
			if p[l]+13 > 'Z' {
				p[l] = 12 + 'A' - ('Z' - p[l])
			} else {
				p[l] += 13
			}
		}
	}

	return n, err
}

func main() {
	s := strings.NewReader("Lbh penpxrq gur pbqr!")
	r := rot13Reader{s}
//	io.Copy(os.Stdout, &r)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	s1 := buf.String()
	fmt.Println(s1)
	r = rot13Reader{strings.NewReader(s1)}
	io.Copy(os.Stdout, &r)
}
