package scope

import (
	"strings"
	"testing"
)

func TestTokenFixedLength(t *testing.T) {
	tok, err := Token("bot")
	if err != nil || len(tok) != tokenLen {
		t.Fatalf("bot token len %d err %v", len(tok), err)
	}
}

func TestTokenInjectivePairs(t *testing.T) {
	pairs := []struct {
		a, b string
	}{
		{"routine:row1", "routine:row10"},
		{"task:1", "task:10"},
	}
	for _, p := range pairs {
		ta, err := Token(p.a)
		if err != nil {
			t.Fatalf("%s: %v", p.a, err)
		}
		tb, err := Token(p.b)
		if err != nil {
			t.Fatalf("%s: %v", p.b, err)
		}
		if ta == tb {
			t.Fatalf("%s and %s collided: %q", p.a, p.b, ta)
		}
		if len(ta) != tokenLen || len(tb) != tokenLen {
			t.Fatalf("lengths %d %d", len(ta), len(tb))
		}
	}
}

func TestTokenUsesNonNamePadChar(t *testing.T) {
	tok, err := Token("routine:row1")
	if err != nil {
		t.Fatal(err)
	}
	suffix := tok[len("rcesroutinerow1"):]
	if suffix != strings.Repeat(string(padChar), tokenLen-len("rcesroutinerow1")) {
		t.Fatalf("bad pad suffix %q", suffix)
	}
}
