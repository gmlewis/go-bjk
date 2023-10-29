package nodes

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gmlewis/go-bjk/ast"
	lua "github.com/yuin/gopher-lua"
	"golang.org/x/exp/maps"
)

const (
	nodeXOffset = 360
	nodeYOffset = 60
)

// BuilderFunc is a func used to build up a design.
type BuilderFunc func(b *Builder) *Builder

// Builder represents a BJK builder.
type Builder struct {
	c    *Client
	errs []error

	isGroup                 bool
	groupRecorder           []*recorder
	lastMergeMesh           string
	groupFullInputPortNames map[string]bool

	Nodes     map[string]*ast.Node
	NodeOrder []string
	Groups    map[string]*Builder

	ExternalParameters ast.ExternalParameters

	InputsAlreadyConnected map[string]string

	CheckUnusedGroupInputs bool
}

type recorder struct {
	action string
	args   []string
}

// NewBuilder returns a new BJK Builder.
func (c *Client) NewBuilder() *Builder {
	return &Builder{
		c:                      c,
		Nodes:                  map[string]*ast.Node{},
		Groups:                 map[string]*Builder{},
		InputsAlreadyConnected: map[string]string{},
		CheckUnusedGroupInputs: true,

		groupFullInputPortNames: map[string]bool{},
	}
}

// AddNode adds a new node to the design with the optional args.
// A nodes's name always starts with the type of node that it is, whether
// it is a built-in type like 'MakeScalar' or a new group like 'CoilPair'.
// The type is followed by a dot ('.') then one or more optional
// (but recommended) label(s). A node generated from an instance of
// a Group uses the node's name, a dot, then the group's full name.
// When referring to inputs or outputs of a node, its full name is followed
// by a dot then the name of the input or output port. For example,
// "Helix.wire-1" or "Helix.wire-1.CoilPair.coils-1-2.start_angle".
func (b *Builder) AddNode(name string, args ...string) *Builder {
	if name == "" {
		b.errs = append(b.errs, errors.New("AddNode: name cannot be empty"))
		return b
	}

	if b.isGroup {
		b.groupRecorder = append(b.groupRecorder, &recorder{
			action: "AddNode",
			args:   append([]string{name}, args...),
		})
		return b
	}

	parts := strings.Split(name, ".")
	nodeType := parts[0]
	n, ok := b.c.Nodes[nodeType]
	if !ok {
		if g, ok := b.Groups[nodeType]; ok {
			return b.instantiateGroup(name, g, args...)
		}
		b.errs = append(b.errs, fmt.Errorf("AddNode: unknown node type '%v'", nodeType))
		return b
	}

	return b.instantiateNode(nodeType, name, n, args...)
}

// injectGroupName takes a fullPortName (e.g. "Type.a.b.c.d") and a groupName (e.g. "MyGroup.instance")
// and re-combines them such at the groupName is injected after the Type but before the labels of the fullPortName:
// e.g. newFullPortName = "Type.MyGroup.instance.a.b.c.d", newNodeName = "Type.MyGroup.instance.a.b.c", portName = "d"
func injectGroupName(fullPortName, groupName string) (newFullPortName, newNodeName, portName string) {
	parts := strings.Split(fullPortName, ".")
	baseName, portName := parts[0], strings.Join(parts[1:], ".")
	newNodeName = fmt.Sprintf("%v.%v", baseName, groupName)
	if portName == "" {
		return newNodeName, newNodeName, ""
	}
	// To generate the final "newNodeName" and "portName" correctly, the portName has to be split if
	// it already contains multiple parts.
	newFullPortName = fmt.Sprintf("%v.%v", newNodeName, portName)
	if len(parts) == 2 { // only one part to the portName - return as-is.
		return newFullPortName, newNodeName, portName
	}
	// Re-partition the pieces:
	newNodeName = fmt.Sprintf("%v.%v.%v", baseName, groupName, strings.Join(parts[1:len(parts)-1], "."))
	portName = parts[len(parts)-1]
	return newFullPortName, newNodeName, portName
}

