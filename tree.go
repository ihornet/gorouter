package gorouter

type Node struct {
	path       string  // 该节点的path
	isWildcard bool    // 该节点是否是通配符
	indices    string  // 子节点的首字符集合
	child      []*Node // 子节点
	handle     Handle
}

func (node *Node) addRoute(path string, handle Handle) {

	fullPath := path

	if len(path) == 0 || path[0] != '/' {
		panic("路由以 '/' 开始 ")
	}

	// 去掉path末尾的 '/'
	path = removeLastSlash(path)

	if node.path == "" {
		node.insert(path, handle)
		return
	}

loop:

	for {

		offset := 0
		max := len(node.path)

		if len(path) < max {
			max = len(path)
		}
		if max == 0 {
			node.handle = handle
			return
		}

		// 共同通前缀
		for i := 0; i < max; i++ {
			if path[i] != node.path[i] {
				break
			}
			offset++
		}

		// node.path 分裂
		if offset < len(node.path) {

			// 如果是通配符path，那么不允许分裂。如 :test 不能分为 :tes + t
			if node.isWildcard {
				panic("路由" + fullPath + "的" + path[:offset] + "冲突")
			}

			child := &Node{
				path:       node.path[offset:],
				indices:    node.indices,
				isWildcard: node.isWildcard,
				child:      node.child,
				handle:     node.handle,
			}

			node.handle = nil
			node.path = node.path[0:offset]
			node.child = []*Node{child}
			node.indices = string([]byte{child.path[0]})

		}

		if offset < len(path) {

			if node.isWildcard && offset > 0 && (node.path != path[:offset] || path[offset] != '/') {
				panic("路由" + fullPath + "与" + path[:offset] + "冲突")
			}

			path = path[offset:]

			for i := 0; i < len(node.indices); i++ {
				if node.indices[i] == path[0] {
					node = node.child[i]

					_path, _, _ := getNodePath(path)

					if len(node.child) == 0 && _path == path && node.path == _path {
						panic("路由 " + fullPath + " 已经存在")
					}

					continue loop
				}
			}

			node.insert(path, handle)

			return

		} else {
			return
		}

	}

}

func (node *Node) insert(path string, handle Handle) {

	offset := 0

	if node.path == "" {
		node.path, offset, node.isWildcard = getNodePath(path)
	}

	for {

		if offset == len(path) {
			node.handle = handle
			return
		}

		path = path[offset:]

		child := &Node{}
		node.child = append(node.child, child)
		pNode := node
		node = child
		node.path, offset, node.isWildcard = getNodePath(path)

		pNode.indices += string([]byte{node.path[0]})

		pNode.sortIndices()

	}

}

func (node *Node) sortIndices() {

	indicesLen := len(node.indices)

	if indicesLen == 1 {
		return
	}

	suffix := node.indices[indicesLen-1]
	i := indicesLen - 2

	for ; i >= 0; i-- {
		if node.indices[i] < suffix {
			break
		}
		node.child[i+1], node.child[i] = node.child[i], node.child[i+1]
	}

	node.indices = node.indices[:i+1] + node.indices[indicesLen-1:indicesLen] + node.indices[i+1:indicesLen-1]

}

func (node *Node) getValue(path string) (handle Handle, params *Param, isMatch bool) {

	path = removeLastSlash(path)

	// 和根节点不匹配
	if node.path != path[:len(node.path)] {
		return
	}

	if path == node.path && node.handle != nil {
		return node.handle, nil, true
	}

	handle, params, isMatch = node.seekRoute(path[len(node.path):], params)

	return
}

func (node *Node) seekRoute(path string, params *Param) (Handle, *Param, bool) {

	wildCardIndex := -1
	index := -1

	for i := 0; i < len(node.indices); i++ {

		if node.indices[i] != ':' && node.indices[i] > path[0] {
			break
		}

		isBreak := false
		if node.indices[i] == path[0] && path[0] != ':' {
			if wildCardIndex > -1 {
				isBreak = true
			}
			index = i
		}

		if node.indices[i] == ':' {
			if index > -1 {
				isBreak = true
			}
			wildCardIndex = i
		}

		if isBreak {
			break
		}

	}

	// 没有indices与之匹配
	if index == -1 && wildCardIndex == -1 {
		return nil, params, false
	}

	// 首先匹配静态路径
	if index > -1 {
		cnode := node.child[index]

		if len(path) >= len(cnode.path) && cnode.path == path[:len(cnode.path)] {

			// 匹配成功
			if path == cnode.path && cnode.handle != nil {
				return cnode.handle, params, true
			}

			cpath := path[len(cnode.path):]

			if cpath == "" {
				return nil, params, false
			}

			handle, params, seeked := cnode.seekRoute(cpath, params)

			// 如果静态路由匹配成功，则返回。否则继续匹配通配符
			if seeked {
				return handle, params, true
			}
		}
	}

	// 尝试匹配通配符
	if wildCardIndex > -1 {
		node = node.child[wildCardIndex]

		for i := 0; i <= len(path); i++ {
			if i == len(path) || path[i] == '/' {

				if params == nil {
					params = acquireParam()
				}

				index := params.addKV(node.path[1:], path[:i])

				if i == len(path) {
					if node.handle != nil {
						return node.handle, params, true
					} else {
						return nil, params, false
					}
				} else {
					handle, params, seeked := node.seekRoute(path[i:], params)

					if seeked {
						return handle, params, true
					}
					params.remove(index - 1)
				}

				break
			}
		}
	}

	return nil, params, false
}

/**
 *   去掉末尾的 '/'
 */
func removeLastSlash(path string) string {
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return path
}

func getNodePath(path string) (string, int, bool) {

	var (
		i          = 0
		char       = ':'
		isWildcard = false
	)

	if path[0] == ':' {
		char = '/'
		isWildcard = true
	}

	for i = 0; i < len(path) && path[i] != uint8(char); i++ {
	}

	return path[:i], i, isWildcard
}
