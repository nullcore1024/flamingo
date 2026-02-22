package base

import (
	"time"
)

type Timestamp struct {
	t time.Time
}

func NewTimestamp() *Timestamp {
	return &Timestamp{t: time.Now()}
}

func NewTimestampFromTime(t time.Time) *Timestamp {
	return &Timestamp{t: t}
}

func (ts *Timestamp) Time() time.Time {
	return ts.t
}

func (ts *Timestamp) Unix() int64 {
	return ts.t.Unix()
}

func (ts *Timestamp) UnixNano() int64 {
	return ts.t.UnixNano()
}

func (ts *Timestamp) String() string {
	return ts.t.String()
}

func (ts *Timestamp) Format(layout string) string {
	return ts.t.Format(layout)
}

func (ts *Timestamp) Add(d time.Duration) *Timestamp {
	return &Timestamp{t: ts.t.Add(d)}
}

func (ts *Timestamp) Sub(other *Timestamp) time.Duration {
	return ts.t.Sub(other.t)
}

func (ts *Timestamp) Before(other *Timestamp) bool {
	return ts.t.Before(other.t)
}

func (ts *Timestamp) After(other *Timestamp) bool {
	return ts.t.After(other.t)
}

func (ts *Timestamp) Equal(other *Timestamp) bool {
	return ts.t.Equal(other.t)
}
