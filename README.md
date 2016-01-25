# nnmc
Parser nnm-club forum for get films info

# install
go get github.com/serbe/nnmc

# example (not work if invalid user/password - view source code)
```go
package main

import (
	"fmt"

	"github.com/serbe/nnmc"
)

func main() {
	nnm, err := nnmc.Init("user", "password")
	if err != nil {
		panic(err)
	}
	tree, err := nnm.ParseForumTree("http://nnm-club.me/forum/viewforum.php?f=266")
	if err != nil {
		panic(err)
	}
	film0, err := nnm.ParseTopic(tree[0])
	if err != nil {
		panic(err)
	}
	fmt.Println(film0)
}
```