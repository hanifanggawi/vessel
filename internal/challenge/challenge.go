// Package challenge gates a protected action behind a task the user must
// complete. Challenges are intentionally modular: a command only depends on
// the Challenge interface and the Run helper, so swapping in a different
// challenge type later is a one-line change at the call site.
package challenge

import (
	"bufio"
	"fmt"
	"io"
)

// Challenge is a gate the user must clear before a protected action proceeds.
// Implementations capture whatever they need at construction time and are
// otherwise stateless; a new challenge type only needs to satisfy this
// interface to be usable everywhere Run is called.
type Challenge interface {
	// Name identifies the challenge type.
	Name() string
	// Present returns the text shown to the user before they respond.
	Present() string
	// Verify reports whether the user's response clears the challenge.
	Verify(response string) bool
}

// Run presents c, reads a single line of input, and returns nil only if the
// response clears the challenge. Any failure (read error or wrong answer)
// returns a non-nil error so callers can abort the protected action.
func Run(challenge Challenge, input io.Reader, output io.Writer) error {
	fmt.Fprintln(output, challenge.Present())
	fmt.Fprint(output, "> ")

	response, err := bufio.NewReader(input).ReadString('\n')
	if err != nil && response == "" {
		return fmt.Errorf("could not read challenge response: %w", err)
	}

	if !challenge.Verify(response) {
		return fmt.Errorf("challenge failed: response did not match")
	}
	return nil
}
