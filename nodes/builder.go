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

// Builder represents a BJK builder.
type Builder struct {
	c    *Client
	errs []error

	isGroup       bool
	groupRecorder []*recorder

	Nodes     map[string]*ast.Node
	NodeOrder []string
	Groups    map[string]*Builder

	ExternalParameters ast.ExternalParameters

	InputsAlreadyConnected map[string]bool
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
		InputsAlreadyConnected: map[string]bool{},
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

func (b *Builder) instantiateGroup(groupName string, group *Builder, args ...string) *Builder {
	if b.c.debug {
		log.Printf("Instantiating group '%v' with %v steps", groupName, len(group.groupRecorder))
	}

	injectGroupName := func(fullPortName string) string {
		parts := strings.Split(fullPortName, ".")
		baseName, portName := strings.Join(parts[0:len(parts)-1], "."), parts[len(parts)-1]
		return fmt.Sprintf("%v.%v.%v", baseName, groupName, portName)
	}

	for i, step := range group.groupRecorder {
		if b.c.debug {
			log.Printf("Group '%v' step #%v of %v: %v('%v') ...", groupName, i+1, len(group.groupRecorder), step.action, strings.Join(step.args, "', '"))
		}

		switch step.action {
		case "AddNode":
			fullNodeName := fmt.Sprintf("%v.%v", step.args[0], groupName)
			b = b.AddNode(fullNodeName, step.args[1:]...)
		case "Connect":
			b = b.Connect(injectGroupName(step.args[0]), injectGroupName(step.args[1]))
		case "Input":
			b = b.Input(step.args[0], injectGroupName(step.args[1]))
		case "Output":
			b = b.Output(injectGroupName(step.args[0]), step.args[1])
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

	b.Nodes[name] = &ast.Node{
		OpName:      nodeType,
		ReturnValue: n.ReturnValue, // OK not to make a deep copy of ReturnValue - it doesn't change.
		Inputs:      inputs,
		Outputs:     outputs,

		Label: n.Label,
		Index: uint64(len(b.NodeOrder)), // 0-based indices
	}
	b.NodeOrder = append(b.NodeOrder, name)

	return b
}

// Connect connects the `from` node.output_port to the `to` node.input_port.
func (b *Builder) Connect(from, to string) *Builder {
	if b.isGroup {
		b.groupRecorder = append(b.groupRecorder, &recorder{
			action: "Connect",
			args:   []string{from, to},
		})
		return b
	}

	if b.InputsAlreadyConnected[to] {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) - 'to' node '%[2]v' already connected!", from, to))
	}
	b.InputsAlreadyConnected[to] = true

	fromParts := strings.Split(from, ".")
	if len(fromParts) < 2 {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q): unable to parse 'from' name: %[1]q want at least 2 parts, got %v", from, to, len(fromParts)))
		return b
	}

	fromNodeName := strings.Join(fromParts[0:len(fromParts)-1], ".")
	fromNode, ok := b.Nodes[fromNodeName]
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) unable to find 'from' node: %q; valid choices are: %+v", from, to, fromNodeName, maps.Keys(b.Nodes)))
		return b
	}

	fromOutputName := fromParts[len(fromParts)-1]
	fromOutput, ok := fromNode.GetOutput(fromOutputName)
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) unable to find 'from' node's output pin: %q; valid choices are: %+v", from, to, fromOutputName, fromNode.GetOutputs()))
		return b
	}

	toParts := strings.Split(to, ".")
	if len(toParts) < 2 {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q): unable to parse 'to' name: %[1]q want at least 2 parts, got %v", from, to, len(toParts)))
		return b
	}

	toNodeName := strings.Join(toParts[0:len(toParts)-1], ".")
	toNode, ok := b.Nodes[toNodeName]
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) unable to find 'to' node: %q; valid choices are: %+v", from, to, toNodeName, maps.Keys(b.Nodes)))
		return b
	}

	toInputName := toParts[len(toParts)-1]
	toInput, ok := toNode.GetInput(toInputName)
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) unable to find 'to' node's input pin: %q; valid choices are: %+v", from, to, toInputName, toNode.GetInputs()))
		return b
	}

	// if b.c.debug {
	// 	log.Printf("Connect(%q,%q):\nfrom:\n%#v\nto:\n%#v", from, to, valast.String(fromOutput), valast.String(toInput))
	// }

	toInput.Kind.External = nil
	toInput.Kind.Connection = &ast.Connection{
		NodeIdx:   fromNode.Index,
		ParamName: fromOutput.Name,
	}

	return b
}