func (b *Builder) instantiateGroup(groupName string, group *Builder, args ...string) *Builder {
	if b.c.debug {
		log.Printf("Instantiating group '%v' with %v steps and args: %+v", groupName, len(group.groupRecorder), args)
	}

	staticArgs := map[string]string{}
	for _, arg := range args {
		lhs, rhs, err := splitArg(arg)
		if err != nil {
			b.errs = append(b.errs, err)
			return b
		}
		staticArgs[lhs] = rhs
	}

	errFn := func(i int, step *recorder, msg string) error {
		return fmt.Errorf("Group '%v' step #%v of %v: %v: %v('%v')", groupName, i+1, len(group.groupRecorder), msg, step.action, strings.Join(step.args, "', '"))
	}

	// First pass - make sure we know all possible valid new node names (after instantiating) for this group
	validNewNodeNames := map[string]bool{}
	for _, step := range group.groupRecorder {
		if step.action != "AddNode" {
			continue
		}
		newFullNodeName, _, _ := injectGroupName(step.args[0], groupName) // not expecting a port name here.
		validNewNodeNames[newFullNodeName] = true
	}

	// Next pass - make a map of the declared inputs that match the provided static args
	// This map is keyed by the node name, whose value is one or more assignment statements.
	namedArgs := map[string][]string{}
	for i, step := range group.groupRecorder {
		if step.action != "Input" {
			continue
		}

		newFullInputPortName, newToNodeName, portName := injectGroupName(step.args[1], groupName)
		if portName == "" {
			b.errs = append(b.errs, errFn(i, step, "'to' node missing port"))
			return b
		}

		b.groupFullInputPortNames[newFullInputPortName] = true

		if staticArgs[step.args[0]] == "" {
			continue
		}

		if !validNewNodeNames[newToNodeName] {
			b.errs = append(b.errs, errFn(i, step, fmt.Sprintf("'to' node %q not found, valid choices are: %+v", newToNodeName, maps.Keys(validNewNodeNames))))
			return b
		}

		namedArgs[newToNodeName] = append(namedArgs[newToNodeName], fmt.Sprintf("%v=%v", portName, staticArgs[step.args[0]]))
	}

	// Final pass - whenever a named static argument is used, add it to the list of args to `AddNode`
	for i, step := range group.groupRecorder {
		if b.c.debug {
			log.Printf("Group '%v' step #%v of %v: %v('%v') ...", groupName, i+1, len(group.groupRecorder), step.action, strings.Join(step.args, "', '"))
		}

		switch step.action {
		case "AddNode":
			fullNodeName, _, _ := injectGroupName(step.args[0], groupName) // not expecting a port name here.
			newArgs := append([]string{}, step.args[1:]...)
			if v, ok := namedArgs[fullNodeName]; ok {
				newArgs = append(newArgs, v...)
			}
			if b.c.debug {
				log.Printf("calling: AddNode(%q, %+v)", fullNodeName, newArgs)
			}
			b = b.AddNode(fullNodeName, newArgs...)
		case "Connect":
			fullFromPortName, _, portName := injectGroupName(step.args[0], groupName)
			if portName == "" {
				b.errs = append(b.errs, errFn(i, step, "'from' node missing port"))
				return b
			}
			fullToPortName, _, portName := injectGroupName(step.args[1], groupName)
			if portName == "" {
				b.errs = append(b.errs, errFn(i, step, "'to' node missing port"))
				return b
			}
			if b.c.debug {
				log.Printf("calling: Connect(%q, %q)", fullFromPortName, fullToPortName)
			}
			b = b.Connect(fullFromPortName, fullToPortName)
		case "Input":
		case "Output":
		default:
			b.errs = append(b.errs, fmt.Errorf("programming error: unknown action %q", step.action))
		}
	}

	if b.c.debug {
		log.Printf("Completed group '%v' with %v steps", groupName, len(group.groupRecorder))
	}

	return b
}

