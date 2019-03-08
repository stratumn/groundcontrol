// Copyright 2019 Stratumn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import "context"

// ComputeDependencies computes the dependencies of the Service based on the Services it needs.
func (n *Service) ComputeDependencies(ctx context.Context) error {
	deps, err := n.TopologicalSort(ctx)
	if err != nil {
		return err
	}
	n.DependenciesIDs = nil
	for _, dep := range deps {
		n.DependenciesIDs = append(n.DependenciesIDs, dep.ID)
	}
	return nil
}

// ComputeAllVariables computes all the Variables of this Service based on the Services it needs.
// It assumes that ComputeDependencies has been called.
func (n *Service) ComputeAllVariables(ctx context.Context) {
	var variablesIDs []string
	// Get all variable IDs, which might include duplicates.
	for _, serviceID := range n.DependenciesIDs {
		service := MustLoadService(ctx, serviceID)
		variablesIDs = append(variablesIDs, service.VariablesIDs...)
		for _, taskID := range service.BeforeIDs {
			variablesIDs = append(variablesIDs, MustLoadTask(ctx, taskID).VariablesIDs...)
		}
		for _, taskID := range service.AfterIDs {
			variablesIDs = append(variablesIDs, MustLoadTask(ctx, taskID).VariablesIDs...)
		}
	}
	n.AllVariablesIDs = nil
	variableSet := map[string]*Variable{}
	// Remove duplicates.
	for _, variableID := range variablesIDs {
		variable := MustLoadVariable(ctx, variableID)
		// The first instance to have a default value is prioritized.
		if curr, ok := variableSet[variable.Name]; ok {
			if curr.Default != nil || variable.Default == nil {
				continue
			}
		}
		variableSet[variable.Name] = variable
	}
	for _, variable := range variableSet {
		n.AllVariablesIDs = append(n.AllVariablesIDs, variable.ID)
	}
}

// TopologicalSort sorts the needed services topologically, including this service.
// In layman's terms, it returns the services in an order they can be started to respect the dependency graph.
func (n *Service) TopologicalSort(ctx context.Context) ([]*Service, error) {
	return n.depSort(ctx, &map[string]bool{})
}

// A mark is undefined if the Service hasn't been visited, false if being visited, and true if visited.
func (n *Service) depSort(ctx context.Context, marks *map[string]bool) ([]*Service, error) {
	var deps []*Service
	// Check if we've already been here.
	if mark, ok := (*marks)[n.ID]; ok {
		if !mark {
			return nil, ErrCyclic
		}
		// We already visited this node.
		return nil, nil
	}
	// Mark current service as being visited.
	(*marks)[n.ID] = false
	// Visit needed services.
	for _, serviceID := range n.NeedsIDs {
		subdeps, err := MustLoadService(ctx, serviceID).depSort(ctx, marks)
		if err != nil {
			return nil, err
		}
		// Add subdependencies.
		deps = append(deps, subdeps...)
	}
	// Mark current service as visited.
	(*marks)[n.ID] = true
	// Add this service as a dependency.
	deps = append(deps, n)
	return deps, nil
}
