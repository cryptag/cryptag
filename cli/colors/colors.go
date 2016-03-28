// Steven Phillips / elimisteve
// 2016.03.28

package colors

import "github.com/fatih/color"

var (
	BlackOnWhite = color.New(color.FgBlack, color.BgHiWhite).SprintFunc()
	BlackOnCyan  = color.New(color.FgBlack, color.BgHiCyan).SprintFunc()
)

func Map(colorize func(...interface{}) string, strs []string) []string {
	outStrs := make([]string, len(strs))
	for i := range strs {
		outStrs[i] = colorize(strs[i])
	}
	return outStrs
}
