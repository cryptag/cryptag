// Steven Phillips / elimisteve
// 2016.06.05

package cryptask

type Task struct {
	ID          string `json:"-"` // Save as tag, not in Row.decrypted
	Title       string
	Description string
	Assignee    string `json:"-"` // Save as tag, not in Row.decrypted
}
