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

	"github.com/skycoin/skycoin-cx/t6-cx-library-graphics"
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
					graphics.SetGLContext(glctx)
					graphics.SetGOApp(a)
					graphics.EventPush(graphics.AppStart)
				case lifecycle.CrossOff:
					graphics.EventPush(graphics.AppStop)
				}
			case size.Event:
				graphics.FramebufferSizeEventPush(float64(e.WidthPx), float64(e.HeightPx))
				graphics.WindowSizeEventPush(float64(e.WidthPx), float64(e.HeightPx))
			case paint.Event:
				graphics.EventPush(graphics.AppPaint)
				//if glctx == nil || e.External {
				//	continue
				//}
			case touch.Event:
				cxfxAction := graphics.ActionRelease
				switch e.Type {
				case touch.TypeBegin:
					cxfxAction = graphics.ActionPress
				case touch.TypeMove:
					cxfxAction = graphics.ActionMove
				case touch.TypeEnd:
					cxfxAction = graphics.ActionRelease
				}
				graphics.MouseEventPush(graphics.AppMouse, cxfxAction, int32(e.Sequence), int64(e.Sequence), 0, float64(e.X), float64(e.Y))
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
