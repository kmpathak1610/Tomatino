package orders

import "testing"

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from, to string
		ok       bool
	}{
		{"PENDING_RESTAURANT_ACCEPTANCE", "ACCEPTED", true},
		{"PENDING_RESTAURANT_ACCEPTANCE", "DELIVERED", false},
		{"ACCEPTED", "PREPARING", true},
		{"PREPARING", "READY_FOR_PICKUP", true},
		{"READY_FOR_PICKUP", "PICKED_UP", true},
		{"PICKED_UP", "OUT_FOR_DELIVERY", true},
		{"OUT_FOR_DELIVERY", "DELIVERED", true},
		{"DELIVERED", "CANCELLED", false},
	}
	for _, c := range cases {
		if got := canTransition(c.from, c.to); got != c.ok {
			t.Errorf("canTransition(%s,%s)=%v want %v", c.from, c.to, got, c.ok)
		}
	}
}