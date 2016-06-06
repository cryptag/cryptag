// Steven Phillips / elimisteve
// 2016.06.06

package cli

import (
	"fmt"
	"log"
	"os"
)

func ArgFatal(s interface{}) {
	log.Println("Not enough arguments included")
	Fatal(s)
}

func Fatal(s interface{}) {
	fmt.Printf("%v\n", s)
	os.Exit(0)
}
