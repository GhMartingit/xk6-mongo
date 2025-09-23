package main

import (
	"fmt"
	"github.com/xyproto/randomstring"
)

func main() {
	randomstring.Seed()
	fmt.Println(randomstring.HumanFriendlyString(7))
}
