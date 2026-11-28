package main

import (
	"fmt"
	"strings"

	"github.com/skycoin/skycoin-cx/t2-cx-runtime-ir"
)

// buildCallGraphReport returns a warning for each call of image that closes
// a cycle of calls, decision D144. A program without a cycle gets its
// deepest call chain and the stack it needs instead, unless quiet is set.
func buildCallGraphReport(image *ir.Image, quiet bool) string {
	var report strings.Builder
	graph := ir.CallGraphBuild(image)
	seen := map[string]bool{}
	for _, site := range graph.CallGraphCycleCalls {
		where := image.ImageFuncNames[site.CallSiteCaller]
		if line, found := ir.LineAt(image, site.CallSiteCaller, site.CallSitePc); found {
			where = fmt.Sprintf("%s:%d", image.ImageFileNames[line.LineFile], line.LineNumber)
		}
		warning := fmt.Sprintf("cx build: warning: %s: recursive call of %s from %s, the call depth of this program is not known at build time\n",
			where, image.ImageFuncNames[site.CallSiteCallee], image.ImageFuncNames[site.CallSiteCaller])
		if !seen[warning] {
			seen[warning] = true
			report.WriteString(warning)
		}
	}
	if quiet || !graph.CallGraphBounded || len(graph.CallGraphChain) == 0 {
		return report.String()
	}
	names := make([]string, len(graph.CallGraphChain))
	for i, id := range graph.CallGraphChain {
		names[i] = image.ImageFuncNames[id]
	}
	fmt.Fprintf(&report, "cx build: call depth %d (%s), stack %d of %d bytes\n",
		graph.CallGraphDepth, strings.Join(names, " -> "), graph.CallGraphStack, image.ImageStackSize)
	return report.String()
}
