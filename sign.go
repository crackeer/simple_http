package simple_http

import (
	"fmt"
	"sync"
)

// SignHandle
type SignHandle interface {
	Sign(*ServiceAPI, map[string]interface{}, map[string]string) (*ServiceAPI, map[string]interface{}, map[string]string, error)
	ID() string
	Introduction() string
	SignConfigTemplate() map[string]interface{}
}

var (
	signHandleMap map[string]SignHandle
	locker        *sync.RWMutex
)

func init() {
	signHandleMap = map[string]SignHandle{}
	locker = &sync.RWMutex{}
}

// GetSignHandleMap
//
//	@return map
func GetSignHandleMap() map[string]SignHandle {
	return signHandleMap
}

// GetSign GetSignature
//
//	@param name
//	@return Signature
func GetSignHandle(name string) SignHandle {
	if value, ok := signHandleMap[name]; ok {
		return value
	}
	return nil
}

// RegisterSign RegisterSignature
//
//	@param name
//	@param aaa
//	@return error
func RegisterSign(handle SignHandle) error {
	locker.Lock()
	defer locker.Unlock()
	if _, ok := signHandleMap[handle.ID()]; ok {
		return fmt.Errorf("`%s` already registered", handle.ID())
	}
	signHandleMap[handle.ID()] = handle
	return nil
}
