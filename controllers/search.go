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
	searchResultList, err := models.GetSearchResultFromPair(pairs)
	fmt.Println("查询结果: ", searchResultList)
	if err != nil {
		c.Data["Query"] = q
		c.Data["Results"] = []models.SearchResult{}
		c.TplName = "search.tpl"
		return
	}
	// 将searchResultList的Description字段截取前50个字符
	for i, v := range searchResultList {
		if len(v.Description) > 300 {
			searchResultList[i].Description = v.Description[:300]
		}
	}

	c.Data["Query"] = q
	c.Data["Results"] = searchResultList
	c.TplName = "search.tpl"
}

func (c *SearchController) WxSearch() {
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
	searchResultList, err := models.GetSearchResultFromPair(pairs)
	fmt.Println("查询结果: ", searchResultList)
	if err != nil {
		c.Data["json"] = map[string]string{"error": "查询失败"}
		c.ServeJSON()
		return
	}

	for i, v := range searchResultList {
		if len(v.Description) > 300 {
			searchResultList[i].Description = v.Description[:300]
		}
	}
	c.Data["json"] = searchResultList
	c.ServeJSON()
}
