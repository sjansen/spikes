package main

import (
	"fmt"
	"os"
	"time"

	"github.com/teambition/rrule-go"
)

func main() {
	r, _ := rrule.NewRRule(rrule.ROption{
		Freq:    rrule.DAILY,
		Count:   10,
		Dtstart: time.Date(1997, 9, 2, 9, 0, 0, 0, time.UTC),
	})

	fmt.Println(r.String())
	for _, t := range r.All() {
		fmt.Println(t)
	}

	r, err := rrule.StrToRRule(
		"DTSTART;TZID=America/New_York:19970905T090000\nRRULE:FREQ=MONTHLY;COUNT=10;BYDAY=1FR\n",
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	} else {
		fmt.Printf("%#v\n", r)
	}

	fmt.Println(r.String())
	for _, t := range r.All() {
		fmt.Println(t)
	}
}
