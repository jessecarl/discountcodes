package main

import (
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
	head := make(chan code)
	tail := make(chan code)
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

	var n code
	for {
		var source chan code
		if len(codes) < int(*count) {
			source = gen
		} else {
			close(head)
			close(quit)
			break
		}
		select {
		case n = <-source:
			if len(codes) == 0 {
				codes = append(codes, n)
				go n.noDuplicates(head, tail)
				n = nil
			}
		case m := <-tail:
			codes = append(codes, m)
			ch := make(chan code)
			go m.noDuplicates(tail, ch)
			tail = ch
		case head <- n:
		}
	}
	for i := 0; i < int(*page); i++ {
		fmt.Print("Discount Code ", i+1, ",")
	}
	fmt.Print("\n")
	for i, c := range codes {
		fmt.Print(c)
		if (i+1)%int(*page) != 0 {
			fmt.Print(",")
		} else {
			fmt.Print("\n")
		}
	}
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

func (c code) noDuplicates(in, out chan code) {
	for n := range in {
		if !c.equals(n) {
			out <- n
		}
	}
	close(out)
}

func (c code) String() string {
	var s = ""
	for _, n := range c {
		s = s + strconv.FormatInt(int64(n), 36)
	}
	return s
}
