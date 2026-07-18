package main

import "testing"

func TestClassifyGesture(t *testing.T) {
	tests := []struct {
		name                       string
		startX, startY, endX, endY int32
		want                       Action
	}{
		{"up advances", 600, 1300, 610, 700, ActionNext},
		{"down returns", 600, 400, 590, 1000, ActionPrevious},
		{"tap ignored", 600, 800, 605, 810, ActionNone},
		{"horizontal ignored", 600, 800, 1000, 820, ActionNone},
		{"left edge right exits", 80, 800, 600, 830, ActionExit},
		{"off-axis escape ignored", 80, 800, 600, 1200, ActionNone},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := classifyGesture(test.startX, test.startY, test.endX, test.endY); got != test.want {
				t.Fatalf("classifyGesture() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestDetectorTypeBSequence(t *testing.T) {
	detector := NewGestureDetector()
	events := []InputEvent{
		{Type: evAbs, Code: absMTSlot, Value: 0},
		{Type: evAbs, Code: absMTTrackingID, Value: 42},
		{Type: evAbs, Code: absMTPositionX, Value: 600},
		{Type: evAbs, Code: absMTPositionY, Value: 1300},
		{Type: evAbs, Code: absMTPositionX, Value: 610},
		{Type: evAbs, Code: absMTPositionY, Value: 700},
		{Type: evAbs, Code: absMTTrackingID, Value: -1},
	}

	var got Action
	for _, event := range events {
		if action := detector.Handle(event); action != ActionNone {
			got = action
		}
	}
	if got != ActionNext {
		t.Fatalf("detector action = %v, want %v", got, ActionNext)
	}
}

func TestTripleTapRightAdvances(t *testing.T) {
	detector := NewGestureDetector()
	var got Action
	for tap := int32(0); tap < 3; tap++ {
		seconds := int32(100)
		microseconds := tap * 300_000
		events := []InputEvent{
			{Sec: seconds, Usec: microseconds, Type: evAbs, Code: absMTSlot, Value: 0},
			{Sec: seconds, Usec: microseconds, Type: evAbs, Code: absMTTrackingID, Value: tap + 1},
			{Sec: seconds, Usec: microseconds, Type: evAbs, Code: absMTPositionX, Value: 900},
			{Sec: seconds, Usec: microseconds, Type: evAbs, Code: absMTPositionY, Value: 800},
			{Sec: seconds, Usec: microseconds + 80_000, Type: evAbs, Code: absMTTrackingID, Value: -1},
		}
		for _, event := range events {
			if action := detector.Handle(event); action != ActionNone {
				got = action
			}
		}
	}
	if got != ActionNext {
		t.Fatalf("triple tap action = %v, want %v", got, ActionNext)
	}
}

func TestTripleTapLeftReturns(t *testing.T) {
	detector := NewGestureDetector()
	var got Action
	for tap := int32(0); tap < 3; tap++ {
		events := []InputEvent{
			{Sec: 200, Usec: tap * 250_000, Type: evAbs, Code: absMTTrackingID, Value: tap + 1},
			{Sec: 200, Usec: tap * 250_000, Type: evAbs, Code: absMTPositionX, Value: 300},
			{Sec: 200, Usec: tap * 250_000, Type: evAbs, Code: absMTPositionY, Value: 900},
			{Sec: 200, Usec: tap*250_000 + 60_000, Type: evAbs, Code: absMTTrackingID, Value: -1},
		}
		for _, event := range events {
			if action := detector.Handle(event); action != ActionNone {
				got = action
			}
		}
	}
	if got != ActionPrevious {
		t.Fatalf("triple tap action = %v, want %v", got, ActionPrevious)
	}
}
