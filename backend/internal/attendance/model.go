package attendance

import "time"

type Record struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	Date         time.Time  `json:"date"`
	CheckinTime  *time.Time `json:"checkin_time,omitempty"`
	CheckoutTime *time.Time `json:"checkout_time,omitempty"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}

type Summary struct {
	From       time.Time `json:"from"`
	To         time.Time `json:"to"`
	Present    int       `json:"present"`
	OnTime     int       `json:"on_time"`
	Late       int       `json:"late"`
	Absent     int       `json:"absent"`
	WorkingDays int      `json:"working_days"`
}
