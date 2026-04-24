package bolt

import "strings"

type nodeType uint8

const (
	static nodeType = iota
	param
	catchAll
)

type node struct {
	path      string
	indices   string
	children  []*node
	wildChild bool
	nType     nodeType
	handler   Handler
	paramName string
}

func (n *node) addRoute(path string, handler Handler) {
	fullPath := path

	if n.path == "" && len(n.children) == 0 {
		n.insertChild(path, fullPath, handler)
		return
	}

	prefixLen := longestCommonPrefix(path, n.path)
	if prefixLen < len(n.path) {
		child := &node{
			path:      n.path[prefixLen:],
			indices:   n.indices,
			children:  n.children,
			wildChild: n.wildChild,
			handler:   n.handler,
			nType:     n.nType,
			paramName: n.paramName,
		}

		n.path = n.path[:prefixLen]
		n.indices = string(child.path[0])
		n.children = []*node{child}
		n.wildChild = false
		n.handler = nil
	}

	if prefixLen < len(path) {
		remaining := path[prefixLen:]

		if i := strings.IndexByte(n.indices, remaining[0]); i != -1 {
			n.children[i].addRoute(remaining, handler)
			return
		}

		child := &node{}
		child.insertChild(remaining, fullPath, handler)

		indexChar := remaining[0]
		n.indices += string(indexChar)
		if child.wildChild || child.nType == param || child.nType == catchAll || indexChar == ':' || indexChar == '*' {
			n.wildChild = true
		}
		n.children = append(n.children, child)
		return
	}

	n.handler = handler

}

func (n *node) insertChild(path, fullPath string, handler Handler) {
	for {
		wildcard, index, valid := findWildcard(path)

		if index == -1 {
			n.path += path
			n.handler = handler
			return
		}

		if !valid {
			panic("invalid wildcard in path '" + fullPath + "'")
		}

		if index > 0 {
			n.path += path[:index]
			path = path[index:]
		}

		switch path[0] {
		case ':':
			paramName := wildcard[1:]

			for _, existing := range n.children {
				if existing.nType == param && existing.paramName != paramName {
					panic("conflicting param names: ':" + existing.paramName +
						"' and ':" + paramName + "' for path '" + fullPath + "'")
				}
			}

			child := &node{
				path:      wildcard,
				nType:     param,
				paramName: paramName,
			}
			n.indices += ":"
			n.children = append(n.children, child)
			n.wildChild = true

			if len(wildcard) < len(path) {
				remaining := path[len(wildcard):]
				grandchild := &node{}
				grandchild.insertChild(remaining, fullPath, handler)
				child.indices += string(grandchild.path[0])
				child.children = append(child.children, grandchild)
				return
			}

			child.handler = handler
			return
		case '*':
			n.children = append(n.children, &node{
				path:      wildcard,
				nType:     catchAll,
				handler:   handler,
				paramName: wildcard[1:],
			})
			n.indices += "*"
			n.wildChild = true
			return
		}
	}
}

func (n *node) search(path string) (Handler, Params) {
	if n.nType == param {
		end := strings.Index(path, "/")
		if end == -1 {
			end = len(path)
		}
		paramValue := path[:end]

		if end == len(path) {
			return n.handler, Params{{Key: n.paramName, Value: paramValue}}
		}

		remaining := path[end:]
		for _, child := range n.children {
			handler, params := child.search(remaining)
			if handler != nil {
				return handler, append(Params{{Key: n.paramName, Value: paramValue}}, params...)
			}
		}
		return nil, nil
	}

	if n.nType == catchAll {
		return n.handler, Params{{Key: n.paramName, Value: path}}
	}

	if !strings.HasPrefix(path, n.path) {
		return nil, nil
	}

	if path == n.path {
		return n.handler, nil
	}

	remaining := path[len(n.path):]

	if !n.wildChild {
		if i := strings.IndexByte(n.indices, remaining[0]); i != -1 {
			return n.children[i].search(remaining)
		}
		return nil, nil
	}

	for i, child := range n.children {
		if n.indices[i] != ':' && n.indices[i] != '*' && remaining[0] == n.indices[i] {
			handler, params := child.search(remaining)
			if handler != nil {
				return handler, params
			}
		}
	}
	for i, child := range n.children {
		if n.indices[i] == ':' || n.indices[i] == '*' {
			handler, params := child.search(remaining)
			if handler != nil {
				return handler, params
			}
		}
	}

	return nil, nil
}

func longestCommonPrefix(a, b string) int {
	max := min(len(a), len(b))

	for i := range max {
		if a[i] != b[i] {
			return i
		}
	}

	return max
}

func findWildcard(path string) (wildcard string, index int, valid bool) {
	for start, char := range []byte(path) {
		if char != ':' && char != '*' {
			continue
		}

		valid = true
		for end, char := range []byte(path[start+1:]) {
			switch char {
			case '/':
				return path[start : start+1+end], start, valid
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid
	}

	return "", -1, false
}
