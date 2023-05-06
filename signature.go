package simple_http

import (
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

// Signature
//
//	@param map[string]interface{}
//	@param map[string]string
//	@param map[string]interface{}
//	@return map[string]interface{}
//	@return map[string]string
type Signature func(map[string]interface{}, map[string]string, map[string]interface{}) (map[string]interface{}, map[string]string, error)

var SignatureList map[string]Signature

// GetSignature
//
//	@param name
//	@return Signature
func GetSignature(name string) Signature {
	if value, ok := SignatureList[name]; ok {
		return value
	}
	return nil
}

// RegisterSignature
//
//	@param name
//	@param aaa
func RegisterSignature(name string, aaa Signature) error {
	return nil
}

// RegisterLuaSignature
//
//	@param name
//	@param luaCode
//	@return error
func RegisterLuaSignature(name string, luaCode string) error {
	return nil
}

func genLuaSignature(name string, luaCode string) (Signature, error) {
	proto, err := compileLuaCode(name, luaCode)
	if err != nil {
		return nil, fmt.Errorf("compile lua code error:%s", err.Error())
	}

	return func(input map[string]interface{}, header map[string]string, config map[string]interface{}) (map[string]interface{}, map[string]string, error) {
		// 创建一个lua解释器实例
		l := lua.NewState()
		defer l.Close()
		// 需要执行的lua代码
		lfunc := l.NewFunctionFromProto(proto)
		l.Push(lfunc)
		inputTable := newLuaTable(l, input)
		tmp := map[string]interface{}{}
		for key, value := range header {
			tmp[key] = value
		}
		headerTable := newLuaTable(l, tmp)
		configTable := newLuaTable(l, config)

		err = l.CallByParam(lua.P{
			Fn:      l.GetGlobal(name), // 获取info函数引用
			NRet:    2,                 // 指定返回值数量
			Protect: true,              // 如果出现异常，是panic还是返回err
		}, inputTable, headerTable, configTable) // 传递输入参数n=1
		if err != nil {
			return input, header, err
		}

		var (
			retHeader map[string]string      = map[string]string{}
			retInput  map[string]interface{} = map[string]interface{}{}
		)

		// 获取返回结果
		luaHeader := l.Get(-1)
		// 从堆栈中删除返回值
		l.Pop(1)
		// 打印返回结果

		if table, ok := luaHeader.(*lua.LTable); ok {
			table.ForEach(func(l1, l2 lua.LValue) {
				retHeader[l1.String()] = l2.String()
			})
		}
		luaInput := l.Get(-1)
		l.Pop(1)
		if table, ok := luaInput.(*lua.LTable); ok {
			table.ForEach(func(l1, l2 lua.LValue) {
				retInput[l1.String()] = l2
			})
		}

		return retInput, retHeader, nil
	}, nil
}

// newLuaTable
//
//	@param l
//	@param data
//	@return *lua.LTable
func newLuaTable(l *lua.LState, data map[string]interface{}) *lua.LTable {
	if data == nil {
		return nil
	}
	input := l.NewTable()
	for key, value := range data {
		if tmp, ok := value.(string); ok {
			input.RawSetString(key, lua.LString(tmp))
			continue
		}
		if tmp, ok := value.(float64); ok {
			input.RawSetString(key, lua.LNumber(tmp))
			continue
		}
		if tmp, ok := value.(map[string]interface{}); ok {
			input.RawSetString(key, newLuaTable(l, tmp))
			continue
		}
		if tmp, ok := value.(bool); ok {
			input.RawSetString(key, lua.LBool(tmp))
			continue
		}
	}
	return input
}

// compileLuaCode ...
//
//	@param name
//	@param luaCode
//	@return *lua.FunctionProto
//	@return error
func compileLuaCode(name string, luaCode string) (*lua.FunctionProto, error) {
	reader := strings.NewReader(luaCode)
	chunk, err := parse.Parse(reader, name)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, name)
	if err != nil {
		return nil, err
	}
	return proto, nil
}
