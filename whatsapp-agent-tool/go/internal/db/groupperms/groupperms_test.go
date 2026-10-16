package groupperms

import "testing"

// The driver returns integer columns as int64/uint64 when a query carries
// parameters (binary protocol) and as float64/[]byte otherwise. mapRow must
// decode the group id and the three flags from every one of those shapes — a
// float64-only decoder silently mapped every group to id 0 (regression).
func TestMapRowDecodesEveryDriverShape(t *testing.T) {
	shapes := []struct {
		name string
		id   any
		flag any
	}{
		{"int64 (prepared stmt)", int64(8), int64(1)},
		{"uint64", uint64(8), uint64(1)},
		{"float64 (text protoc)", float64(8), float64(1)},
		{"[]byte", []byte("8"), []byte("1")},
		{"string", "8", "1"},
	}
	for _, s := range shapes {
		t.Run(s.name, func(t *testing.T) {
			p := mapRow(map[string]any{
				"tenant":                    "mig_26",
				"id_chat_group":             s.id,
				"can_give_edit_instruction": s.flag,
				"can_c_suppliers_info":      s.flag,
				"can_c_cost":                s.flag,
			})
			if p.GroupID != 8 {
				t.Fatalf("GroupID = %d, want 8", p.GroupID)
			}
			if p.Tenant != "mig_26" {
				t.Fatalf("Tenant = %q, want mig_26", p.Tenant)
			}
			if !p.Edit || !p.Suppliers || !p.Cost {
				t.Fatalf("flags = %+v, want all true", p)
			}
		})
	}
}

func TestMapRowZeroAndNil(t *testing.T) {
	p := mapRow(map[string]any{
		"tenant":                    "t",
		"id_chat_group":             int64(0),
		"can_give_edit_instruction": nil,
		"can_c_suppliers_info":      int64(0),
		"can_c_cost":                []byte("0"),
	})
	if p.GroupID != 0 || p.Edit || p.Suppliers || p.Cost {
		t.Fatalf("zero rows must decode to all-false, got %+v", p)
	}
}