func (b *Builder) instantiateNode(nodeType, name string, n *ast.Node, args ...string) *Builder {
	parts := strings.Split(name, ".")
	if len(parts) == 1 {
		// auto-generate a label if the node doesn't have one.
		name = fmt.Sprintf("%v.node-%v", name, len(b.NodeOrder)+1)
	}

	if _, ok := b.Nodes[name]; ok {
		b.errs = append(b.errs, fmt.Errorf("AddNode: node '%v' already exists", name))
		return b
	}

	// Make a deep copy of the node since this is a new instance and we don't want to share values.
	inputs, err := b.setInputValues(name, n.Inputs, args...)
	if err != nil {
		b.errs = append(b.errs, fmt.Errorf("setInputValues: %v", err))
		return b
	}
	outputs := make([]*ast.Output, 0, len(n.Outputs))
	for _, out := range n.Outputs {
		outputs = append(outputs, &ast.Output{Name: out.Name, DataType: out.DataType})
	}

	var nodePosition *ast.Vec2
	for _, arg := range args {
		const label = "node_position="
		if strings.HasPrefix(arg, label) {
			pos, ok := parseVec2(arg[len(label):])
			if !ok {
				b.errs = append(b.errs, fmt.Errorf("unable to parse node '%v' arg: '%v'", name, arg))
			}
			nodePosition = pos
		}
	}

	b.Nodes[name] = &ast.Node{
		OpName:      nodeType,
		ReturnValue: n.ReturnValue, // OK not to make a deep copy of ReturnValue - it doesn't change.
		Inputs:      inputs,
		Outputs:     outputs,

		Label: n.Label,
		Index: uint64(len(b.NodeOrder)), // 0-based indices

		NodePosition: nodePosition,
	}
	b.NodeOrder = append(b.NodeOrder, name)

	return b
}

// Connect connects the `from` node.output_port to the `to` node.input_port.
func (b *Builder) Connect(from, to string) *Builder {
	if b.c.debug {
		log.Printf("Connect(%q, %q)", from, to)
	}

	if b.isGroup {
		b.groupRecorder = append(b.groupRecorder, &recorder{
			action: "Connect",
			args:   []string{from, to},
		})
		return b
	}

	fromParts := strings.Split(from, ".")
	if len(fromParts) < 2 {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q, %q): unable to parse 'from' name: %[1]q want at least 2 parts, got %[3]v", from, to, len(fromParts)))
		return b
	}

	// pre-emptively fail on common mistakes that are otherwise hard to debug
	if _, ok := b.Nodes[from]; ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q, %q): missing port name on 'from' node: %q", from, to, from))
		return b
	}
	if _, ok := b.Nodes[to]; ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q, %q): missing port name on 'to' node: %q", from, to, to))
		return b
	}

	fromNodeName := strings.Join(fromParts[0:len(fromParts)-1], ".")
	fromOutputName := fromParts[len(fromParts)-1]
	fromNode, ok := b.Nodes[fromNodeName]
	if !ok {
		var connectionsMade int
		if g, ok := b.Groups[fromParts[0]]; ok {
			for _, step := range g.groupRecorder {
				if b.c.debug {
					log.Printf("Searching for group connection: fromNodeName=%q, fromOutputName=%q, step.action=%q, step.args=%+v", fromNodeName, fromOutputName, step.action, step.args)
				}
				if step.action == "Output" && step.args[1] == fromOutputName {
					connectionsMade++
					newFromPortName, _, portName := injectGroupName(step.args[0], fromNodeName)
					if portName == "" {
						msg := "Connect(%q, %q): 'from' node missing port"
						if b.c.debug {
							log.Fatalf("DEBUG MODE - %v PRIOR ERRORS! - ABORTING EARLY: "+msg, len(b.errs), from, to)
						}
						b.errs = append(b.errs, fmt.Errorf(msg, from, to))
						return b
					}
					if b.c.debug {
						log.Printf("Found group output connection: fromNodeName=%q, fromOutputName=%q, step.action=%q, step.args=%+v, newFromPortName=%q", fromNodeName, fromOutputName, step.action, step.args, newFromPortName)
					}
					b = b.Connect(newFromPortName, to)
				}
			}
		}
		if connectionsMade > 0 {
			return b
		}

		msg := "Connect(%q, %q) unable to find 'from' node: %q; valid choices are: %+v"
		if b.c.debug {
			log.Fatalf("DEBUG MODE - %v PRIOR ERRORS! - ABORTING EARLY: "+msg, len(b.errs), from, to, fromNodeName, maps.Keys(b.Nodes))
		}

		b.errs = append(b.errs, fmt.Errorf(msg, from, to, fromNodeName, maps.Keys(b.Nodes)))
		return b
	}

	fromOutput, ok := fromNode.GetOutput(fromOutputName)
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q, %q) unable to find 'from' node's output pin: %q; valid choices are: %+v", from, to, fromOutputName, fromNode.GetOutputs()))
		return b
	}

	toParts := strings.Split(to, ".")
	if len(toParts) < 2 {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q, %q): unable to parse 'to' name: %[1]q want at least 2 parts, got %[3]v", from, to, len(toParts)))
		return b
	}

	toNodeName := strings.Join(toParts[0:len(toParts)-1], ".")
	toInputName := toParts[len(toParts)-1]
	toNode, ok := b.Nodes[toNodeName]
	if !ok {
		if b.c.debug {
			log.Printf("Checking groups %+v for '%v'", maps.Keys(b.Groups), toParts[0])
		}

		var connectionsMade int
		if g, ok := b.Groups[toParts[0]]; ok {
			for _, step := range g.groupRecorder {
				if step.action == "Input" && step.args[0] == toInputName {
					connectionsMade++
					newToPortName, _, portName := injectGroupName(step.args[1], toNodeName)
					if portName == "" {
						b.errs = append(b.errs, fmt.Errorf("Input(%q, %q): 'to' node missing port", from, to))
						return b
					}
					b = b.Connect(from, newToPortName)
				}
			}
		}
		if connectionsMade > 0 {
			return b
		}

		b.errs = append(b.errs, fmt.Errorf("Connect(%q, %q) unable to find 'to' node: %q; valid choices are: %+v", from, to, toNodeName, maps.Keys(b.Nodes)))
		return b
	}

	toInput, ok := toNode.GetInput(toInputName)
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q, %q) unable to find 'to' node's input pin: %q; valid choices are: %+v", from, to, toInputName, toNode.GetInputs()))
		return b
	}

	toInput.Kind.External = nil
	toInput.Kind.Connection = &ast.Connection{
		NodeIdx:   fromNode.Index,
		ParamName: fromOutput.Name,
	}

	if v, ok := b.InputsAlreadyConnected[to]; ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q, %q) - 'to' node '%[2]v' already connected OR statically assigned to %q!", from, to, v))
		return b
	}
	b.InputsAlreadyConnected[to] = from

	if toInput.DataType != fromOutput.DataType {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q, %q) - 'from' node type '%v' not compatible with 'to' node type '%v'.", from, to, fromOutput.DataType, toInput.DataType))
	}

	return b
}

