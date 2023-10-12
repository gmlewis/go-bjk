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

	"github.com/Shopify/go-lua"
)

const (
	luaEngineDir = "blackjack_engine/src/lua_engine"
)

// Client represents all the known nodes in Blackjack from its Luau bindings.
type Client struct {
	ls *lua.State
}

func showTop(ls *lua.State) {
	log.Printf("top of stack IsBoolean: %v", ls.IsBoolean(-1))
	log.Printf("IsFunction: %v", ls.IsFunction(-1))
	log.Printf("IsGoFunction: %v", ls.IsGoFunction(-1))
	log.Printf("IsLightUserData: %v", ls.IsLightUserData(-1))
	log.Printf("IsNil: %v", ls.IsNil(-1))
	log.Printf("IsNone: %v", ls.IsNone(-1))
	log.Printf("IsNoneOrNil: %v", ls.IsNoneOrNil(-1))
	log.Printf("IsNumber: %v", ls.IsNumber(-1))
	log.Printf("IsString: %v", ls.IsString(-1))
	log.Printf("IsTable: %v", ls.IsTable(-1))
	log.Printf("IsThread: %v", ls.IsThread(-1))
	log.Printf("IsUserData: %v", ls.IsUserData(-1))
}

// New creates a new instance of nodes.Client.
func New(blackjackRepoPath string) (*Client, error) {
	ls := lua.NewState()
	// lua.OpenLibraries(ls)

	// First, process the files in the correct bootstrap order:
	var preloaded []lua.RegistryFunction
	for _, preload := range bootstrapOrder {
		fullPath := filepath.Join(blackjackRepoPath, preload.fn)
		log.Printf("Preloading file: %v", fullPath)
		if err := lua.LoadFile(ls, fullPath, ""); err != nil {
			return nil, err
		}
		// showTop(ls)

		// ls.Global("Params")

		if preload.name == "" {
			log.Fatalf("missing preload name: %v", preload.fn)
		}

		// ls.Register(preload.name,
		fn := ls.ToGoFunction(-1)
		preloaded = append(preloaded, lua.RegistryFunction{
			Name:     preload.name,
			Function: fn,
		})

		// 		// loaders, ok := ls.GetField(ls.Get(lua.RegistryIndex), "_LOADERS").(*lua.LTable)
		// 		// // loaded := ls.GetGlobal("_LOADED")
		// 		// log.Printf("loaders=%#v, ok=%v", loaders, ok)
		//
		// 		// log.Printf("fn.Env=%#v", fn.Env)
		// 		// log.Printf("fn.Proto=%v", valast.String(fn.Proto))
		// 		f := func(L *lua.State) int {
		// 			// // From: baselib.go
		// 			// // src := L.ToString(1)
		// 			// top := L.GetTop()
		// 			// fn, err := L.LoadFile(fullPath) // src)
		// 			// if err != nil {
		// 			// 	L.Push(lua.LString(err.Error()))
		// 			// 	L.Panic(L)
		// 			// }
		// 			// L.Push(fn)
		// 			// L.Call(0, lua.MultRet)
		// 			// return L.GetTop() - top
		// 			L.Push(fn)
		// 			return 1
		// 		}
		// 		ls.PreloadModule(preload.name, f)
		// 		// loaders.RawSetString(preload.name, fn)
		// 		// }
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
			return lua.DoFile(ls, fullPath)
		}); err != nil {
			return nil, err
		}
	}

	return &Client{ls: ls}, nil
}

// // Close closes the current client.
// func (c *Client) Close() {
// 	c.ls.Close()
// }

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

var packagePaths = []string{
	"blackjack_engine/src/lua_engine",
	"blackjack_lua/lib",
}
