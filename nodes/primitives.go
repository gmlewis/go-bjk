package nodes

import (
	"log"

	lua "github.com/yuin/gopher-lua"
)

const luaPrimitivesTypeName = "Primitives"

func registerPrimitivesType(ls *lua.LState) {
	mt := ls.NewTypeMetatable(luaPrimitivesTypeName)
	ls.SetGlobal(luaPrimitivesTypeName, mt)
	ls.SetField(mt, "__index", ls.SetFuncs(ls.NewTable(), primitivesMethods))
	for name, fn := range primitivesMethods {
		ls.SetField(mt, name, ls.NewFunction(fn))
	}
}

// // Checks whether the first lua argument is a *LUserData with *Primitives and returns this *Primitives.
// func checkPrimitives(ls *lua.LState) *Primitives {
// 	ud := ls.CheckUserData(1)
// 	if v, ok := ud.Value.(*Primitives); ok {
// 		return v
// 	}
// 	ls.ArgError(1, "primitives expected")
// 	return nil
// }

var primitivesMethods = map[string]lua.LGFunction{
	"cube": cube,
}

func cube(ls *lua.LState) int {
	log.Printf("cube called!")
	if ls.GetTop() != 2 {
		log.Fatalf("cube: GetTop=%v, want 2", ls.GetTop())
	}

	ud := ls.CheckUserData(1)
	center, ok := ud.Value.(*Vec3)
	if !ok {
		log.Fatalf("cube: center=%T, want Vec3", ud.Value)
	}

	ud = ls.CheckUserData(2)
	size, ok := ud.Value.(*Vec3)
	if !ok {
		log.Fatalf("cube: size=%T, want Vec3", ud.Value)
	}

	log.Printf("cube: center=%#v, size=%#v", center, size)

	// if ls.GetTop() == 2 {
	// 	center :=
	// 		size :=
	// }
	return 0
}
