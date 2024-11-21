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
	pageDics, err := models.GetPageDicFromPair(pairs)
	fmt.Println("查询结果: ", pageDics)
	if err != nil {
		c.Data["json"] = map[string]string{"error": "查询失败: " + err.Error()}
		c.ServeJSON()
		return
	}

	// 返回 JSON 数据
	c.Data["json"] = pageDics
	c.ServeJSON()
}
