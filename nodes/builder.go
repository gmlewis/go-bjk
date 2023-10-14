package nodes

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gmlewis/go-bjk/ast"
	lua "github.com/yuin/gopher-lua"
)

// Builder represents a BJK builder.
type Builder struct {
	c    *Client
	errs []error

	Nodes     map[string]*ast.Node
	NodeOrder []string

	ExternalParameters ast.ExternalParameters

	InputsAlreadyConnected map[string]bool
}

// NewBuilder returns a new BJK Builder.
func (c *Client) NewBuilder() *Builder {
	return &Builder{
		c:                      c,
		Nodes:                  map[string]*ast.Node{},
		InputsAlreadyConnected: map[string]bool{},
	}
}

// AddNode adds a new node to the design with the optional args.
func (b *Builder) AddNode(name string, args ...string) *Builder {
	if name == "" {
		b.errs = append(b.errs, errors.New("AddNode: name cannot be empty"))
		return b
	}

	parts := strings.Split(name, ".")
	n, ok := b.c.Nodes[parts[0]]
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("AddNode: unknown node type '%v'", parts[0]))
		return b
	}

	if _, ok := b.Nodes[name]; ok {
		b.errs = append(b.errs, fmt.Errorf("AddNode: node '%v' already exists", name))
		return b
	}

	inputs, err := b.setInputValues(name, n.Inputs, args...)
	if err != nil {
		b.errs = append(b.errs, fmt.Errorf("setInputValues: %v", err))
		return b
	}

	b.Nodes[name] = &ast.Node{
		OpName:      parts[0],
		ReturnValue: n.ReturnValue,
		Inputs:      inputs,
		Outputs:     n.Outputs,

		Label: n.Label,
		Index: uint64(len(b.NodeOrder)), // 0-based indices
	}
	b.NodeOrder = append(b.NodeOrder, name)

	return b
}

// Connect connects the `from` node.output to the `to` node.input.
func (b *Builder) Connect(from, to string) *Builder {
	if b.InputsAlreadyConnected[to] {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) - 'to' node '%[2]v' already connected!", from, to))
	}
	b.InputsAlreadyConnected[to] = true

	fromParts := strings.Split(from, ".")
	if len(fromParts) != 3 {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q): unable to parse 'from' name: %[1]q want 3 parts, got %v", from, to, len(fromParts)))
		return b
	}

	fromNodeName := fmt.Sprintf("%v.%v", fromParts[0], fromParts[1])
	fromNode, ok := b.Nodes[fromNodeName]
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) unable to find 'from' node: %q", from, to, fromNodeName))
		return b
	}

	fromOutput, ok := fromNode.GetOutput(fromParts[2])
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) unable to find 'from' node output: %q", from, to, fromParts[2]))
		return b
	}

	toParts := strings.Split(to, ".")
	if len(toParts) != 3 {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q): unable to parse 'to' name: %[1]q want 3 parts, got %v", from, to, len(toParts)))
		return b
	}

	toNodeName := fmt.Sprintf("%v.%v", toParts[0], toParts[1])
	toNode, ok := b.Nodes[toNodeName]
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) unable to find 'to' node: %q", from, to, toNodeName))
		return b
	}

	toInput, ok := toNode.GetInput(toParts[2])
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) unable to find 'to' node input: %q", from, to, toParts[2]))
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
		if len(parts) != 2 {
			return nil, fmt.Errorf("unable to parse arg '%v', want x=y", arg)
		}
		k := strings.TrimSpace(parts[0])

		fullInputName := fmt.Sprintf("%v.%v", nodeName, k)
		if b.InputsAlreadyConnected[fullInputName] {
			return nil, fmt.Errorf("input '%v' already assigned", fullInputName)
		}
		b.InputsAlreadyConnected[fullInputName] = true

		v := strings.TrimSpace(parts[1])
		assignments[k] = v
		if b.c.debug {
			log.Printf("setting input node '%v' = %v", fullInputName, v)
		}
	}

	for _, input := range inputs {
		if v, ok := assignments[input.Name]; ok {
			in, err := setInputProp(input, v)
			if err != nil {
				return nil, err
			}
			result = append(result, in)
			continue
		}
		result = append(result, &ast.Input{
			Name:     input.Name,
			DataType: input.DataType,
			Kind:     input.Kind,
			Props:    input.Props,
		})
	}

	return result, nil
}

func setInputProp(input *ast.Input, valStr string) (*ast.Input, error) {
	tAny, ok := input.Props["type"]
	if !ok {
		return nil, fmt.Errorf("setInputProp: could not find 'type' for input %q: props=%#v", input.Name, input.Props)
	}
	t, ok := tAny.(lua.LString)
	if !ok {
		return nil, fmt.Errorf("setInputProp: tAny=%T, want lua.LString", tAny)
	}

	switch t {
	case "vec3":
		return setInputVectorValue(t, input, valStr)
	case "scalar":
		return setInputScalarValue(t, input, valStr)
	case "enum":
		return setInputEnumValue(t, input, valStr)
	default:
		return nil, fmt.Errorf("setInputProp: unknown t=%v, input.Name='%v', props=%#v", t, input.Name, input.Props)
	}
}