func splitArg(arg string) (lhs, rhs string, err error) {
	parts := strings.Split(arg, "=")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("bad arg '%v', want x=y", arg)
	}

	lhs, rhs = strings.TrimSpace(parts[0]), strings.Join(parts[1:], "=") // Do NOT trim rhs! Lose trailing "\n"
	return lhs, rhs, nil
}

func (b *Builder) setInputValues(nodeName string, inputs []*ast.Input, args ...string) ([]*ast.Input, error) {
	var result []*ast.Input

	assignments := map[string]string{}
	for _, arg := range args {
		k, rhs, err := splitArg(arg)
		if err != nil {
			return nil, err
		}

		fullInputName := fmt.Sprintf("%v.%v", nodeName, k)
		if v, ok := b.InputsAlreadyConnected[fullInputName]; ok {
			return nil, fmt.Errorf("input '%v' already assigned to %q!", fullInputName, v)
		}
		v := strings.TrimSpace(rhs)
		b.InputsAlreadyConnected[fullInputName] = v

		assignments[k] = v
		if b.c.debug {
			log.Printf("setting input node '%v' = %v", fullInputName, v)
		}
	}

	validInputNodes := map[string]bool{}
	for _, origInput := range inputs {
		// Make deep copy of inputs
		input := &ast.Input{
			Name:     origInput.Name,
			DataType: origInput.DataType,
			Kind:     ast.DependencyKind{},
			Props:    deepCopyProps(origInput.Props),
		}
		if origInput.Kind.External != nil {
			input.Kind.External = &ast.External{Promoted: origInput.Kind.External.Promoted}
		}

		validInputNodes[input.Name] = true
		if v, ok := assignments[input.Name]; ok {
			if err := setInputProp(input, v); err != nil {
				return nil, err
			}
			result = append(result, input)
			if b.c.debug {
				log.Printf("input node '%v.%v' props=%p", nodeName, input.Name, input.Props)
			}
			continue
		}
		result = append(result, input)
	}

	// Now double-check that all assignments that were made were indeed made to valid input ports.
	for k := range assignments {
		if k == "node_position" { // special setting used to control the drawing of the nodes within blackjack_ui.
			continue
		}
		if !validInputNodes[k] {
			b.errs = append(b.errs, fmt.Errorf("assignment to non-existing input port '%v' on node %q", k, nodeName))
		}
	}

	return result, nil
}

