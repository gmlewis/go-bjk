package nodes

import (
	"fmt"
	"log"
)

// NewGroup creates a new group of nodes and connections that can then later be instantiated
// one or more times to make a more complex design.
func (b *Builder) NewGroup(groupName, inputs, outputs string, fn func(b *Builder) *Builder) *Builder {
	if _, ok := b.Groups[groupName]; ok || groupName == "" {
		b.errs = append(b.errs, fmt.Errorf("already defined group %q", groupName))
		return b
	}

	if b.c.debug {
		log.Printf("NewGroup(%q,%q,%q) calling NewBuilder", groupName, inputs, outputs)
	}

	gb := b.c.NewBuilder()
	gb.isGroup = true
	if b.c.debug {
		log.Printf("NewGroup(%q,%q,%q) calling fn(gb)", groupName, inputs, outputs)
	}
	gb = fn(gb)
	b.Groups[groupName] = gb

	if b.c.debug {
		log.Printf("NewGroup(%q,%q,%q) returning with %v steps in recorded group", groupName, inputs, outputs, len(gb.groupRecorder))
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
