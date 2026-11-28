//go:build !nocxos

package fileopcodes

import (
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
)

// OsLibraryState is the state of the os library, its one global, see D139 on
// #412. OsLibraryCallPool holds the text of one opcode call, a cleaned path
// or the words of a command, and is reset when the call starts.
// OsLibraryPool holds the paths that outlive a call, such as the path of a
// running CPU profile, interned once each.
type OsLibraryState struct {
	OsLibraryCallPool   stringpool.StringPool
	OsLibraryPool       stringpool.StringPool
	OsLibraryProfileMng OsProfileMng
	// OsLibraryCommandWords holds the Gids of the words of the os.Run command
	// in OsLibraryCallPool.
	OsLibraryCommandWords []stringpool.StringGid
}

// OsLibrary is the state of the os library. The run loop calls the os
// opcodes one at a time, so the call pool serves one call at a time.
var OsLibrary OsLibraryState

// OsLibraryCallPoolReset empties the call pool for a new opcode call and
// returns it.
func (state *OsLibraryState) OsLibraryCallPoolReset() *stringpool.StringPool {
	state.OsLibraryCallPool.StringPoolReset()
	return &state.OsLibraryCallPool
}

// OsProfile is one running CPU profile, the Gid of its path in
// OsLibraryPool and the function that stops it.
type OsProfile struct {
	OsProfilePath stringpool.StringGid
	OsProfileStop func() error
}

// OsProfileMng holds the running CPU profiles in a backing array. A stopped
// profile leaves its slot with OsProfilePath StringGidNil for the next one.
type OsProfileMng struct {
	OsProfileMngSlots []OsProfile
}

// OsProfileFind returns the slot of the running profile at path, or -1.
func (profiles *OsProfileMng) OsProfileFind(path stringpool.StringGid) int {
	for slot := range profiles.OsProfileMngSlots {
		if profiles.OsProfileMngSlots[slot].OsProfilePath == path {
			return slot
		}
	}
	return -1
}

// OsProfileAdd records a running profile at path, reusing a free slot.
func (profiles *OsProfileMng) OsProfileAdd(path stringpool.StringGid, stop func() error) {
	profile := OsProfile{OsProfilePath: path, OsProfileStop: stop}
	if slot := profiles.OsProfileFind(path); slot >= 0 {
		profiles.OsProfileMngSlots[slot] = profile
		return
	}
	if slot := profiles.OsProfileFind(stringpool.StringGidNil); slot >= 0 {
		profiles.OsProfileMngSlots[slot] = profile
		return
	}
	profiles.OsProfileMngSlots = append(profiles.OsProfileMngSlots, profile)
}