func setInputEnumValue(t lua.LString, input *ast.Input, valStr string) (*ast.Input, error) {
	valuesLVal, ok := input.Props["values"]
	if !ok {
		return nil, fmt.Errorf("setInputEnumValue: t=%v, could not find 'values' for input %q: props=%#v", t, input.Name, input.Props)
	}
	values, ok := valuesLVal.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("setInputEnumValue: t=%v, valuesLVal=%T, want *lua.LTable", t, valuesLVal)
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
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("setInputEnumValue: t=%v, input.Name='%v', props=%#v, values=%#v: enum '%v' not found", t, input.Name, input.Props, values, valStr)
	}

	input.Props["selected"] = lua.LNumber(index)

	return &ast.Input{
		Name:     input.Name,
		DataType: input.DataType,
		Kind:     input.Kind,
		Props:    input.Props,
	}, nil
}

func setInputScalarValue(t lua.LString, input *ast.Input, valStr string) (*ast.Input, error) {
	if _, ok := input.Props["default"]; !ok {
		return nil, fmt.Errorf("setInputScalarValue: t=%v, could not find 'default' for input %q: props=%#v", t, input.Name, input.Props)
	}

	x, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return nil, fmt.Errorf("setInputScalarValue: t=%v, input=%q, unable to parse value: '%v'", t, input.Name, valStr)
	}

	if minLVal, ok := input.Props["min"]; ok {
		min, ok := minLVal.(lua.LNumber)
		if !ok {
			return nil, fmt.Errorf("setInputScalarValue: t=%v, input=%q, min=%T, expected LNumber", t, input.Name, minLVal)
		}
		if x < float64(min) {
			return nil, fmt.Errorf("setInputScalarValue: t=%v, input=%q, attempt to set scalar (%v) < min (%v)", t, input.Name, x, min)
		}
	}

	log.Printf("BEFORE: (x=%v) - setInputScalarValue: t=%v, input=%q, props=%#v", x, t, input.Name, input.Props)
	input.Props["default"] = lua.LNumber(x)
	log.Printf("AFTER: (x=%v) - setInputScalarValue: t=%v, input=%q, props=%#v", x, t, input.Name, input.Props)

	return &ast.Input{
		Name:     input.Name,
		DataType: input.DataType,
		Kind:     input.Kind,
		Props:    input.Props,
	}, nil
}

func setInputVectorValue(t lua.LString, input *ast.Input, valStr string) (*ast.Input, error) {
	defLVal, ok := input.Props["default"]
	if !ok {
		return nil, fmt.Errorf("setInputVectorValue: t=%v, could not find 'default' for input %q: props=%#v", t, input.Name, input.Props)
	}
	defVal, ok := defLVal.(*lua.LUserData)
	if !ok {
		return nil, fmt.Errorf("setInputVectorValue: t=%v, defLVal=%T, want *ast.LUserData", t, defLVal)
	}

	const prefix = "vector("
	if !strings.HasPrefix(valStr, prefix) || valStr[len(valStr)-1:] != ")" {
		return nil, fmt.Errorf("setInputVectorValue: t=%v, input=%q, want vector(x,y,z), got %v", t, input.Name, valStr)
	}
	valStr = strings.TrimPrefix(valStr[:len(valStr)-1], prefix)
	parts := strings.Split(valStr, ",")
	if len(parts) != 3 {
		return nil, fmt.Errorf("setInputVectorValue: t=%v, input=%q, want vector(x,y,z), got %v", t, input.Name, valStr)
	}
	xStr := strings.TrimSpace(parts[0])
	yStr := strings.TrimSpace(parts[1])
	zStr := strings.TrimSpace(parts[2])
	x, err := strconv.ParseFloat(xStr, 64)
	if err != nil {
		return nil, fmt.Errorf("setInputVectorValue: t=%v, input=%q, unable to parse X value: '%v'", t, input.Name, xStr)
	}
	y, err := strconv.ParseFloat(yStr, 64)
	if err != nil {
		return nil, fmt.Errorf("setInputVectorValue: t=%v, input=%q, unable to parse Y value: '%v'", t, input.Name, yStr)
	}
	z, err := strconv.ParseFloat(zStr, 64)
	if err != nil {
		return nil, fmt.Errorf("setInputVectorValue: t=%v, input=%q, unable to parse Z value: '%v'", t, input.Name, zStr)
	}

	defVal.Value = &Vec3{X: x, Y: y, Z: z}

	return &ast.Input{
		Name:     input.Name,
		DataType: input.DataType,
		Kind:     input.Kind,
		Props:    input.Props,
	}, nil
}

// Builder builds the design and returns the result.
func (b *Builder) Build() (*ast.BJK, error) {
	if len(b.errs) > 0 {
		return nil, errors.Join(b.errs...)
	}

	bjk := ast.New()
	if len(b.NodeOrder) == 0 {
		return bjk, nil
	}

	g := bjk.Graph
	ep := &b.ExternalParameters
	addPV := func(pv *ast.ParamValue) { ep.ParamValues = append(ep.ParamValues, pv) }

	for _, k := range b.NodeOrder {
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

	log.Printf("GETTING SCALAR: t=%v, input=%q, props=%#v, val=%v", t, input.Name, input.Props, val)

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
