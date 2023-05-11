package simple_http

import (
	"fmt"
	"os"
	"strings"
	"sync"

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

var (
	signFunc map[string]Signature
	locker   *sync.RWMutex
)

func init() {
	signFunc = map[string]Signature{}
	locker = &sync.RWMutex{}
}

// GetSign GetSignature
//
//	@param name
//	@return Signature
func GetSign(name string) Signature {
	if value, ok := signFunc[name]; ok {
		return value
	}
	return nil
}

// RegisterSign RegisterSignature
//
//	@param name
//	@param aaa
//	@return error
func RegisterSign(name string, aaa Signature) error {
	locker.Lock()
	defer locker.Unlock()
	signFunc[name] = aaa
	return nil
}

// RegisterLuaSignByFile
//
//	@param name
//	@param file
//	@return error
func RegisterLuaSignByFile(name string, file string) error {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("open file %s: %v", file, err)
	}

	return RegisterLuaSign(name, string(bytes))
}

// RegisterLuaSign RegisterLuaSignature
//
//	@param name
//	@param luaCode
//	@return error
func RegisterLuaSign(name string, luaCode string) error {

	s, err := genLuaSignature(name, luaCode)
	if err != nil {
		return fmt.Errorf("gen lua signature %s: %v", name, err)
	}
	locker.Lock()
	defer locker.Unlock()

	signFunc[name] = s
	return nil
}

// genLuaSignature
//
//	@param name
//	@param luaCode
//	@return Signature
//	@return error
func genLuaSignature(name string, luaCode string) (Signature, error) {
	/*
		proto, err := compileLuaCode(name, luaCode)
		if err != nil {
			return nil, fmt.Errorf("compile lua code error:%s", err.Error())
		}
	*/
	l := lua.NewState()
	err := l.DoString(luaCode)
	if err != nil {
		return nil, fmt.Errorf("compile lua code error:%s", err.Error())
	}

	return func(input map[string]interface{}, header map[string]string, config map[string]interface{}) (map[string]interface{}, map[string]string, error) {
		// 创建一个lua解释器实例
		// 需要执行的lua代码
		inputTable := newLuaTable(l, input)
		headerTable := newLuaTable(l, ToInterfaceMap(header))
		configTable := newLuaTable(l, config)
		err = l.CallByParam(lua.P{
			Fn:      l.GetGlobal(name), // 获取info函数引用
			NRet:    2,                 // 指定返回值数量
			Protect: true,              // 如果出现异常，是panic还是返回err
		}, inputTable, headerTable, configTable) // 传递输入参数n=3
		if err != nil {
			fmt.Println(err.Error(), name)
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
	input.RawSetString("simple", lua.LString("simple"))
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
