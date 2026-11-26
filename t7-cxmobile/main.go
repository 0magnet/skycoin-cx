//go:build cxfx && android

// Does not compile and no make target or CI job builds it, see #35.
package main

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"

	"github.com/skycoin/skycoin-cx/t6-cxfx"
)

func startCXVM() {
	log.Printf("-----------------------------------------------------------------------------")
	log.Printf("STARTING CXDROID N")
	log.Printf("-----------------------------------------------------------------------------")
	CXLogFile(true)
	if AssetsCopyToFilesDir() {
		filesDir := asset.GetFilesDir()
		CXSetWorkingDir(fmt.Sprintf("%s/", filesDir))
		args := parseRunCli(filesDir)
		log.Printf("ARGS %v", args)
		log.Printf("All assets were copied successfuly")
		Run(args)
	}
	log.Printf("-----------------------------------------------------------------------------")
	log.Printf("STOPPING CXDROID N")
	log.Printf("-----------------------------------------------------------------------------")
}

func eventLoop(a app.App) {
	go func() {
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageFocused) {
				case lifecycle.CrossOn:
					var glctx gl.Context
					glctx, _ = e.DrawContext.(gl.Context)
					cxfx.SetGLContext(glctx)
					cxfx.SetGOApp(a)
					cxfx.EventPush(cxfx.AppStart)
				case lifecycle.CrossOff:
					cxfx.EventPush(cxfx.AppStop)
				}
			case size.Event:
				cxfx.FramebufferSizeEventPush(float64(e.WidthPx), float64(e.HeightPx))
				cxfx.WindowSizeEventPush(float64(e.WidthPx), float64(e.HeightPx))
			case paint.Event:
				cxfx.EventPush(cxfx.AppPaint)
				//if glctx == nil || e.External {
				//	continue
				//}
			case touch.Event:
				cxfxAction := cxfx.ActionRelease
				switch e.Type {
				case touch.TypeBegin:
					cxfxAction = cxfx.ActionPress
				case touch.TypeMove:
					cxfxAction = cxfx.ActionMove
				case touch.TypeEnd:
					cxfxAction = cxfx.ActionRelease
				}
				cxfx.MouseEventPush(cxfx.AppMouse, cxfxAction, int32(e.Sequence), int64(e.Sequence), 0, float64(e.X), float64(e.Y))
			}
		}
	}()

	startCXVM()
}

func parseRunCli(filesDir string) []string {
	cli, _ := CXReadFile("run.cli")
	cliString := string(cli)
	log.Printf("RUN_CLI_STRING %s\n", cliString)
	args := make([]string, 0, 0)
	args = append(args, "--cxpath=.")
	args = append(args, strings.Split(strings.Replace(cliString, "\r\n", "\n", -1), "\n")...)
	argCount := len(args)
	validArgs := make([]string, 0, 0)
	for i := 0; i < argCount; i++ {
		if len(args[i]) > 0 {
			validArgs = append(validArgs, args[i])
		}
	}
	return validArgs
}

func main() {
	app.Main(func(a app.App) {
		eventLoop(a)
	})
}
