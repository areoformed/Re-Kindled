package main

const (
	evAbs           = 0x03
	absMTSlot       = 0x2f
	absMTPositionX  = 0x35
	absMTPositionY  = 0x36
	absMTTrackingID = 0x39
)

type InputEvent struct {
	Sec   int32
	Usec  int32
	Type  uint16
	Code  uint16
	Value int32
}

type Action uint8

const (
	ActionNone Action = iota
	ActionNext
	ActionPrevious
	ActionExit
)

type contact struct {
	active       bool
	haveX, haveY bool
	startedAt    int64
	startX       int32
	startY       int32
	x            int32
	y            int32
}

type GestureDetector struct {
	slot      int32
	contacts  map[int32]*contact
	haveTap   bool
	tapCount  int
	lastTapAt int64
	lastTapX  int32
	lastTapY  int32
	tapXTotal int32
}

func NewGestureDetector() *GestureDetector {
	return &GestureDetector{contacts: make(map[int32]*contact)}
}

func (detector *GestureDetector) Handle(event InputEvent) Action {
	if event.Type != evAbs {
		return ActionNone
	}
	if event.Code == absMTSlot {
		detector.slot = event.Value
		return ActionNone
	}

	current := detector.contacts[detector.slot]
	if current == nil {
		current = &contact{}
		detector.contacts[detector.slot] = current
	}

	switch event.Code {
	case absMTTrackingID:
		if event.Value >= 0 {
			*current = contact{active: true, startedAt: event.milliseconds()}
			return ActionNone
		}
		if !current.active || !current.haveX || !current.haveY {
			*current = contact{}
			return ActionNone
		}
		action := classifyGesture(current.startX, current.startY, current.x, current.y)
		if action == ActionNone && isTap(*current, event.milliseconds()) {
			action = detector.recordTap(current.x, current.y, event.milliseconds())
		}
		*current = contact{}
		return action
	case absMTPositionX:
		current.x = event.Value
		if current.active && !current.haveX {
			current.startX = event.Value
			current.haveX = true
		}
	case absMTPositionY:
		current.y = event.Value
		if current.active && !current.haveY {
			current.startY = event.Value
			current.haveY = true
		}
	}
	return ActionNone
}

func (event InputEvent) milliseconds() int64 {
	return int64(event.Sec)*1000 + int64(event.Usec)/1000
}

func isTap(current contact, endedAt int64) bool {
	duration := endedAt - current.startedAt
	return abs32(current.x-current.startX) <= 80 &&
		abs32(current.y-current.startY) <= 80 &&
		duration >= 0 && duration <= 650
}

func (detector *GestureDetector) recordTap(x, y int32, at int64) Action {
	continued := detector.haveTap &&
		at >= detector.lastTapAt && at-detector.lastTapAt <= 1600 &&
		abs32(x-detector.lastTapX) <= 180 && abs32(y-detector.lastTapY) <= 180

	if !continued {
		detector.tapCount = 0
		detector.tapXTotal = 0
	}

	detector.haveTap = true
	detector.tapCount++
	detector.tapXTotal += x
	detector.lastTapAt = at
	detector.lastTapX = x
	detector.lastTapY = y

	if detector.tapCount < 3 {
		return ActionNone
	}

	averageX := detector.tapXTotal / int32(detector.tapCount)
	detector.haveTap = false
	detector.tapCount = 0
	detector.tapXTotal = 0
	if averageX < 618 {
		return ActionPrevious
	}
	return ActionNext
}

func classifyGesture(startX, startY, endX, endY int32) Action {
	dx := endX - startX
	dy := endY - startY
	absX := abs32(dx)
	absY := abs32(dy)

	// A deliberate right swipe beginning at the left bezel is the escape hatch.
	if startX < 220 && dx > 320 && absY < 220 {
		return ActionExit
	}
	if absY < 150 || absY*5 < absX*6 {
		return ActionNone
	}
	if dy < 0 {
		return ActionNext
	}
	return ActionPrevious
}

func abs32(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}
