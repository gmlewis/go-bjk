package nodes

import lua "github.com/yuin/gopher-lua"

const luaPrimitivesTypeName = "Primitives"

func registerPrimitivesType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaPrimitivesTypeName)
	L.SetGlobal(luaPrimitivesTypeName, mt)
	// static attributes
	// L.SetField(mt, "new", L.NewFunction(newPrimitives))
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), primitivesMethods))
}

// // Checks whether the first lua argument is a *LUserData with *Primitives and returns this *Primitives.
// func checkPrimitives(L *lua.LState) *Primitives {
// 	ud := L.CheckUserData(1)
// 	if v, ok := ud.Value.(*Primitives); ok {
// 		return v
// 	}
// 	L.ArgError(1, "primitives expected")
// 	return nil
// }

var primitivesMethods = map[string]lua.LGFunction{
	"cube": cube,
}

func cube(L *lua.LState) int {
	// if L.GetTop() == 2 {
	// 	center :=
	// 		size :=
	// }
	return 0
}
