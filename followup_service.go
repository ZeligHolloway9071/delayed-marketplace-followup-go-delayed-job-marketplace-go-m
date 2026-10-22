package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const followUpQueue = "marketplace-follow-ups"

type FollowUp struct {
	SellerAsset, BuyerUpdate, OrderHandoff string
	DelayHours                             int
}

func shouldFollowUp(f FollowUp) bool {
	return f.SellerAsset != "" && f.BuyerUpdate != "" && f.OrderHandoff != "" && f.DelayHours > 0
}
func scheduleHandler(c *InfraiClient, wurl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var f FollowUp
		if json.NewDecoder(r.Body).Decode(&f) != nil || !shouldFollowUp(f) {
			http.Error(w, "invalid follow-up", 400)
			return
		}
		if e := c.QueuePublish(followUpQueue, f); e != nil {
			http.Error(w, e.Error(), 502)
			return
		}
		jid, e := c.CronCreate(fmt.Sprintf("0 %d * * *", (time.Now().Hour()+f.DelayHours)%24), wurl)
		if e != nil {
			http.Error(w, e.Error(), 502)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"job_id": jid, "status": "scheduled"})
	}
}
func main() {
	c, e := NewInfraiClient()
	if e != nil {
		panic(e)
	}
	http.HandleFunc("/schedule", scheduleHandler(c, "http://localhost:8080/follow-up"))
	http.HandleFunc("/follow-up", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(202) })
	_ = http.ListenAndServe(":8080", nil)
}
