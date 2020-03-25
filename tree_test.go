package gorouter

import (
	"fmt"
	"net/http"
	"testing"
)

func TestTree(t *testing.T) {

	node := &Node{}

	node.addRoute("/hai/:test/a/b/c", nil)

	node.addRoute("/hai/:test/a/b", func(resp http.ResponseWriter, req *http.Request, params *Param) {

	})

	node.addRoute("/hai/:test/abc/b", nil)

	node.addRoute("/hai/test/abc/b", nil)

	node.addRoute("/zhang/:test/abc/:b/c", nil)

	node.addRoute("/zhang/z", nil)

	fmt.Print(node)

}
