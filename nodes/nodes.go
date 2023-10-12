// Package nodes loads the Blackjack Luau scripts and makes them available
// for building up BJK AST trees with a simple connection API.
// See: https://github.com/setzer22/blackjack
package nodes

import (
	"fmt"
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
	var results []string
	f := func(n string, v bool) {
		if v {
			results = append(results, n)
		}
	}
	f("IsBoolean", ls.IsBoolean(-1))
	f("IsFunction", ls.IsFunction(-1))
	f("IsGoFunction", ls.IsGoFunction(-1))
	f("IsLightUserData", ls.IsLightUserData(-1))
	f("IsNil", ls.IsNil(-1))
	f("IsNone", ls.IsNone(-1))
	f("IsNoneOrNil", ls.IsNoneOrNil(-1))
	f("IsNumber", ls.IsNumber(-1))
	f("IsString", ls.IsString(-1))
	f("IsTable", ls.IsTable(-1))
	f("IsThread", ls.IsThread(-1))
	f("IsUserData", ls.IsUserData(-1))
	log.Printf("\n\nshowTop: Top=%v - %v", ls.Top(), strings.Join(results, " - "))
}

// New creates a new instance of nodes.Client.
func New(blackjackRepoPath string) (*Client, error) {
	ls := lua.NewState()
	lua.OpenLibraries(ls)
	log.Printf("At start: Top=%v", ls.Top())

	ls.Global("package")
	showTop(ls)
	ls.Field(-1, "path")
	showTop(ls)
	packagePath, ok := ls.ToString(-1)
	log.Printf("packagePath='%v', ok=%v", packagePath, ok)
	ls.Pop(1)
	showTop(ls)
	for _, s := range packagePaths {
		// cannot use filepath.Join here because it strips the final '/'
		packagePath = fmt.Sprintf("%v;%v/%v?.lua", packagePath, blackjackRepoPath, s)
	}
	log.Printf("packagePath='%v'", packagePath)
	ls.PushString(packagePath)
	showTop(ls)
	ls.SetField(-2, "path")
	showTop(ls)
	ls.Pop(1)
	showTop(ls)

	/*
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
	*/

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
	"blackjack_engine/src/lua_engine/",
	"blackjack_engine/src/lua_engine/node_",
	"blackjack_engine/src/lua_engine/blackjack_",
	"blackjack_lua/lib/",
}
