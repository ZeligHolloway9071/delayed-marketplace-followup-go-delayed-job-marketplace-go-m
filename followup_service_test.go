package main

import "testing"

func TestShouldFollowUp(t *testing.T) {
	ts := []struct {
		name string
		in   FollowUp
		want bool
	}{{"complete", FollowUp{"asset", "update", "order", 3}, true}, {"missing update", FollowUp{"asset", "", "order", 3}, false}, {"zero delay", FollowUp{"asset", "update", "order", 0}, false}}
	for _, tt := range ts {
		t.Run(tt.name, func(t *testing.T) {
			if g := shouldFollowUp(tt.in); g != tt.want {
				t.Fatalf("got %v want %v", g, tt.want)
			}
		})
	}
}
