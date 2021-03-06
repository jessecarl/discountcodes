package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"strconv"
)

func main() {
	count := flag.Uint("count", 100, "number of codes to generate")
	page := flag.Uint("page", 30, "number of codes per page")
	length := flag.Uint("length", 6, "length of alphanumeric code (min 3)")
	flag.Parse()

	codes := []code{}
	gen := make(chan code)
	quit := make(chan struct{})
	go (func() {
		for {
			select {
			case gen <- newCode(int(*length)):
			case <-quit:
				close(gen)
				return
			}
		}
	})()

	var next chan code = gen
	for {
		m := <-next
		codes = append(codes, m)
		if len(codes) == int(*count) {
			break
		}
		next = m.noDuplicates(next)
	}
	close(quit)

	var output bytes.Buffer
	for i := 0; i < int(*page); i++ {
		fmt.Fprint(&output, "Discount Code ", i+1, ",")
	}
	fmt.Fprint(&output, "\n")
	for i, c := range codes {
		fmt.Fprint(&output, c)
		if (i+1)%int(*page) != 0 {
			fmt.Fprint(&output, ",")
		} else {
			fmt.Fprint(&output, "\n")
		}
	}
	fmt.Print(output.String())
}

type code []int

func newCode(size int) code {
	c := code{}
	b := make([]byte, size)
	n, err := io.ReadFull(rand.Reader, b)
	if n < len(b) || err != nil {
		panic(err)
	}
	for x := range b {
		c = append(c, int(b[x]%36)) // gets a base36 number that should still be random
	}
	return c
}

func (c code) equals(d code) bool {
	for _, cb := range c {
		for _, db := range d {
			if cb != db {
				return false
			}
		}
	}
	return true
}

func (c code) noDuplicates(in chan code) chan code {
	out := make(chan code)
	go func() {
		defer close(out)
		for n := range in {
			if !c.equals(n) {
				out <- n
			}
		}
	}()
	return out
}

func (c code) String() string {
	var s = ""
	for _, n := range c {
		s = s + strconv.FormatInt(int64(n), 36)
	}
	return s
}
