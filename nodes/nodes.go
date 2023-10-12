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

	lua "github.com/yuin/gopher-lua"
)

const (
	luaEngineDir = "blackjack_engine/src/lua_engine"
)

// Client represents all the known nodes in Blackjack from its Luau bindings.
type Client struct {
	ls *lua.LState
}

func (c *Client) showTop() {
	log.Printf("\n\nshowTop: Top=%v type: %v", c.ls.GetTop(), c.ls.Get(-1).Type())
}

// New creates a new instance of nodes.Client.
func New(blackjackRepoPath string) (*Client, error) {
	ls := lua.NewState()
	ls.OpenLibs()
	log.Printf("At start: Top=%v", ls.GetTop())

	pkg := ls.GetGlobal("package")
	packagePath := ls.GetField(pkg, "path").String()
	for _, s := range packagePaths {
		// cannot use filepath.Join here because it strips the trailing '/'
		packagePath = fmt.Sprintf("%v;%v/%v?.lua", packagePath, blackjackRepoPath, s)
	}
	ls.SetField(pkg, "path", lua.LString(packagePath))

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

var blackjackSubdirs = []string{
	"blackjack_lua/run",
}

var packagePaths = []string{
	"blackjack_engine/src/lua_engine/",
	"blackjack_engine/src/lua_engine/node_",
	"blackjack_engine/src/lua_engine/blackjack_",
	"blackjack_lua/lib/",
}
