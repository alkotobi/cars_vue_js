package groupauth

import "testing"

func TestLoadUnrestrictedWhenAbsent(t *testing.T) {
	load("", false)
	if Restricted() {
		t.Fatal("absent WBOT_PERMS must be unrestricted")
	}
	for _, n := range Names() {
		if !Allow(n) {
			t.Fatalf("unrestricted caller must allow %q", n)
		}
	}
}

func TestLoadRestrictedEmpty(t *testing.T) {
	load("", true)
	if !Restricted() {
		t.Fatal("present-but-empty WBOT_PERMS must be restricted")
	}
	for _, n := range Names() {
		if Allow(n) {
			t.Fatalf("empty grant must deny %q", n)
		}
	}
}

func TestLoadGrantedSubset(t *testing.T) {
	load(Cost+", "+Suppliers, true)
	if !Restricted() {
		t.Fatal("must be restricted")
	}
	if !Allow(Cost) || !Allow(Suppliers) {
		t.Fatal("granted capabilities must be allowed")
	}
	if Allow(Edit) {
		t.Fatal("ungranted capability must be denied")
	}
}

func TestLoadIgnoresBlanks(t *testing.T) {
	load(" , "+Edit+" ,", true)
	if !Allow(Edit) {
		t.Fatal("Edit should be granted")
	}
	if Allow(Cost) {
		t.Fatal("Cost should not be granted")
	}
}

func TestValid(t *testing.T) {
	if !Valid(Edit) || !Valid(Suppliers) || !Valid(Cost) {
		t.Fatal("known names must be valid")
	}
	if Valid("nope") {
		t.Fatal("unknown name must be invalid")
	}
}