func deepCopyProps(inProps map[string]lua.LValue) map[string]lua.LValue {
	outProps := map[string]lua.LValue{}
	for k, lv := range inProps {
		switch lv.Type() {

		case lua.LTNil, lua.LTBool, lua.LTNumber, lua.LTString:
			outProps[k] = lv

		case lua.LTUserData:
			ud, ok := lv.(*lua.LUserData)
			if !ok {
				log.Fatalf("deepCopyProps: key=%q, expected *LUserData, got %T: %#v", k, lv, lv)
			}

			//2023/10/14 14:14:50 deepCopyProps: unhandled property (k="values") type "table"=&lua.LTable{Metatable:(*lua.LNilType)(0x10487c360), array:[]lua.LValue{"Clockwise", "Counter-Clockwise"}, dict:map[lua.LValue]lua.LValue(nil), strdict:map[string]lua.LValue(nil), keys:[]lua.LValue(nil), k2i:map[lua.LValue]int(nil)}

			var value any
			switch t := ud.Value.(type) {
			case *Vec3:
				value = &Vec3{X: t.X, Y: t.Y, Z: t.Z}
			default:
				log.Fatalf("deepCopyProps: key=%q, expected *Vec3, got %T: %#v", k, t, t)
			}

			outProps[k] = &lua.LUserData{
				Value:     value,
				Env:       ud.Env,
				Metatable: ud.Metatable,
			}

		case lua.LTTable:
			// Copy the table over without a deep copy, as this is typically used for lookup of all
			// possible enum values which are read-only and not going to be modified.
			outProps[k] = lv

		default:
			log.Fatalf("deepCopyProps: unhandled property (k=%q) type %q=%#v", k, lv.Type(), lv)
		}
	}
	return outProps
}

func setInputProp(input *ast.Input, valStr string) error {
	tAny, ok := input.Props["type"]
	if !ok {
		return fmt.Errorf("setInputProp: could not find 'type' for input %q: props=%#v", input.Name, input.Props)
	}
	t, ok := tAny.(lua.LString)
	if !ok {
		return fmt.Errorf("setInputProp: tAny=%T, want lua.LString", tAny)
	}

	switch t {
	case "vec3":
		return setInputVectorValue(t, input, valStr)
	case "scalar":
		return setInputScalarValue(t, input, valStr)
	case "enum":
		return setInputEnumValue(t, input, valStr)
	case "string":
		return setInputStringValue(t, input, valStr)
	default:
		return fmt.Errorf("setInputProp: unknown t=%v, input.Name='%v', props=%#v", t, input.Name, input.Props)
	}
}

func setInputStringValue(t lua.LString, input *ast.Input, valStr string) error {
	if _, ok := input.Props["default"]; !ok {
		return fmt.Errorf("setInputStringValue: t=%v, could not find 'default' for input %q: props=%#v", t, input.Name, input.Props)
	}

	input.Props["default"] = lua.LString(valStr)

	return nil
}

