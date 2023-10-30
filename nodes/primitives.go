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
	"cube":              cube,
	"line_with_normals": lineWithNormals,
	"quad":              quad,
}

func cube(ls *lua.LState) int {
	// log.Printf("cube called!")
	if ls.GetTop() != 2 {
		log.Fatalf("cube: GetTop=%v, want 2", ls.GetTop())
	}

	center := checkVec3(ls, 1)
	if center == nil {
		log.Fatalf("cube: center=%q, want Vec3", ls.Get(1).Type())
	}

	size := checkVec3(ls, 2)
	if size == nil {
		log.Fatalf("cube: size=%q, want Vec3", ls.Get(2).Type())
	}

	// log.Printf("cube: center=%#v, size=%#v", center, size)

	halfSize := size.MulScalar(0.5)

	v1 := center.Add(NewVec3(-halfSize.X, -halfSize.Y, -halfSize.Z))
	v2 := center.Add(NewVec3(halfSize.X, -halfSize.Y, -halfSize.Z))
	v3 := center.Add(NewVec3(halfSize.X, -halfSize.Y, halfSize.Z))
	v4 := center.Add(NewVec3(-halfSize.X, -halfSize.Y, halfSize.Z))

	v5 := center.Add(NewVec3(-halfSize.X, halfSize.Y, -halfSize.Z))
	v6 := center.Add(NewVec3(-halfSize.X, halfSize.Y, halfSize.Z))
	v7 := center.Add(NewVec3(halfSize.X, halfSize.Y, halfSize.Z))
	v8 := center.Add(NewVec3(halfSize.X, halfSize.Y, -halfSize.Z))

	polygon := NewMeshFromPolygons(
		[]Vec3{v1, v2, v3, v4, v5, v6, v7, v8},
		[][]int{
			{0, 1, 2, 3},
			{4, 5, 6, 7},
			{4, 7, 1, 0},
			{3, 2, 6, 5},
			{5, 4, 0, 3},
			{6, 2, 1, 7},
		})

	ud := ls.NewUserData()
	ud.Value = polygon
	ls.SetMetatable(ud, ls.GetTypeMetatable(luaMeshTypeName))
	ls.Push(ud)
	return 1
}

func lineWithNormals(ls *lua.LState) int {
	pointsTbl := ls.CheckTable(1)
	normalsTbl := ls.CheckTable(2)
	tangentsTbl := ls.CheckTable(3)
	numSegments := int(ls.CheckNumber(4))

	points := make([]Vec3, 0, numSegments)
	normals := make([]Vec3, 0, numSegments)
	tangents := make([]Vec3, 0, numSegments)

	getVec3 := func(tbl *lua.LTable, index int) *Vec3 {
		ud, ok := tbl.RawGetInt(index).(*lua.LUserData)
		if !ok {
			log.Fatalf("lineWithNormals: tbl[i=%v], want vec3, got %T", index, tbl.RawGetInt(index))
		}
		v, ok := ud.Value.(*Vec3)
		if !ok {
			log.Fatalf("lineWithNormals: tbl[i=%v], want vec3, got %T", index, ud.Value)
		}
		return v
	}

	for i := 1; i <= numSegments; i++ {
		v := getVec3(pointsTbl, i)
		points = append(points, *v)
		v = getVec3(normalsTbl, i)
		normals = append(normals, *v)
		v = getVec3(tangentsTbl, i)
		normals = append(normals, *v)
	}

	mesh := NewMeshFromLineWithNormals(points, normals, tangents)

	ud := ls.NewUserData()
	ud.Value = mesh
	ls.SetMetatable(ud, ls.GetTypeMetatable(luaMeshTypeName))
	ls.Push(ud)
	return 1
}

func quad(ls *lua.LState) int {
	if ls.GetTop() != 4 {
		log.Fatalf("quad: GetTop=%v, want 2", ls.GetTop())
	}

	center := checkVec3(ls, 1)
	if center == nil {
		log.Fatalf("quad: center=%q, want Vec3", ls.Get(1).Type())
	}

	normal := checkVec3(ls, 2)
	if normal == nil {
		log.Fatalf("quad: normal=%q, want Vec3", ls.Get(2).Type())
	}

	right := checkVec3(ls, 3)
	if right == nil {
		log.Fatalf("quad: right=%q, want Vec3", ls.Get(3).Type())
	}

	size := checkVec3(ls, 4)
	if size == nil {
		log.Fatalf("quad: size=%q, want Vec3", ls.Get(4).Type())
	}

	normal.Normalize()
	right.Normalize()
	forward := normal.Cross(right)

	halfSize := size.MulScalar(0.5)
	scaledRight := right.MulScalar(halfSize.X)
	scaledForward := forward.MulScalar(halfSize.Y)

	v1 := center.Add(scaledRight.Add(scaledForward))
	v2 := center.Sub(scaledRight.Add(scaledForward))
	v3 := center.Sub(scaledRight.Sub(scaledForward))
	v4 := center.Add(scaledRight.Sub(scaledForward))

	polygon := NewMeshFromPolygons(
		[]Vec3{v1, v2, v3, v4},
		[][]int{
			{0, 1, 2, 3},
		})

	ud := ls.NewUserData()
	ud.Value = polygon
	ls.SetMetatable(ud, ls.GetTypeMetatable(luaMeshTypeName))
	ls.Push(ud)
	return 1
}
