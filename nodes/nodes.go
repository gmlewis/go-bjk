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

// Client represents all the known nodes in Blackjack from its Luau bindings.
type Client struct {
	ls *lua.State
}

// New creates a new instance of nodes.Client.
func New() *Client {
	ls := lua.NewState()
	lua.OpenLibraries(ls)
	return &Client{ls: ls}
}

// Bootstrap parses all the initialization scripts used by Blackjack.
func (c *Client) Bootstrap(blackjackRepoPath string) error {
	// First, process the files in the correct bootstrap order:
	for _, path := range bootstrapOrder {
		fullPath := filepath.Join(blackjackRepoPath, path)
		log.Printf("Processing file: %v", fullPath)
		if err := lua.DoFile(c.ls, fullPath); err != nil {
			return err
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
			return lua.DoFile(c.ls, fullPath)
		}); err != nil {
			return err
		}
	}

	return nil
}

var bootstrapOrder = []string{
	"blackjack_engine/src/lua_engine/node_params.lua",
	"blackjack_engine/src/lua_engine/node_library.lua",
	"blackjack_engine/src/lua_engine/blackjack_utils.lua",
	"blackjack_lua/lib/priority_queue.lua",
	"blackjack_lua/lib/table_helpers.lua",
	"blackjack_lua/lib/vector_math.lua",
	"blackjack_lua/lib/gizmo_helpers.lua",
}

var blackjackSubdirs = []string{
	// "blackjack_engine/src/lua_engine",
	// "blackjack_lua/lib",
	"blackjack_lua/run",
}
