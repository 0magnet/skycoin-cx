//go:build cxfx

package graphics

import (
	"sync"
	"time"

	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin-cx/t4-cx-runtime-interpreter" // thin: only Callback; no depth gain
)

const (
	ActionRelease ActionType = 0 // TODO : break from glfw.Action
	ActionPress              = 1
	ActionRepeat             = 2
	ActionMove               = 3
)

const (
	AppNone EventType = iota
	AppStart
	AppStop
	AppKeyboard
	AppMouse
	AppFocusOn
	AppFocusOff
	AppFramebufferSize
	AppWindowSize
	AppWindowPosition
	AppPaint
	AppCursorPos   // TODO : to deprecate
	AppMouseButton // TODO : to deprecate
	// AppEventCount is one past the last event type.
	AppEventCount
)

type ActionType uint32
type EventType uint32
type Event struct {
	eventType EventType
	action    ActionType
	time      uint64
	x         float64
	y         float64
	key       int32
	scancode  int64
	mods      int32
}

// cxfxAppCallbackSet records the CX function the events of eventType call and
// the window handle they pass. The compiler resolved the function, see #432.
func cxfxAppCallbackSet(prgrm *ast.CXProgram, eventType EventType, inputs []ast.CXValue) {
	cxfxState.CxfxAppCallbackTable[eventType] = Callback{
		CallbackWindow:   inputs[0].I32Get(prgrm),
		CallbackFunction: ast.CXFunctionGid(inputs[1].I64Get(prgrm)),
		CallbackSet:      true,
	}
}

// cxfxAppCallbackCall calls the callback of eventType with the window handle
// and the count inputs after it that are already written to the scratch.
func cxfxAppCallbackCall(prgrm *ast.CXProgram, eventType EventType, count int) {
	callback := &cxfxState.CxfxAppCallbackTable[eventType]
	if !callback.CallbackSet {
		return
	}
	cxfxCallbackInputsI32(0, callback.CallbackWindow)
	interpreter.Callback(prgrm, callback.CallbackFunction, cxfxState.CxfxCallbackInputs[:count+1])
}

// VIOLATION: state outside State. Owner: t6-cx-library-graphics; moves to EventMng in phase 1.
var eventCount int

// VIOLATION: state outside State. Owner: t6-cx-library-graphics; moves to EventMng in phase 1.
var events []Event

// VIOLATION: state outside State. Owner: t6-cx-library-graphics; moves to EventMng in phase 1.
var polled []Event
var polledCount int

// VIOLATION: state outside State. Owner: t6-cx-library-graphics; moves to EventMng in phase 1.
var mutex sync.Mutex

func EventPush(eventType EventType) {
	mutex.Lock()
	defer mutex.Unlock()

	// VIOLATION: created outside manager; phase 1.
	event := Event{
		eventType,
		ActionRelease, // TODO : ACTION_NONE
		uint64(time.Now().UnixNano()),
		0.0,
		0.0,
		0,
		0,
		0,
	}
	if eventCount < len(events) {
		events[eventCount] = event
	} else {
		events = append(events, event)
	}
	eventCount++
}

func WindowSizeEventPush(width float64, height float64) {
	index := eventCount
	EventPush(AppWindowSize)
	events[index].x = width
	events[index].y = height
}

func WindowPositionEventPush(x float64, y float64) {
	index := eventCount
	EventPush(AppWindowPosition)
	events[index].x = x
	events[index].y = y
}

func FramebufferSizeEventPush(width float64, height float64) {
	index := eventCount
	EventPush(AppFramebufferSize)
	events[index].x = width
	events[index].y = height
}

func KeyboardEventPush(action ActionType, key int32, scancode int64, mods int32) {
	index := eventCount
	EventPush(AppKeyboard)
	events[index].action = action
	events[index].key = key
	events[index].scancode = scancode
	events[index].mods = mods
}

func MouseEventPush(eventType EventType, action ActionType, button int32, touch int64, mods int32, x float64, y float64) {
	index := eventCount
	EventPush(eventType)
	events[index].action = action
	events[index].key = button
	events[index].scancode = touch
	events[index].mods = mods
	events[index].x = x
	events[index].y = y
}

func purgeEvents() {
	mutex.Lock()
	defer mutex.Unlock()

	if eventCount > 0 {
		polled = append(polled[0:0], events[0:eventCount]...)
		eventCount = 0
	}
}

func PollEvents(prgrm *ast.CXProgram) {
	purgeEvents()
	eventCount := len(polled)
	for i := 0; i < eventCount; i++ {
		e := &polled[i]
		switch e.eventType {
		case AppStart, AppStop:
			cxfxAppCallbackCall(prgrm, e.eventType, 0)
		case AppKeyboard:
			cxfxCallbackInputsI32(1, e.key)
			cxfxCallbackInputsI32(2, int32(e.scancode))
			cxfxCallbackInputsI32(3, int32(e.action))
			cxfxCallbackInputsI32(4, e.mods)
			cxfxAppCallbackCall(prgrm, e.eventType, 4)
		case AppMouse:
			cxfxCallbackInputsI32(1, e.key)
			cxfxCallbackInputsI64(2, e.scancode)
			cxfxCallbackInputsI32(3, int32(e.action))
			cxfxCallbackInputsI32(4, e.mods)
			cxfxCallbackInputsF64(5, e.x)
			cxfxCallbackInputsF64(6, e.y)
			cxfxAppCallbackCall(prgrm, e.eventType, 6)
		case AppFramebufferSize, AppWindowSize, AppWindowPosition:
			cxfxCallbackInputsI32(1, int32(e.x)) // TODO : use float64 (deprecate int32)
			cxfxCallbackInputsI32(2, int32(e.y)) // TODO : use float64 (deprecate int32)
			cxfxAppCallbackCall(prgrm, e.eventType, 2)
		case AppCursorPos: // TODO : to deprecate
			cxfxCallbackInputsF64(1, e.x)
			cxfxCallbackInputsF64(2, e.y)
			cxfxAppCallbackCall(prgrm, e.eventType, 2)
		case AppMouseButton: // TODO to deprecate
			cxfxCallbackInputsI32(1, e.key)
			cxfxCallbackInputsI32(2, int32(e.action))
			cxfxCallbackInputsI32(3, e.mods)
			cxfxAppCallbackCall(prgrm, e.eventType, 3)
		}
	}
	polled = polled[0:0]
}
