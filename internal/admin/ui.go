package admin

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type UI struct {
	reader *bufio.Reader
}

func NewUI() *UI {
	return &UI{reader: bufio.NewReader(os.Stdin)}
}

func (u *UI) Prompt(label string) string {
	fmt.Print(label)
	value, _ := u.reader.ReadString('\n')
	return strings.TrimSpace(value)
}

func (u *UI) ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

func (u *UI) Pause() {
	fmt.Print("\nНажмите Enter, чтобы продолжить...")
	u.reader.ReadString('\n')
}
