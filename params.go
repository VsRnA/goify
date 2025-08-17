package goify

import (
	"strings"
)

type RouteNode struct {
	path string
	handlers map[string]HandlerFunc
	children map[string]*RouteNode
	paramKey string
	isParam bool
	isWild bool
}

func NewRouteNode() *RouteNode {
	return &RouteNode{
		handlers: make(map[string]HandlerFunc),
		children: make(map[string]*RouteNode),
	}
}

func (node *RouteNode) addRoute (path, method string, handler HandlerFunc) {
	segments := splitPath(path)
	current := node

	for _, segment := range segments {
		if segment == "" {
			continue
		}

		if strings.HasPrefix(segment, ":") {
			paramKey := segment[1:]
			if current.children["*param*"] == nil {
				current.children["*param*"] = NewRouteNode()
				current.children["*param*"].isParam = true
				current.children["*param*"].paramKey = paramKey
			}
			current = current.children["*param*"]
		} else if strings.HasPrefix(segment, "*") {
			paramKey := segment[1:]
			if current.children["*wild*"] == nil {
				current.children["*wild*"] = NewRouteNode()
				current.children["*wild*"].isWild = true
				current.children["*wild*"].paramKey = paramKey
			}
			current = current.children["*wild*"]
			break
		} else {
			if current.children[segment] == nil {
				current.children[segment] = NewRouteNode()
			}
			current = current.children[segment]
		}
	}
	
	current.handlers[method] = handler
	current.path = path
}

func (node *RouteNode) findRoute(path string, method string) (HandlerFunc, map[string]string) {
	segments := splitPath(path)
	params := make(map[string]string)
	
	handler := node.searchRoute(segments, 0, method, params)
	return handler, params
}

func (node *RouteNode) searchRoute(segments []string, index int, method string, params map[string]string) HandlerFunc {
	if index >= len(segments) {
		if handler, exists := node.handlers[method]; exists {
			return handler
		}
		return nil
	}
	
	segment := segments[index]
	if segment == "" {
		return node.searchRoute(segments, index+1, method, params)
	}

	if child, exists := node.children[segment]; exists {
		if handler := child.searchRoute(segments, index+1, method, params); handler != nil {
			return handler
		}
	}

	if paramNode, exists := node.children["*param*"]; exists {
		params[paramNode.paramKey] = segment
		if handler := paramNode.searchRoute(segments, index+1, method, params); handler != nil {
			return handler
		}
		delete(params, paramNode.paramKey)
	}

	if wildNode, exists := node.children["*wild*"]; exists {
		remaining := strings.Join(segments[index:], "/")
		params[wildNode.paramKey] = remaining
		if handler, exists := wildNode.handlers[method]; exists {
			return handler
		}
	}
	
	return nil
}

func splitPath(path string) []string {
	if path == "/" {
		return []string{""}
	}
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{""}
	}
	
	return strings.Split(path, "/")
}

func cleanPath(path string) string {
	if path == "" {
		return "/"
	}
	
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	
	return path
}
