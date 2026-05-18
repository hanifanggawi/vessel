package challenge

import (
	"bytes"
	"strings"
	"testing"
)

func TestBackwardsVerify(test *testing.T) {
	backwards := &Backwards{sentence: "stay focused"}

	if !backwards.Verify("desucof yats") {
		test.Fatal("exact reversed answer should pass")
	}
	if !backwards.Verify("  desucof yats \n") {
		test.Fatal("surrounding whitespace should be tolerated")
	}
	if backwards.Verify("stay focused") {
		test.Fatal("the non-reversed sentence must not pass")
	}
}

func TestRunChallenge(test *testing.T) {
	backwards := &Backwards{sentence: "stay focused"}

	var output bytes.Buffer
	if err := Run(backwards, strings.NewReader("desucof yats\n"), &output); err != nil {
		test.Fatalf("correct answer should pass: %v", err)
	}
	if err := Run(backwards, strings.NewReader("wrong\n"), &output); err == nil {
		test.Fatal("wrong answer must return an error")
	}
}
