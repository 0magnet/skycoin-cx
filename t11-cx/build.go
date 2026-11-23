package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/skycoin/skycoin-cx/t0-types"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
	"github.com/skycoin/skycoin-cx/t2-cx-runtime-ir"
	"github.com/skycoin/skycoin-cx/t3-ast"
	llvmir "github.com/skycoin/skycoin-cx/t3-cx-compiler-llvmir"
)

// The --target values of the llvmir prototype. llvmir writes the module,
// native also compiles it and the runtime with clang.
const (
	targetLlvmir = "llvmir"
	targetNative = "native"
)

// buildTarget writes image as LLVM IR to output, or a native executable for
// targetNative. It returns the exit code.
func buildTarget(options cxCmdFlags, prgrm *ast.CXProgram, image *ir.Image) int {
	if options.target != targetLlvmir && options.target != targetNative {
		fmt.Fprintf(os.Stderr, "unknown --target %q, want llvmir or native\n", options.target)
		return types.CX_COMPILATION_ERROR
	}
	var emitOptions llvmir.Options
	if !options.llvmExact {
		emitOptions.ReadOnly = buildReadOnly(prgrm, image)
	}
	module, stats, err := llvmir.Emit(image, emitOptions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot emit LLVM IR: %v\n", err)
		return types.CX_RUNTIME_NOT_IMPLEMENTED
	}
	if options.llvmStats {
		buildStatsPrint(stats)
	}
	output := options.output
	if options.target == targetLlvmir {
		if output == "" {
			output = "a.ll"
		}
		if err := os.WriteFile(output, module, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return types.CX_COMPILATION_ERROR
		}
		return types.CX_SUCCESS
	}
	if output == "" {
		output = "a.out"
	}
	dir, err := os.MkdirTemp("", "cx-native-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return types.CX_COMPILATION_ERROR
	}
	defer os.RemoveAll(dir)
	ll, rt := filepath.Join(dir, "prog.ll"), filepath.Join(dir, "runtime.c")
	if err := os.WriteFile(ll, module, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return types.CX_COMPILATION_ERROR
	}
	if err := os.WriteFile(rt, llvmir.RuntimeC, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return types.CX_COMPILATION_ERROR
	}
	cc := os.Getenv("CC")
	if cc == "" {
		cc = "clang"
	}
	cmd := exec.Command(cc, "-O2", "-Wno-override-module", "-o", output, ll, rt, "-lm")
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %v\n", cc, err)
		return types.CX_COMPILATION_ERROR
	}
	return types.CX_SUCCESS
}

// buildStatsPrint lists each opcode of the image by how it was emitted.
func buildStatsPrint(stats llvmir.Stats) {
	seen := map[ir.Opcode]bool{}
	for op := range stats.StatsDirect {
		seen[op] = true
	}
	for op := range stats.StatsFallback {
		seen[op] = true
	}
	ops := make([]int, 0, len(seen))
	for op := range seen {
		ops = append(ops, int(op))
	}
	sort.Ints(ops)
	for _, op := range ops {
		o := ir.Opcode(op)
		fmt.Fprintf(os.Stderr, "llvmstat %s direct %d runtime %d\n", llvmir.OpName(o), stats.StatsDirect[o], stats.StatsFallback[o])
	}
	fmt.Fprintf(os.Stderr, "llvmstat regions alloca %d shadow %d\n", stats.StatsAlloca, stats.StatsShadow)
}

// buildReadOnly lists the literals of prgrm that lie in the data segment of
// image. A literal is an argument or type signature without a name, see the
// Name field of ast.CXArgument and actions.WritePrimary.
func buildReadOnly(prgrm *ast.CXProgram, image *ir.Image) []llvmir.Span {
	start, end := image.ImageDataStart, ir.ImageDataEnd(image)
	var spans []llvmir.Span
	add := func(offset types.Pointer, size types.Pointer) {
		o, s := uint64(offset), uint64(size)
		if s == 0 || o < uint64(start) || o+s > uint64(end) {
			return
		}
		spans = append(spans, llvmir.Span{SpanStart: uint32(o), SpanEnd: uint32(o + s)})
	}
	args := &prgrm.CxIrState.CXArgumentMng
	for i := range args.CXArgumentMngObjects {
		if args.CXArgumentMngLive[i] && args.CXArgumentMngObjects[i].Name == stringpool.StringGidNil {
			arg := &args.CXArgumentMngObjects[i]
			add(arg.Offset, arg.GetSize())
		}
	}
	signatures := &prgrm.CxIrState.CXTypeSignatureMng
	for i := range signatures.CXTypeSignatureMngObjects {
		if signatures.CXTypeSignatureMngLive[i] && signatures.CXTypeSignatureMngObjects[i].Name == stringpool.StringGidNil {
			signature := &signatures.CXTypeSignatureMngObjects[i]
			if signature.Type == ast.TYPE_ATOMIC {
				add(signature.Offset, signature.GetSize(prgrm, false))
			}
		}
	}
	return spans
}
