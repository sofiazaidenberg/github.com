package main

import "code.google.com/p/go-tour/tree"
import "fmt"

// Walk walks the tree t sending all values
// from the tree to the channel ch.
func Walk(t *tree.Tree, ch chan int) {
	fmt.Println("walking...")
	if t == nil {
		fmt.Println("tree finished")
	} else {
		ch <- t.Value
		if t.Left == nil && t.Right == nil {
			close(ch)
			return
		}
		Walk(t.Left, ch)
		Walk(t.Right, ch)
	}
}

// Same determines whether the trees
// t1 and t2 contain the same values.
//func Same(t1, t2 *tree.Tree) bool

func main() {
	ch := make(chan int, 10)
	go Walk(tree.New(1), ch)
	for i := range ch {
		fmt.Println(i)
	}
}
