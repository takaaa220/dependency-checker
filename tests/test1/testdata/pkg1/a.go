package pkg1

import (
	"fmt"

	"github.com/takaaa220/dependency-checker-test1/pkg2"
)

func Hello() {
	fmt.Println("hello")
	pkg2.Hello()
}
