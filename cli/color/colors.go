// Steven Phillips / elimisteve
// 2016.03.28

package color

import (
	"fmt"
	"strings"

	"github.com/elimisteve/cryptag/types"
	"github.com/fatih/color"
)

var (
	BlackOnWhite = color.New(color.FgBlack, color.BgHiWhite).SprintFunc()
	BlackOnCyan  = color.New(color.FgBlack, color.BgHiCyan).SprintFunc()
)

// Map returns a copy of the given []string with each contained string
// colorized with the 'colorize' function.
//
// Example: fmt.Printf("%s", Map(BlackOnWhite, []string{"tag1", "tag2", "tag3"}))
func Map(colorize func(...interface{}) string, strs []string) []string {
	outStrs := make([]string, len(strs))
	for i := range strs {
		outStrs[i] = colorize(strs[i])
	}
	return outStrs
}

// Tags returns a colorized list of the given tags
func Tags(tags []string) string {
	return strings.Join(Map(BlackOnWhite, tags), "   ")
}

func TextRow(r *types.Row) string {
	text := BlackOnCyan(string(r.Decrypted()))
	tags := Map(BlackOnWhite, r.PlainTags())
	return fmt.Sprintf("%s    %s", text, strings.Join(tags, "   "))
}

func TextRows(rows types.Rows) string {
	cRows := make([]string, 0, len(rows))
	for i := range rows {
		cRows = append(cRows, TextRow(rows[i]))
	}
	return strings.Join(cRows, "\n\n")
}
