package pkg1

import (
	"fmt"

	"github.com/takaaa220/dependency-checker-test2/domain/pkg2"
	usecasepkg1 "github.com/takaaa220/dependency-checker-test2/usecase/pkg1"
)

func Hello() {
	fmt.Println("hello")
	pkg2.Hello()
	usecasepkg1.Hello()
}
