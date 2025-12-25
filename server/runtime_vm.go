// Copyright 2018 The Nakama Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"errors"
	"fmt"
	"sync"

	"github.com/dop251/goja"
	"go.uber.org/zap"
)

// JavaScriptVM represents a JavaScript virtual machine instance using goja
type JavaScriptVM struct {
	mu      sync.RWMutex
	runtime *goja.Runtime
	logger  *zap.Logger
	vmID    string
}

// NewJavaScriptVM creates a new JavaScript VM instance
func NewJavaScriptVM(vmID string, logger *zap.Logger) *JavaScriptVM {
	return &JavaScriptVM{
		runtime: goja.New(),
		logger:  logger,
		vmID:    vmID,
	}
}

// LoadScript loads and executes a JavaScript script in the VM
func (vm *JavaScriptVM) LoadScript(script string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if script == "" {
		return errors.New("script cannot be empty")
	}

	_, err := vm.runtime.RunString(script)
	if err != nil {
		vm.logger.Error("Failed to load script in VM", zap.String("vmID", vm.vmID), zap.Error(err))
		return fmt.Errorf("failed to load script: %w", err)
	}

	vm.logger.Debug("Script loaded successfully", zap.String("vmID", vm.vmID))
	return nil
}

// CallFunction calls a JavaScript function in the VM
func (vm *JavaScriptVM) CallFunction(functionName string, args ...interface{}) (interface{}, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	// Get the function from the runtime
	fn, ok := goja.AssertFunction(vm.runtime.Get(functionName))
	if !ok {
		return nil, fmt.Errorf("function '%s' not found or is not a function", functionName)
	}

	// Convert Go values to goja values
	gojaArgs := make([]goja.Value, len(args))
	for i, arg := range args {
		gojaArgs[i] = vm.runtime.ToValue(arg)
	}

	// Call the function
	result, err := fn(goja.Undefined(), gojaArgs...)
	if err != nil {
		vm.logger.Error("Failed to call function in VM",
			zap.String("vmID", vm.vmID),
			zap.String("function", functionName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to call function '%s': %w", functionName, err)
	}

	// Convert goja value back to Go interface{}
	return result.Export(), nil
}

// GetVMID returns the VM identifier
func (vm *JavaScriptVM) GetVMID() string {
	return vm.vmID
}

// Destroy cleans up the VM resources
func (vm *JavaScriptVM) Destroy() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.runtime = nil
}
