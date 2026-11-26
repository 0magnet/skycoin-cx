//go:build cxfx

package graphics

import (
	"github.com/go-gl/glfw/v3.3/glfw"

	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// A window is an i32 handle into State.CxfxWindowTable, returned by
// glfw.CreateWindow, see #432.

func OpGlfwFullscreen(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	window := cxfxWindowGet(inputs[0].I32Get(prgrm))
	monitor := glfw.GetPrimaryMonitor()
	mode := monitor.GetVideoMode()
	if inputs[1].BoolGet(prgrm) {
		window.SetMonitor(monitor, 0, 0, mode.Width, mode.Height, mode.RefreshRate)
	} else {
		window.SetMonitor(nil, int(inputs[2].I32Get(prgrm)), int(inputs[3].I32Get(prgrm)), int(inputs[4].I32Get(prgrm)), int(inputs[5].I32Get(prgrm)), mode.RefreshRate)
	}
	window.MakeContextCurrent()
}

func OpGlfwInit(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	glfw.Init()
	cxfxState.CxfxGlfwInitialized = true
}

func OpGlfwGetCursorPos(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	x, y := cxfxWindowGet(inputs[0].I32Get(prgrm)).GetCursorPos()
	outputs[0].F64Set(prgrm, x)
	outputs[1].F64Set(prgrm, y)
}

// VIOLATION STRHOT-170: B, hands the window title to GLFW, #432.
func OpGlfwCreateWindow(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	window, err := glfw.CreateWindow(int(inputs[0].I32Get(prgrm)), int(inputs[1].I32Get(prgrm)),
		stringpool.StringView(ast.ProgramStringBytesRead(prgrm, &inputs[2])), nil, nil)
	if err != nil {
		panic(err)
	}
	outputs[0].I32Set(prgrm, cxfxWindowAdd(window))
}

func OpGlfwGetWindowContentScale(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	xscale, yscale := cxfxWindowGet(inputs[0].I32Get(prgrm)).GetContentScale()
	outputs[0].F32Set(prgrm, xscale)
	outputs[1].F32Set(prgrm, yscale)
}

func OpGlfwGetMonitorContentScale(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	xscale, yscale := glfw.GetPrimaryMonitor().GetContentScale()
	outputs[0].F32Set(prgrm, xscale)
	outputs[1].F32Set(prgrm, yscale)
}

func OpGlfwGetFramebufferSize(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	width, height := cxfxWindowGet(inputs[0].I32Get(prgrm)).GetFramebufferSize()
	outputs[0].I32Set(prgrm, int32(width))
	outputs[1].I32Set(prgrm, int32(height))
}

func OpGlfwGetWindowPos(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	x, y := cxfxWindowGet(inputs[0].I32Get(prgrm)).GetPos()
	outputs[0].I32Set(prgrm, int32(x))
	outputs[1].I32Set(prgrm, int32(y))
}

func OpGlfwGetWindowSize(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	width, height := cxfxWindowGet(inputs[0].I32Get(prgrm)).GetSize()
	outputs[0].I32Set(prgrm, int32(width))
	outputs[1].I32Set(prgrm, int32(height))
}

func OpGlfwPollEvents(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	if cxfxState.CxfxGlfwInitialized {
		glfw.PollEvents()
	}
	PollEvents(prgrm)
}

func glfwSetKeyCallback(window *glfw.Window) {
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		KeyboardEventPush(ActionType(action), int32(key), int64(scancode), int32(mods))
	})
}

func glfwSetCursorPosCallback(window *glfw.Window, eventType EventType) {
	window.SetCursorPosCallback(func(w *glfw.Window, xpos float64, ypos float64) {
		MouseEventPush(eventType, ActionMove, 0, -1, 0, xpos, ypos)
	})
}

func glfwSetMouseButtonCallback(window *glfw.Window, eventType EventType) {
	window.SetMouseButtonCallback(func(w *glfw.Window, key glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		x, y := w.GetCursorPos()
		MouseEventPush(eventType, ActionType(action), int32(key), -1, int32(mods), x, y)
	})
}

func OpGlfwSetKeyCallback(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) { // TODO : to deprecate
	glfwSetKeyCallback(cxfxWindowGet(inputs[0].I32Get(prgrm)))
	cxfxAppCallbackSet(prgrm, AppKeyboard, inputs)
}

func OpGlfwSetCursorPosCallback(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) { // TODO : to deprecate
	glfwSetCursorPosCallback(cxfxWindowGet(inputs[0].I32Get(prgrm)), AppCursorPos)
	cxfxAppCallbackSet(prgrm, AppCursorPos, inputs)
}

func OpGlfwSetMouseButtonCallback(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) { // TODO : to deprecate
	glfwSetMouseButtonCallback(cxfxWindowGet(inputs[0].I32Get(prgrm)), AppMouseButton)
	cxfxAppCallbackSet(prgrm, AppMouseButton, inputs)
}

func OpGlfwSetKeyboardCallback(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	glfwSetKeyCallback(cxfxWindowGet(inputs[0].I32Get(prgrm)))
	cxfxAppCallbackSet(prgrm, AppKeyboard, inputs)
}

func OpGlfwSetMouseCallback(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	window := cxfxWindowGet(inputs[0].I32Get(prgrm))
	glfwSetCursorPosCallback(window, AppMouse)
	glfwSetMouseButtonCallback(window, AppMouse)
	cxfxAppCallbackSet(prgrm, AppMouse, inputs)
}

func OpGlfwSetFramebufferSizeCallback(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	cxfxWindowGet(inputs[0].I32Get(prgrm)).SetFramebufferSizeCallback(func(w *glfw.Window, width int, height int) {
		FramebufferSizeEventPush(float64(width), float64(height)) // TODO : to deprecate, use float64
	})
	cxfxAppCallbackSet(prgrm, AppFramebufferSize, inputs)
}

func OpGlfwSetWindowSizeCallback(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	cxfxWindowGet(inputs[0].I32Get(prgrm)).SetSizeCallback(func(w *glfw.Window, width int, height int) {
		WindowSizeEventPush(float64(width), float64(height)) // TODO : to deprecate, use float64
	})
	cxfxAppCallbackSet(prgrm, AppWindowSize, inputs)
}

func OpGlfwSetWindowPosCallback(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	cxfxWindowGet(inputs[0].I32Get(prgrm)).SetPosCallback(func(w *glfw.Window, x int, y int) {
		WindowPositionEventPush(float64(x), float64(y)) // TODO to deprecate, use float64
	})
	cxfxAppCallbackSet(prgrm, AppWindowPosition, inputs)
}

func OpGlfwSetStartCallback(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	cxfxWindowGet(inputs[0].I32Get(prgrm))
	cxfxAppCallbackSet(prgrm, AppStart, inputs)
	EventPush(AppStart)
}

func OpGlfwSetStopCallback(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	cxfxWindowGet(inputs[0].I32Get(prgrm))
	cxfxAppCallbackSet(prgrm, AppStop, inputs)
}
