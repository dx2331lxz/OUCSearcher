package routers

import (
	"OUCSearcher/controllers"
	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	beego.Router("/", &controllers.IndexController{})
	beego.Router("/search", &controllers.SearchController{})
}
