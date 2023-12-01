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

	"github.com/gmlewis/go-bjk/ast"
	"github.com/mitchellh/go-homedir"
	lua "github.com/yuin/gopher-lua"
)

// Client represents all the known nodes in Blackjack from its Luau bindings.
type Client struct {
	Nodes map[string]*ast.Node

	debug bool
	ls    *lua.LState

	// cached Eval ensures that a design is only evaluated once even if it
	// is written ToSTL, ToObj, or any other formats.
	cachedMesh *Mesh

	// used during Eval:
	extParamsLookup map[string]*ast.ValueEnum
}

// New creates a new instance of nodes.Client.
// blackjackRepoPath is either the absolute path to the Blackjack repo or
// is the relative-to-$HOME-dir path of the repo.
func New(blackjackRepoPath string, debug bool) (*Client, error) {
	if _, err := os.Stat(blackjackRepoPath); err != nil {
		homeDir, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		blackjackRepoPath = filepath.Join(homeDir, blackjackRepoPath)
		if _, err = os.Stat(blackjackRepoPath); err != nil {
			return nil, err
		}
	}

	ls := lua.NewState()
	ls.OpenLibs()
	if debug {
		log.Printf("At start: Top=%v", ls.GetTop())
	}

	registerMeshType(ls)
	registerNativeMathType(ls)
	registerOpsType(ls)
	registerPrimitivesType(ls)
	registerSelectionExpressionType(ls)
	registerVec3Type(ls)
	if err := ls.DoString("vector = Vec3.new"); err != nil {
		log.Fatalf("lua.DoString('vector = Vec3.new'): %v", err)
	}

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
			if d.IsDir() || !strings.HasSuffix(path, ".lua") {
				return nil
			}
			fullPath := filepath.Join(root, path)
			if debug {
				log.Printf("Processing file: %v", fullPath)
			}
			return ls.DoFile(fullPath)
		}); err != nil {
			return nil, err
		}
	}

	c := &Client{debug: debug, ls: ls}
	ns, err := c.list()
	if err != nil {
		return nil, err
	}
	c.Nodes = ns

	return c, nil
}

// Close closes the current client.
func (c *Client) Close() {
	c.ls.Close()
}

var blackjackSubdirs = []string{
	"blackjack_lua/fonts",
	"blackjack_lua/run",
}

var packagePaths = []string{
	"blackjack_engine/src/lua_engine/",
	"blackjack_engine/src/lua_engine/node_",
	"blackjack_engine/src/lua_engine/blackjack_",
	"blackjack_lua/lib/",
}
