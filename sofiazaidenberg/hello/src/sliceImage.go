package main

import "code.google.com/p/go-tour/pic"

func Pic(dx, dy int) [][]uint8 {
	res := make([][]uint8, dy)
	for i := range res {
		elem := make([]uint8, dx)
		for j := range elem {
			elem[j] = (uint8(i) * uint8(j))
		}
		res[i] = elem
	}
	return res
}

func main() {
	pic.Show(Pic)
}