func (b *Builder) setInputValues(nodeName string, inputs []*ast.Input, args ...string) ([]*ast.Input, error) {
	var result []*ast.Input

	assignments := map[string]string{}
	for _, arg := range args {
		parts := strings.Split(arg, "=")
		if len(parts) < 2 {
			return nil, fmt.Errorf("bad arg '%v', want x=y", arg)
		}

		lhs, rhs := parts[0], strings.Join(parts[1:], "=")
		k := strings.TrimSpace(lhs)

		fullInputName := fmt.Sprintf("%v.%v", nodeName, k)
		if b.InputsAlreadyConnected[fullInputName] {
			return nil, fmt.Errorf("input '%v' already assigned", fullInputName)
		}
		b.InputsAlreadyConnected[fullInputName] = true

		v := strings.TrimSpace(rhs)
		assignments[k] = v
		if b.c.debug {
			log.Printf("setting input node '%v' = %v", fullInputName, v)
		}
	}

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

	return result, nil
}

func deepCopyProps(inProps map[string]any) map[string]any {
	outProps := map[string]any{}
	for k, v := range inProps {
		lv, ok := v.(lua.LValue)
		if !ok {
			log.Fatalf("deepCopyProps: key=%q, expected LValue, got %T: %#v", k, v, v)
		}
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
	default:
		return fmt.Errorf("setInputProp: unknown t=%v, input.Name='%v', props=%#v", t, input.Name, input.Props)
	}
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

// Builder builds the design and returns the result.
func (b *Builder) Build() (*ast.BJK, error) {
	if len(b.errs) > 0 {
		return nil, fmt.Errorf("%v ERRORS FOUND:\n%w", len(b.errs), errors.Join(b.errs...))
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

	for i, k := range b.NodeOrder {
		g.UIData.NodePositions[i] = &ast.Vec2{X: float64(nodeXOffset * i), Y: float64(nodeYOffset * i)}
		g.UIData.NodeOrder[i] = uint64(i)

		node, ok := b.Nodes[k]
		if !ok {
			return nil, fmt.Errorf("programming error: missing node '%v'", k)
		}
		g.Nodes = append(g.Nodes, node)
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
	case "mesh":
		return nil, fmt.Errorf("unconnected input '%v' of type 'mesh'", input.Name)
	default:
		return nil, fmt.Errorf("getValueEnum: unknown t=%v, input.Name='%v', props=%#v", t, input.Name, input.Props)
	}
}

func getScalarValue(t lua.LString, input *ast.Input) (*ast.ValueEnum, error) {
	defLVal, ok := input.Props["default"]
	if !ok {
		return nil, fmt.Errorf("getScalarValue: t=%v, could not find 'default' for input %q: props=%#v", t, input.Name, input.Props)
	}
	val, ok := defLVal.(lua.LNumber)
	if !ok {
		return nil, fmt.Errorf("getScalarValue: defVal.Value=%T, want *Vec3", defLVal)
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

	selectedLVal, ok := input.Props["selected"]
	if !ok {
		return nil, fmt.Errorf("getEnumValue: t=%v, could not find 'selected' for input %q: props=%#v", t, input.Name, input.Props)
	}
	selectedLNum, ok := selectedLVal.(lua.LNumber)
	if !ok {
		return nil, fmt.Errorf("getEnumValue: t=%v, input %q: selectedLVal=%T, want lua.LNumber", t, input.Name, selectedLVal)
	}
	selected := int(selectedLNum)

	val, ok := values.RawGetInt(selected + 1).(lua.LString) // 1-indexed
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
