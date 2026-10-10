package annotation

import "testing"

func TestBorderStyleForCML(t *testing.T) {
	for _, test := range []struct{ legacy, modern string }{
		{"", "solid"}, {"2,2", "dotted"}, {"4,2", "dashed"},
	} {
		if got := borderStyleForCML(test.legacy, true); got != test.modern {
			t.Errorf("borderStyleForCML(%q, true) = %q, want %q", test.legacy, got, test.modern)
		}
		if got := borderStyleForCML(test.legacy, false); got != test.legacy {
			t.Errorf("borderStyleForCML(%q, false) = %q, want %q", test.legacy, got, test.legacy)
		}
	}
}

func TestBorderStyleForTerraform(t *testing.T) {
	for _, test := range []struct{ modern, legacy string }{
		{"solid", ""}, {"dotted", "2,2"}, {"dashed", "4,2"},
	} {
		if got := borderStyleForTerraform(test.modern, true); got != test.legacy {
			t.Errorf("borderStyleForTerraform(%q, true) = %q, want %q", test.modern, got, test.legacy)
		}
		if got := borderStyleForTerraform(test.modern, false); got != test.modern {
			t.Errorf("borderStyleForTerraform(%q, false) = %q, want %q", test.modern, got, test.modern)
		}
	}
}
