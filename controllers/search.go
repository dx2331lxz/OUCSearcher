package controllers

import (
	"OUCSearcher/models"
	"OUCSearcher/tools"
	"fmt"
	beego "github.com/beego/beego/v2/server/web"
)

type SearchController struct {
	beego.Controller
}

type PageData struct {
	Query   string                // 用户查询的关键词
	Results []models.SearchResult // 搜索结果列表
}

func (c *SearchController) Get() {
	// 获取参数 q
	q := c.GetString("q")
	if q == "" {
		c.Data["json"] = map[string]string{"error": "查询参数 q 不能为空"}
		c.ServeJSON()
		return
	}
	fmt.Println("查询参数: ", q)
	// 获取排序后的页面列表
	pairs := tools.GetSortedPageList(q, 20)
	fmt.Println("排序后的页面列表: ", pairs)
	seachResultList, err := models.GetSearchResultFromPair(pairs)
	fmt.Println("查询结果: ", seachResultList)
	if err != nil {
		c.Data["Query"] = q
		c.Data["Results"] = []models.SearchResult{}
		c.TplName = "search.tpl"
		return
	}

	c.Data["Query"] = q
	c.Data["Results"] = seachResultList
	c.TplName = "search.tpl"

}
