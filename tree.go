package rum

type nodeType uint8

const (
	static nodeType = iota // default
	root
	param
	catchAll
)

type node struct {
	path         string
	nType        nodeType
	idxcs        string
	child        []*node
	hasWildChild bool
	handlers     HandlersChain
}

func findWildcard(path string) (wilcard string, i int, valid bool) {
	// Find start
	for start, c := range []byte(path) {
		// A wildcard starts with ':' (param) or '*' (catch-all)
		if c != ':' && c != '*' {
			continue
		}

		// Find end and check for invalid characters
		valid = true
		for end, c := range []byte(path[start+1:]) {
			switch c {
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

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// addChild will add a child node, keeping wildcards at the end
func (n *node) addChild(child *node) {
	if n.hasWildChild && len(n.child) > 0 {
		wildcardChild := n.getWildChild()
		n.child = append(n.child[:len(n.child)-1], child, wildcardChild)
	} else {
		n.child = append(n.child, child)
	}
}

func (n *node) getWildChild() *node {
	return n.child[len(n.child)-1]
}

func longestCommonPrefix(a, b string) int {
	i, end := 0, min(len(a), len(b))
	for i < end && a[i] == b[i] {
		i++
	}
	return i
}

// addRoute adds a node with the given handlers to the path
func (n *node) addRoute(path string, handlers HandlersChain) {

	// Empty tree
	if n.path == "" && n.idxcs == "" {
		n.insertChild(path, path, handlers)
		n.nType = root
		return
	}
	n.splitOrMakeNode(path, path, handlers)
}

// Radix tree operation
// The function finds the longest commonPrefix of path and node'path(the commonPrefix string len is inx)
// If inx smaller than node's path length, split the node
// If inx smaller than path length, make new node a child of this node
func (n *node) splitOrMakeNode(path string, fullPath string, handlers HandlersChain) {
	inx := longestCommonPrefix(path, n.path)

	if inx < len(n.path) {
		child := node{
			path:         n.path[inx:],
			hasWildChild: n.hasWildChild,
			child:        n.child,
			handlers:     n.handlers,
			idxcs:        n.idxcs,
		}
		n.child = []*node{&child}
		n.idxcs = string([]byte{n.path[inx]})
		n.path = n.path[:inx]
		n.hasWildChild = false
		n.handlers = nil
	}

	if inx < len(path) {
		path = path[inx:]

		idxc := path[0]
		if n.nType == param && idxc == '/' && len(n.child) == 1 {
			n = n.child[0]
			n.splitOrMakeNode(path, fullPath, handlers)
			return
		}

		// Check if a child with the next path byte exists
		for i := 0; i < len(n.idxcs); i++ {
			if idxc == n.idxcs[i] {
				n = n.child[i]
				n.splitOrMakeNode(path, fullPath, handlers)
				return
			}
		}

		if idxc != ':' && idxc != '*' && n.nType != catchAll {
			// []byte for proper unicode char conversion
			n.idxcs += string([]byte{idxc})
			child := &node{}
			n.addChild(child)
			n = child
		} else if n.hasWildChild {
			n = n.getWildChild()

			// Check if the wildcard match or conflict
			if len(path) >= len(n.path) && n.path == path[:len(n.path)] &&
				n.nType != catchAll &&
				(len(n.path) == len(path) || path[len(n.path)] == '/') {
				n.addRoute(path, handlers)
				return
			} else {
				panic("wildcar conflict! fullpath is '" + fullPath + "' node path is '" + n.path + "' subpath is '" + path + "'")
			}
		}
		n.insertChild(path, fullPath, handlers)
		return
	}

	if n.handlers != nil {
		panic("handlers are already registered for path '" + fullPath + "'")
	}
	n.handlers = handlers
}

func (n *node) insertChild(path, fullPath string, handlers HandlersChain) {
	for {
		wildcard, i, valid := findWildcard(path)

		// No wildcard found
		if i < 0 {
			break
		}

		// The wildcard name must only contain one ':' or '*' charactor
		if !valid {
			panic("only one wildcard per path segment is allowed, has: '" +
				wildcard + "' in path '" + fullPath + "'")
		}

		// check if the wildcard has a name
		if len(wildcard) < 2 {
			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		}

		if wildcard[0] == ':' {
			if i > 0 {
				n.path = path[:i]
				path = path[i:]
			}

			child := &node{
				path:         wildcard,
				nType:        param,
				hasWildChild: false,
			}

			n.addChild(child)
			n.hasWildChild = true
			n = child

			if len(wildcard) < len(path) {
				path = path[len(wildcard):]

				child := &node{}
				n.addChild(child)
				n = child
				continue
			}

			n.handlers = handlers
			return
		}

		// catchAll case
		if i+len(wildcard) != len(path) {
			panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
		}

		if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
			panic("catch-all conflicts with existing handle for the path segment root in path '" + fullPath + "'")
		}

		i--
		if path[i] != '/' {
			panic("no / before catch-all in path '" + fullPath + "'")
		}

		n.path = path[:i]

		// first node to catchAll
		child := &node{
			hasWildChild: true,
			nType:        catchAll,
		}

		n.addChild(child)
		n.idxcs = string('/')
		n = child

		// second node to hold variable
		child = &node{
			path:     path[i:],
			nType:    catchAll,
			handlers: handlers,
		}
		n.child = []*node{child}

		return

	}

	n.path = path
	n.handlers = handlers
}

func (n *node) saveParam(params *Params, ps *Params, key, value string) {
	if params != nil {
		li := len(*ps)
		*ps = (*ps)[:li+1]
		(*ps)[li] = Param{
			Key:   key,
			Value: value,
		}
	}
}

func (n *node) getValue(path string, params *Params) (handlers HandlersChain, ps *Params) {
	for {
		prefix := n.path
		if len(path) > len(prefix) && path[:len(prefix)] == prefix {
			path = path[len(prefix):]

			// first match non-wildChild
			if !n.hasWildChild {
				idxc := path[0]
				flag := false
				for i, c := range []byte(n.idxcs) {
					if c == idxc {
						n = n.child[i]
						flag = true
						break
					}
				}
				if flag {
					continue
				} else {
					break
				}
			}

			n = n.getWildChild()

			switch n.nType {
			case param:
				end := 0
				for end < len(path) && path[end] != '/' {
					end++
				}
				// save param value
				if ps == nil {
					ps = params
				}
				n.saveParam(params, ps, n.path[1:], path[:end])
				// go deeper
				if end < len(path) {
					if len(n.child) > 0 {
						path = path[end:]
						n = n.child[0]
						continue
					}

					return
				}
			case catchAll:
				// save param value
				if ps == nil {
					ps = params
				}
				n.saveParam(params, ps, n.path[2:], path)
			default:
				panic("invaild node type")
			}

			handlers = n.handlers
			break
		}

		if path == prefix {
			handlers = n.handlers
			break
		}

		// Nothing found.
		break
	}
	return handlers, ps
}
