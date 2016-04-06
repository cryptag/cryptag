// Steven Phillips / elimisteve
// 2016.04.06

package color

import (
	"fmt"

	colorable "github.com/mattn/go-colorable"
)

func Print(a ...interface{}) {
	fmt.Fprint(colorable.NewColorableStdout(), a...)
}

func Printf(format string, a ...interface{}) {
	fmt.Fprintf(colorable.NewColorableStdout(), format, a...)
}

func Println(a ...interface{}) {
	fmt.Fprintln(colorable.NewColorableStdout(), a...)
}
