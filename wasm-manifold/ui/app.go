package ui

import (
	"syscall/js"

	"github.com/gowasm/vecty"
	"github.com/progrium/prototypes/go-webui"
)

func init() {
	webui.Register(App{})
}

type App struct {
	vecty.Core

	TreeView *TreeView `vecty:"ref"`
}

func (c *App) OnReset(e *vecty.Event) {
	js.Global().Get("localStorage").Call("setItem", "tree_nodes", "[]")
	js.Global().Get("localStorage").Call("setItem", "tree_nodeIDs", "{}")
	js.Global().Get("location").Call("reload")
}

func (c *App) OnAdd(e *vecty.Event) {
	var name = js.Global().Call("prompt", "New object").String()
	c.TreeView.CreateNode(TreeNode{
		Text: name,
	})
}

func (c *App) Mount() {

}

func (c *App) Render() vecty.ComponentOrHTML {
	return webui.Render(c)
}