func setInputEnumValue(t lua.LString, input *ast.Input, valStr string) error {
	valuesLVal, ok := input.Props["values"]
	if !ok {
		return fmt.Errorf("setInputEnumValue: t=%v, could not find 'values' for input %q: props=%#v", t, input.Name, input.Props)
	}
	values, ok := valuesLVal.(*lua.LTable)
	if !ok {
		return fmt.Errorf("setInputEnumValue: t=%v, valuesLVal=%T, want *lua.LTable", t, valuesLVal)
	}

	var index int
	var found bool
	var err error
	values.ForEach(func(k, v lua.LValue) {
		if v.String() == valStr {
			found = true
			kv, ok := k.(lua.LNumber)
			if !ok {
				err = fmt.Errorf("setInputEnumValue: t=%v, input.Name='%v', v='%v', k=%v (%T) want lua.LNumber", t, input.Name, v, k, k)
				return
			}
			index = int(kv) - 1
		}
	})
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("setInputEnumValue: t=%v, input.Name='%v', props=%#v, values=%#v: enum '%v' not found", t, input.Name, input.Props, values, valStr)
	}

	input.Props["selected"] = lua.LNumber(index)

	return nil
}

func setInputScalarValue(t lua.LString, input *ast.Input, valStr string) error {
	if _, ok := input.Props["default"]; !ok {
		return fmt.Errorf("setInputScalarValue: t=%v, could not find 'default' for input %q: props=%#v", t, input.Name, input.Props)
	}

	x, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return fmt.Errorf("setInputScalarValue: t=%v, input=%q, unable to parse value: '%v'", t, input.Name, valStr)
	}

	if minLVal, ok := input.Props["min"]; ok {
		min, ok := minLVal.(lua.LNumber)
		if !ok {
			return fmt.Errorf("setInputScalarValue: t=%v, input=%q, min=%T, expected LNumber", t, input.Name, minLVal)
		}
		if x < float64(min) {
			return fmt.Errorf("setInputScalarValue: t=%v, input=%q, attempt to set scalar (%v) < min (%v)", t, input.Name, x, min)
		}
	}

	input.Props["default"] = lua.LNumber(x)

	return nil
}

func setInputVectorValue(t lua.LString, input *ast.Input, valStr string) error {
	defLVal, ok := input.Props["default"]
	if !ok {
		return fmt.Errorf("setInputVectorValue: t=%v, could not find 'default' for input %q: props=%#v", t, input.Name, input.Props)
	}
	defVal, ok := defLVal.(*lua.LUserData)
	if !ok {
		return fmt.Errorf("setInputVectorValue: t=%v, defLVal=%T, want *ast.LUserData", t, defLVal)
	}

	const prefix = "vector("
	if !strings.HasPrefix(valStr, prefix) || valStr[len(valStr)-1:] != ")" {
		return fmt.Errorf("setInputVectorValue: t=%v, input=%q, want vector(x,y,z), got %v", t, input.Name, valStr)
	}
	valStr = strings.TrimPrefix(valStr[:len(valStr)-1], prefix)
	parts := strings.Split(valStr, ",")
	if len(parts) != 3 {
		return fmt.Errorf("setInputVectorValue: t=%v, input=%q, want vector(x,y,z), got %v", t, input.Name, valStr)
	}
	xStr := strings.TrimSpace(parts[0])
	yStr := strings.TrimSpace(parts[1])
	zStr := strings.TrimSpace(parts[2])
	x, err := strconv.ParseFloat(xStr, 64)
	if err != nil {
		return fmt.Errorf("setInputVectorValue: t=%v, input=%q, unable to parse X value: '%v'", t, input.Name, xStr)
	}
	y, err := strconv.ParseFloat(yStr, 64)
	if err != nil {
		return fmt.Errorf("setInputVectorValue: t=%v, input=%q, unable to parse Y value: '%v'", t, input.Name, yStr)
	}
	z, err := strconv.ParseFloat(zStr, 64)
	if err != nil {
		return fmt.Errorf("setInputVectorValue: t=%v, input=%q, unable to parse Z value: '%v'", t, input.Name, zStr)
	}

	defVal.Value = &Vec3{X: x, Y: y, Z: z}

	return nil
}

// MergeMesh creates a new 'MergeMeshes' node if necessary to combine the last
// mesh with this current mesh. Typically, a design will end with a `MergeMesh`
// before the call to `Build` such that it is the last (and therefore "active")
// node in the graph network.
func (b *Builder) MergeMesh(name string) *Builder {
	if b.lastMergeMesh == "" {
		b.lastMergeMesh = name
		return b
	}

	newNode := fmt.Sprintf("MergeMeshes.%v", len(b.Nodes))
	b = b.
		AddNode(newNode).
		Connect(b.lastMergeMesh, newNode+".mesh_a").
		Connect(name, newNode+".mesh_b")
	b.lastMergeMesh = newNode + ".out_mesh"
	return b
}

