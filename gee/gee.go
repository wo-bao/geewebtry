package gee

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(*Context)

// Engine implement the interface of ServeHTTP
type Engine struct {
	router *router
	*RouterGroup
	groups []*RouterGroup // 储存所有的groups
	htmlTemplates *template.Template // 用于html渲染
	funcMap template.FuncMap // 用于html渲染
}

type RouterGroup struct {
	prefix string
	middlewares []HandlerFunc // 存储该group的中间件, engine有RouterGroup, 也可以调用use来生成中间件,存储到engine.middlewares
	parent *RouterGroup
	engine *Engine // 所有group都属于同一个engine
}

// New is the constructor of gee.Engine
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup} // 这里就添加了根group
	return engine
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine  // 相同engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine, // 继承同一个engine
	}
	engine.groups = append(engine.groups, newGroup) // 所有的group都会添加到engine.groups
	return newGroup
}

func (group *RouterGroup) Use(middleWares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middleWares...)
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
// 这里修改了方法的接收者,从engine变为了group,但engine封装了group,所以engine依旧可以照旧添加路由(r.Get()),也可以生成分组,再添加路由
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

func (group *RouterGroup) creatStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := os.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.creatStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middleWares []HandlerFunc
	for _, group := range engine.groups {  // 遍历所有group
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middleWares = append(middleWares, group.middlewares...)  //有前缀与这个group相同, 添加这个group要执行的中间件到middlewares
		}
	}
	c := newContext(w, req)
	c.handlers = middleWares
	c.engine = engine
	engine.router.handle(c)
}

func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		c.Next()
	}
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}