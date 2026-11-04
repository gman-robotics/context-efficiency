package rules

import (
	"strings"
	"testing"
)

func TestDeniedTokenTable(t *testing.T) {
	deny := []string{
		"send", "Send", "SEND", "merge", "Merge", "deploy", "payment", "pay", "payout",
		"deployment", "sendmerge", "charge", "Charge", "refund", "wire", "WiRe", "invoice", "money",
		"send-merge", "money, please",
	}
	for _, tok := range deny {
		if DeniedToken(tok) == "" {
			t.Fatalf("want deny %q", tok)
		}
	}
	allow := []string{"sender", "payload", "Sender", "Payload", "wireless", "charged", "Wireless", "Charged"}
	for _, tok := range allow {
		if got := DeniedToken(tok); got != "" {
			t.Fatalf("want allow %q got %q", tok, got)
		}
	}
}

func TestRefuseActualMerge(t *testing.T) {
	err := Refuse(PathGo, "write_file", "merge", nil)
	if err == nil || err.Error() != "refused: denied token" {
		t.Fatal(err)
	}
	err = Refuse(PathGo, "write_file", "nope", nil)
	if err == nil || err.Error() != "refused: actual" {
		t.Fatal(err)
	}
	err = Refuse("other", "write_file", "go", nil)
	if err == nil || err.Error() != "refused: path" {
		t.Fatal(err)
	}
	err = Refuse(PathGo, "payment", "go", nil)
	if err == nil || err.Error() != "refused: denied token" {
		t.Fatal(err)
	}
	if err := Refuse(PathGo, "sender", "go", nil); err != nil {
		t.Fatal(err)
	}
}

func TestScreenBodyLimits(t *testing.T) {
	if err := Screen(PathGo, "write_file", "go", nil, strings.Repeat("a", 2000)); err != nil {
		t.Fatal(err)
	}
	err := Screen(PathGo, "write_file", "go", nil, strings.Repeat("a", 2001))
	if err == nil || err.Error() != "refused: body too large" {
		t.Fatal(err)
	}
	err = Screen(PathGo, "write_file", "go", nil, "  ")
	if err == nil || err.Error() != "refused: empty body" {
		t.Fatal(err)
	}
	err = Screen(PathGo, "write_file", "go", nil, "user@host")
	if err == nil || err.Error() != "refused: body" {
		t.Fatal(err)
	}
	err = Screen(PathGo, "write_file", "go", nil, "id 1234567 end")
	if err == nil || err.Error() != "refused: body" {
		t.Fatal(err)
	}
	if err := Screen(PathGo, "write_file", "go", nil, "id 123456 end"); err != nil {
		t.Fatal(err)
	}
	err = Screen(PathGo, "write_file", "go", nil, "see <private>x</private> now")
	if err == nil || err.Error() != "refused: private" {
		t.Fatal(err)
	}
	if err := Screen(PathGo, "write_file", "go", nil, "money is in the state and merge too"); err != nil {
		t.Fatal(err)
	}
	err = Screen(PathGo, "write_file", "go", []string{"Invoice"}, "hello state")
	if err == nil || err.Error() != "refused: denied token" {
		t.Fatal(err)
	}
}