// Builder builds the design and returns the result.
func (b *Builder) Build() (*ast.BJK, error) {
	if b.CheckUnusedGroupInputs {
		for k := range b.groupFullInputPortNames {
			if _, ok := b.InputsAlreadyConnected[k]; !ok {
				b.errs = append(b.errs, fmt.Errorf("unused group input port: %q", k))
			}
		}
	}

	if len(b.errs) > 0 {
		if b.c.debug || len(b.errs) <= 5 {
			return nil, fmt.Errorf("%v ERRORS FOUND:\n%w", len(b.errs), errors.Join(b.errs...))
		}
		return nil, fmt.Errorf("%v ERRORS FOUND - HERE ARE THE FIRST 5:\n%w", len(b.errs), errors.Join(b.errs[:5]...))
	}

	bjk := ast.New()
	if len(b.NodeOrder) == 0 {
		return bjk, nil
	}

	g := bjk.Graph
	ep := &b.ExternalParameters
	addPV := func(pv *ast.ParamValue) { ep.ParamValues = append(ep.ParamValues, pv) }

	g.UIData.NodePositions = make([]*ast.Vec2, len(b.NodeOrder))
	g.UIData.NodeOrder = make([]uint64, len(b.NodeOrder))

	lastXOffset, lastYOffset := float64(nodeXOffset), float64(-nodeYOffset)
	for i, k := range b.NodeOrder {
		lastXOffset += nodeXOffset
		lastYOffset += nodeYOffset
		g.UIData.NodePositions[i] = &ast.Vec2{X: lastXOffset, Y: lastYOffset}
		g.UIData.NodeOrder[i] = uint64(i)

		node, ok := b.Nodes[k]
		if !ok {
			return nil, fmt.Errorf("programming error: missing node '%v'", k)
		}
		g.Nodes = append(g.Nodes, node)

		if node.NodePosition != nil {
			g.UIData.NodePositions[i] = node.NodePosition
			lastXOffset = node.NodePosition.X
			lastYOffset = node.NodePosition.Y
		}

		// For each unconnected input, add an ExternalParameters value.
		for _, input := range node.Inputs {
			if input.Kind.Connection != nil {
				continue
			}

			ve, err := getValueEnum(input)
			if err != nil {
				return nil, fmt.Errorf("Build: node '%v': %w", k, err)
			}

			addPV(&ast.ParamValue{
				NodeIdx:   node.Index,
				ParamName: input.Name,
				ValueEnum: *ve,
			})
		}
	}

	dn := uint64(len(b.NodeOrder) - 1)
	g.DefaultNode = &dn
	g.ExternalParameters = ep

	return bjk, nil
}

func getValueEnum(input *ast.Input) (*ast.ValueEnum, error) {
	tAny, ok := input.Props["type"]
	if !ok {
		return nil, fmt.Errorf("getValueEnum: could not find 'type' for input %q: props=%#v", input.Name, input.Props)
	}
	t, ok := tAny.(lua.LString)
	if !ok {
		return nil, fmt.Errorf("getValueEnum: tAny=%T, want lua.LString", tAny)
	}

	switch t {
	case "vec3":
		return getVectorValue(t, input)
	case "scalar":
		return getScalarValue(t, input)
	case "enum":
		return getEnumValue(t, input)
	case "string":
		return getStringValue(t, input)
	case "file", "lua_string":
		log.Printf("getValueEnum: WARNING: value of type '%v' not supported yet.", t)
		return &ast.ValueEnum{StrVal: &ast.StringValue{S: "TODO"}}, nil
	case "selection":
		log.Printf("getValueEnum: WARNING: value of type '%v' not supported yet.", t)
		return &ast.ValueEnum{Selection: &ast.SelectionValue{Selection: "TODO"}}, nil
	case "mesh":
		return nil, fmt.Errorf("unconnected input '%v' of type 'mesh'", input.Name)
	default:
		return nil, fmt.Errorf("getValueEnum: unknown t=%v, input.Name='%v', props=%#v", t, input.Name, input.Props)
	}
}

