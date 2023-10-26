package nodes

import (
	"fmt"
	"log"
	"strings"
)

// NewGroup creates a new group of nodes and connections that can then later be instantiated
// one or more times to make a more complex design. If the fullName includes a dot ('.')
// and an instance name, then the creation of the group is immediately followed by
// a call to `AddNode` using that same group and instance name. Otherwise, if no dot
// is included, the `fullName` will be used as the name of the group.
func (b *Builder) NewGroup(fullName string, fn func(b *Builder) *Builder, args ...string) *Builder {
	var hasInstanceName bool
	groupName := fullName
	if parts := strings.Split(fullName, "."); len(parts) > 1 {
		groupName = parts[0]
		hasInstanceName = true
	}

	if _, ok := b.Groups[groupName]; ok || groupName == "" {
		b.errs = append(b.errs, fmt.Errorf("already defined group %q", groupName))
		return b
	}

	if b.c.debug {
		log.Printf("NewGroup(%q) calling NewBuilder", groupName)
	}

	gb := b.c.NewBuilder()
	gb.isGroup = true
	if b.c.debug {
		log.Printf("NewGroup(%q) calling builder fn(gb)", groupName)
	}
	gb = fn(gb)
	b.Groups[groupName] = gb

	if b.c.debug {
		log.Printf("NewGroup(%q) returning with %v steps in recorded group", groupName, len(gb.groupRecorder))
	}

	if hasInstanceName {
		if b.c.debug {
			log.Printf("NewGroup(%q) instantiating new instance of node", fullName)
		}
		b = b.AddNode(fullName, args...)
	}

	return b
}

// Input is used within a group to connect one of its inputs to an internal input.
// It can only be used within a group.
func (b *Builder) Input(inputName, connectTo string) *Builder {
	if !b.isGroup {
		b.errs = append(b.errs, fmt.Errorf("Input(%q,%q) must only be called within a NewGroup builder", inputName, connectTo))
		return b
	}

	b.groupRecorder = append(b.groupRecorder, &recorder{
		action: "Input",
		args:   []string{inputName, connectTo},
	})
	return b
}

// Output is used within a group to connect one of its outputs to the group output.
// It can only be used within a group.
func (b *Builder) Output(connectFrom, outputName string) *Builder {
	if !b.isGroup {
		b.errs = append(b.errs, fmt.Errorf("Output(%q,%q) must only be called within a NewGroup builder", connectFrom, outputName))
		return b
	}

	b.groupRecorder = append(b.groupRecorder, &recorder{
		action: "Output",
		args:   []string{connectFrom, outputName},
	})
	return b
}
