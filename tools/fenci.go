package tools

import "github.com/yanyiwu/gojieba"

func fenci(s string) []string {
	var words []string
	use_hmm := true
	x := gojieba.NewJieba()
	defer x.Free()
	words = x.CutForSearch(s, use_hmm)
	return words
}