func getStringValue(t lua.LString, input *ast.Input) (*ast.ValueEnum, error) {
	defLVal, ok := input.Props["default"]
	if !ok {
		return nil, fmt.Errorf("getStringValue: t=%v, could not find 'default' for input %q: props=%#v", t, input.Name, input.Props)
	}
	val, ok := defLVal.(lua.LString)
	if !ok {
		return nil, fmt.Errorf("getStringValue: defVal.Value=%T, want lua.LString", defLVal)
	}

	return &ast.ValueEnum{
		StrVal: &ast.StringValue{S: string(val)},
	}, nil
}

func getScalarValue(t lua.LString, input *ast.Input) (*ast.ValueEnum, error) {
	defLVal, ok := input.Props["default"]
	if !ok {
		return nil, fmt.Errorf("getScalarValue: t=%v, could not find 'default' for input %q: props=%#v", t, input.Name, input.Props)
	}
	val, ok := defLVal.(lua.LNumber)
	if !ok {
		return nil, fmt.Errorf("getScalarValue: defVal.Value=%T, want lua.LNumber", defLVal)
	}

	return &ast.ValueEnum{
		Scalar: &ast.ScalarValue{X: float64(val)},
	}, nil
}

func getEnumValue(t lua.LString, input *ast.Input) (*ast.ValueEnum, error) {
	valuesLVal, ok := input.Props["values"]
	if !ok {
		return nil, fmt.Errorf("getEnumValue: t=%v, could not find 'values' for input %q: props=%#v", t, input.Name, input.Props)
	}
	values, ok := valuesLVal.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("getEnumValue: t=%v, valuesLVal=%T, want *lua.LTable", t, valuesLVal)
	}

	var selected int // 'selected' field is optional - default to 0 for Blackjack - for example, see: EditGeometry
	selectedLVal, ok := input.Props["selected"]
	if ok {
		selectedLNum, ok := selectedLVal.(lua.LNumber)
		if !ok {
			return nil, fmt.Errorf("getEnumValue: t=%v, input %q: selectedLVal=%T, want lua.LNumber", t, input.Name, selectedLVal)
		}
		selected = int(selectedLNum)
	}

	val, ok := values.RawGetInt(selected + 1).(lua.LString) // Lua is 1-indexed, but Blackjack is 0-indexed!
	if !ok {
		return nil, fmt.Errorf("getEnumValue: t=%v, values.RawGetInt(%v)=%T, want string", t, selected, values.RawGetInt(selected))
	}

	return &ast.ValueEnum{
		StrVal: &ast.StringValue{S: string(val)},
	}, nil
}

func getVectorValue(t lua.LString, input *ast.Input) (*ast.ValueEnum, error) {
	defLVal, ok := input.Props["default"]
	if !ok {
		return nil, fmt.Errorf("getValueEnum: t=%v, could not find 'default' for input %q: props=%#v", t, input.Name, input.Props)
	}
	defVal, ok := defLVal.(*lua.LUserData)
	if !ok {
		return nil, fmt.Errorf("getValueEnum: t=%v, defLVal=%T, want *ast.LUserData", t, defLVal)
	}
	val, ok := defVal.Value.(*Vec3)
	if !ok {
		return nil, fmt.Errorf("getValueEnum: defVal.Value=%T, want *Vec3", defVal.Value)
	}

	return &ast.ValueEnum{
		Vector: &ast.VectorValue{X: val.X, Y: val.Y, Z: val.Z},
	}, nil
}

func parseVec2(s string) (*ast.Vec2, bool) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "(") || !strings.HasSuffix(s, ")") {
		return nil, false
	}
	parts := strings.Split(s[1:len(s)-1], ",")
	if len(parts) != 2 {
		return nil, false
	}
	xStr := strings.TrimSpace(parts[0])
	x, err := strconv.ParseFloat(xStr, 64)
	if err != nil {
		return nil, false
	}
	yStr := strings.TrimSpace(parts[1])
	y, err := strconv.ParseFloat(yStr, 64)
	if err != nil {
		return nil, false
	}
	return &ast.Vec2{X: x, Y: y}, true
}
