package color

import (
	"os"
	"strings"
)

// Disabled can be set to true to suppress all color output.
var Disabled bool

func init() {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		Disabled = true
	}
}

func wrap(code, s string) string {
	if Disabled {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}

func Green(s string) string  { return wrap("32", s) }
func Yellow(s string) string { return wrap("33", s) }
func Red(s string) string    { return wrap("31", s) }
func Bold(s string) string   { return wrap("1", s) }
func Dim(s string) string    { return wrap("2", s) }

// DisableIfFlag checks args for --no-color and sets Disabled.
func DisableIfFlag(args []string) {
	for _, a := range args {
		if strings.TrimLeft(a, "-") == "no-color" {
			Disabled = true
			return
		}
	}
}
