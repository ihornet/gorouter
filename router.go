package gorouter

import (
	"net/http"
)

type Handle func(resp http.ResponseWriter, req *http.Request, params *Param)

type Router struct {
	trees    map[string]*Node
	NotFound http.Handler
}

// var trees = make(map[string]*Node)

func New() *Router {
	return &Router{
		trees: make(map[string]*Node),
	}
}

func (router *Router) GET(path string, handle Handle) {
	router.handleFunc(http.MethodGet, path, handle)
}

func (router *Router) POST(path string, handle Handle) {
	router.handleFunc(http.MethodPost, path, handle)
}

func (router *Router) PUT(path string, handle Handle) {
	router.handleFunc(http.MethodPut, path, handle)
}

func (router *Router) DELETE(path string, handle Handle) {
	router.handleFunc(http.MethodDelete, path, handle)
}

func (router *Router) handleFunc(method, path string, handle Handle) {

	if tree := router.trees[method]; tree == nil {
		router.trees[method] = &Node{}
	}

	router.trees[method].addRoute(path, handle)
}

// http 回调
func (router Router) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	node := router.trees[req.Method]

	if node != nil {
		handle, param, isMatch := node.getValue(req.URL.Path)

		if isMatch {
			if handle != nil {
				handle(resp, req, param)
				if param != nil {
					releaseParam(param)
				}
				return
			}
		}
	}

	if router.NotFound != nil {
		router.NotFound.ServeHTTP(resp, req)
	} else {
		http.NotFound(resp, req)
	}
}
