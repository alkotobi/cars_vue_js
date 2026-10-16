package main

import (
	"testing"

	"whatsapp-agent-tool/internal/groupauth"
)

func TestRequiredPermission(t *testing.T) {
	cases := []struct {
		name string
		cmd  string
		rest []string
		want string
	}{
		{"faq add is a write", "faq", []string{"add"}, groupauth.Edit},
		{"faq update is a write", "faq", []string{"update"}, groupauth.Edit},
		{"faq search is open", "faq", []string{"search"}, ""},
		{"price set-car is a write", "price", []string{"set-car"}, groupauth.Edit},
		{"price list-cars is open", "price", []string{"list-cars"}, ""},
		{"price check is open", "price", []string{"check"}, ""},

		{"suppliers read needs supplier", "suppliers", nil, groupauth.Suppliers},
		{"suppliers add is a write", "suppliers", []string{"add"}, groupauth.Edit},

		{"procurement suppliers read needs supplier", "procurement", []string{"suppliers"}, groupauth.Suppliers},
		{"procurement suppliers add is a write", "procurement", []string{"suppliers", "add"}, groupauth.Edit},
		{"procurement suppliers update is a write", "procurement", []string{"suppliers", "update"}, groupauth.Edit},
		{"procurement quotations read is open (redacted)", "procurement", []string{"quotations"}, ""},
		{"procurement quotations create is a write", "procurement", []string{"quotations", "create"}, groupauth.Edit},
		{"procurement quotations set-price is a write", "procurement", []string{"quotations", "set-price"}, groupauth.Edit},
		{"procurement documents link is a write", "procurement", []string{"documents", "link"}, groupauth.Edit},
		{"procurement payments set-swift is a write", "procurement", []string{"payments", "set-swift"}, groupauth.Edit},
		{"procurement outstanding is open (redacted)", "procurement", []string{"outstanding"}, ""},

		{"inquiries create is a write", "inquiries", []string{"create"}, groupauth.Edit},
		{"inquiries add-car is a write", "inquiries", []string{"add-car"}, groupauth.Edit},
		{"inquiries list is open (redacted)", "inquiries", []string{"list"}, ""},

		{"clearance set is a write", "clearance", []string{"set"}, groupauth.Edit},
		{"clearance read is open", "clearance", nil, ""},

		{"buy-payments read is open (redacted)", "buy-payments", nil, ""},
		{"buy-payments delete is a write", "buy-payments", []string{"delete"}, groupauth.Edit},

		{"brands add is a write", "brands", []string{"add"}, groupauth.Edit},
		{"brands read is open", "brands", nil, ""},
		{"car-names add is a write", "car-names", []string{"add"}, groupauth.Edit},

		{"todo add is a write", "todo", []string{"add"}, groupauth.Edit},
		{"todo done is a write", "todo", []string{"done"}, groupauth.Edit},
		{"todo list is open", "todo", []string{"list"}, ""},
		{"todo due is open", "todo", []string{"due"}, ""},
		{"todo remind add is a write", "todo", []string{"remind", "add"}, groupauth.Edit},
		{"todo remind list is open", "todo", []string{"remind", "list"}, ""},

		{"quote is open (redacted)", "quote", []string{"--car", "X"}, ""},
		{"stock is open (redacted)", "stock", []string{"X"}, ""},
		{"freight needs cost", "freight", []string{"--port", "ALGIERS"}, groupauth.Cost},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := requiredPermission(tc.cmd, tc.rest); got != tc.want {
				t.Fatalf("requiredPermission(%q, %v) = %q, want %q", tc.cmd, tc.rest, got, tc.want)
			}
		})
	}
}
