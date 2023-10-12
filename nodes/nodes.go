// Package nodes loads the Blackjack Luau scripts and makes them available
// for building up BJK AST trees with a simple connection API.
// See: https://github.com/setzer22/blackjack
package nodes

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hexops/valast"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaEngineDir = "blackjack_engine/src/lua_engine"
)

// Client represents all the known nodes in Blackjack from its Luau bindings.
type Client struct {
	ls *lua.LState
}

// New creates a new instance of nodes.Client.
func New(blackjackRepoPath string) (*Client, error) {
	ls := lua.NewState()

	// First, process the files in the correct bootstrap order:
	for _, preload := range bootstrapOrder {
		fullPath := filepath.Join(blackjackRepoPath, preload.fn)
		log.Printf("Processing file: %v", fullPath)
		fn, err := ls.LoadFile(fullPath)
		if err != nil {
			return nil, err
		}

		if preload.name != "" {
			loaders, ok := ls.GetField(ls.Get(lua.RegistryIndex), "_LOADERS").(*lua.LTable)
			// loaded := ls.GetGlobal("_LOADED")
			log.Printf("loaders=%#v, ok=%v", loaders, ok)

			log.Printf("fn.Env=%#v", fn.Env)
			log.Printf("fn.Proto=%v", valast.String(fn.Proto))
			f := func(L *lua.LState) int { L.Push(fn); return 1 }
			ls.PreloadModule(preload.name, f)
			// loaders.RawSetString(preload.name, fn)
		}
	}

	// Now, process all *.lua files found in the blackjackSubdirs:
	for _, subdir := range blackjackSubdirs {
		root := filepath.Join(blackjackRepoPath, subdir)
		fileSystem := os.DirFS(root)
		if err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".lua") {
				return nil
			}
			fullPath := filepath.Join(root, path)
			log.Printf("Processing file: %v", fullPath)
			return ls.DoFile(fullPath)
		}); err != nil {
			return nil, err
		}
	}

	return &Client{ls: ls}, nil
}

// Close closes the current client.
func (c *Client) Close() {
	c.ls.Close()
}

type preloads struct {
	fn   string
	name string
}

var bootstrapOrder = []preloads{
	{fn: "blackjack_engine/src/lua_engine/node_params.lua", name: "params"},
	{fn: "blackjack_engine/src/lua_engine/node_library.lua", name: "node_library"},
	{fn: "blackjack_engine/src/lua_engine/blackjack_utils.lua", name: "utils"},
	{fn: "blackjack_lua/lib/priority_queue.lua", name: "priority_queue"},
	{fn: "blackjack_lua/lib/table_helpers.lua", name: "table_helpers"},
	{fn: "blackjack_lua/lib/vector_math.lua", name: "vector_math"},
	{fn: "blackjack_lua/lib/gizmo_helpers.lua", name: "gizmo_helpers"},
}

var blackjackSubdirs = []string{
	// "blackjack_engine/src/lua_engine",
	// "blackjack_lua/lib",
	"blackjack_lua/run",
}
