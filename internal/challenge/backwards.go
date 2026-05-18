package challenge

import (
	"fmt"
	"math/rand"
	"strings"
)

var motivationalSentences = []string{
	"Discipline is choosing between what you want now and what you want most",
	"The pain of discipline weighs ounces while the pain of regret weighs tons",
	"You will not always be motivated so you must learn to be disciplined",
	"Every action you take is a vote for the person you wish to become",
	"The future depends on what you do in this present moment",
}

// Backwards shows a motivational sentence and asks the user to type it back
// reversed, character for character.
type Backwards struct {
	sentence string
}

// NewBackwards picks a random motivational sentence for this challenge run.
func NewBackwards() *Backwards {
	return &Backwards{sentence: motivationalSentences[rand.Intn(len(motivationalSentences))]}
}

func (backwards *Backwards) Name() string { return "backwards" }

func (backwards *Backwards) Present() string {
	return fmt.Sprintf(
		"%q\n\nType the sentence above backwards (character for character) to continue:",
		backwards.sentence,
	)
}

func (backwards *Backwards) Verify(response string) bool {
	return strings.TrimSpace(response) == reverse(backwards.sentence)
}

func reverse(text string) string {
	runes := []rune(text)
	for left, right := 0, len(runes)-1; left < right; left, right = left+1, right-1 {
		runes[left], runes[right] = runes[right], runes[left]
	}
	return string(runes)
}
