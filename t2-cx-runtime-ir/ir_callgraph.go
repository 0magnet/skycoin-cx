package ir

// CallSite is one OpcodeCall instruction, at CallSitePc of CallSiteCaller,
// that calls CallSiteCallee.
type CallSite struct {
	CallSiteCaller uint32
	CallSitePc     uint32
	CallSiteCallee uint32
}

// CallGraph is the graph of the OpcodeCall instructions of an image,
// decision D144 and optimizations/0004. The slices are indexed by function
// id. A function in a cycle calls itself, or calls a function that calls it
// back. CallGraphCycleCalls are the calls that close a cycle, one per cycle
// at least, in the order a depth first walk from the entries finds them.
//
// CallGraphDepthOf, CallGraphStackOf and CallGraphNext hold for a function
// whose CallGraphReachesCycle is false: the most calls in progress below
// it, the stack bytes it needs with its callees, its frame and the frame or
// library block of its deepest call, and the callee on its deepest chain or
// ImageFuncNone. CallGraphBounded says that no entry reaches a cycle, and
// then CallGraphDepth, CallGraphChain and CallGraphStack are the same for
// the whole program, the stack from address 0 as the entries run.
type CallGraph struct {
	CallGraphCycle        []bool
	CallGraphCycleCalls   []CallSite
	CallGraphReachesCycle []bool
	CallGraphDepthOf      []uint32
	CallGraphStackOf      []uint32
	CallGraphNext         []uint32
	CallGraphBounded      bool
	CallGraphDepth        uint32
	CallGraphChain        []uint32
	CallGraphStack        uint32
}

// callGraphWalk is the state of the depth first walk of CallGraphBuild,
// Tarjan's strongly connected components with the walk stack kept apart.
type callGraphWalk struct {
	image     *Image
	graph     *CallGraph
	index     []uint32
	low       []uint32
	onStack   []bool
	walking   []bool
	stack     []uint32
	nextIndex uint32
}

// CallGraphBuild returns the call graph of image.
func CallGraphBuild(image *Image) CallGraph {
	count := len(image.ImageFuncs)
	graph := CallGraph{
		CallGraphCycle:        make([]bool, count),
		CallGraphReachesCycle: make([]bool, count),
		CallGraphDepthOf:      make([]uint32, count),
		CallGraphStackOf:      make([]uint32, count),
		CallGraphNext:         make([]uint32, count),
	}
	walk := callGraphWalk{
		image:   image,
		graph:   &graph,
		index:   make([]uint32, count),
		low:     make([]uint32, count),
		onStack: make([]bool, count),
		walking: make([]bool, count),
	}
	starts := make([]uint32, 0, count+2)
	for _, entry := range [...]uint32{image.ImageEntryInit, image.ImageEntryMain} {
		if entry != ImageFuncNone {
			starts = append(starts, entry)
		}
	}
	for id := range image.ImageFuncs {
		starts = append(starts, uint32(id))
	}
	for _, id := range starts {
		if walk.index[id] == 0 {
			walk.visit(id)
		}
	}
	graph.CallGraphBounded = true
	for _, entry := range [...]uint32{image.ImageEntryInit, image.ImageEntryMain} {
		if entry != ImageFuncNone && graph.CallGraphReachesCycle[entry] {
			graph.CallGraphBounded = false
		}
	}
	if !graph.CallGraphBounded {
		return graph
	}
	chainStart := ImageFuncNone
	if init := image.ImageEntryInit; init != ImageFuncNone {
		chainStart = init
		graph.CallGraphDepth = graph.CallGraphDepthOf[init]
		graph.CallGraphStack = graph.CallGraphStackOf[init]
	}
	if main := image.ImageEntryMain; main != ImageFuncNone {
		if chainStart == ImageFuncNone || graph.CallGraphDepthOf[main] > graph.CallGraphDepth {
			chainStart = main
			graph.CallGraphDepth = graph.CallGraphDepthOf[main]
		}
		// main runs past the frame of *init, as vm runs it.
		mainFrame := uint32(0)
		if init := image.ImageEntryInit; init != ImageFuncNone {
			mainFrame = image.ImageFuncs[init].FuncFrameSize
		}
		graph.CallGraphStack = max(graph.CallGraphStack, mainFrame+graph.CallGraphStackOf[main])
	}
	for id := chainStart; id != ImageFuncNone; id = graph.CallGraphNext[id] {
		graph.CallGraphChain = append(graph.CallGraphChain, id)
	}
	return graph
}

// visit walks the callees of id. A call of a function still being walked
// closes a cycle. When id is the root of its component, the component is
// complete with every callee outside it, so its depth and stack are known.
func (walk *callGraphWalk) visit(id uint32) {
	walk.nextIndex++
	walk.index[id], walk.low[id] = walk.nextIndex, walk.nextIndex
	walk.stack = append(walk.stack, id)
	walk.onStack[id], walk.walking[id] = true, true
	code := walk.image.ImageFuncs[id].FuncCode
	for pc := range code {
		if code[pc].InstrOpcode != OpcodeCall {
			continue
		}
		callee := code[pc].InstrM
		if callee == id {
			walk.graph.CallGraphCycle[id] = true
		}
		if walk.walking[callee] {
			walk.graph.CallGraphCycleCalls = append(walk.graph.CallGraphCycleCalls, CallSite{CallSiteCaller: id, CallSitePc: uint32(pc), CallSiteCallee: callee})
		}
		if walk.index[callee] == 0 {
			walk.visit(callee)
			walk.low[id] = min(walk.low[id], walk.low[callee])
		} else if walk.onStack[callee] {
			walk.low[id] = min(walk.low[id], walk.index[callee])
		}
	}
	walk.walking[id] = false
	if walk.low[id] != walk.index[id] {
		return
	}
	first := len(walk.stack) - 1
	for walk.stack[first] != id {
		first--
	}
	component := walk.stack[first:]
	walk.stack = walk.stack[:first]
	for _, member := range component {
		walk.onStack[member] = false
		if len(component) > 1 {
			walk.graph.CallGraphCycle[member] = true
		}
	}
	for _, member := range component {
		walk.measure(member)
	}
}

// measure sets the depth and stack of id once its callees have theirs.
func (walk *callGraphWalk) measure(id uint32) {
	graph := walk.graph
	fn := &walk.image.ImageFuncs[id]
	graph.CallGraphNext[id] = ImageFuncNone
	if graph.CallGraphCycle[id] {
		graph.CallGraphReachesCycle[id] = true
		return
	}
	depth, below := uint32(0), uint32(0)
	for pc := range fn.FuncCode {
		in := &fn.FuncCode[pc]
		switch in.InstrOpcode {
		case OpcodeCallFrame:
			below = max(below, in.InstrN)
		case OpcodeCall:
			callee := in.InstrM
			if graph.CallGraphReachesCycle[callee] {
				graph.CallGraphReachesCycle[id] = true
				return
			}
			below = max(below, graph.CallGraphStackOf[callee])
			if graph.CallGraphNext[id] == ImageFuncNone || graph.CallGraphDepthOf[callee]+1 > depth {
				depth, graph.CallGraphNext[id] = graph.CallGraphDepthOf[callee]+1, callee
			}
		}
	}
	graph.CallGraphDepthOf[id] = depth
	graph.CallGraphStackOf[id] = fn.FuncFrameSize + below
}
