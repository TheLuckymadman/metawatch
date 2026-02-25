package model

import (
	"encoding/json"
	"time"
)

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	var duration string
	if err := json.Unmarshal(b, &duration); err != nil {
		return err
	}
	dur, err := time.ParseDuration(duration)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

func (d *Duration) String() string {
	return time.Duration(*d).String()
}

func (d *Duration) Set(s string) error {
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}
