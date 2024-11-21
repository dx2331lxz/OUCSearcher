package tools

import (
	"OUCSearcher/models"
	"OUCSearcher/types"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
)

// 按值排序的函数
func sortMapByValue(m map[string]float64) []types.Pair {
	// 将 map 转换为 Pair 切片
	var pairs []types.Pair
	for k, v := range m {
		pairs = append(pairs, types.Pair{Key: k, Value: v})
	}

	// 排序
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[j].Value // 降序排序
	})

	return pairs
}

// GetNumOfChar 统计`-`在字符串中出现的次数
func GetNumOfChar(q string, char rune) int {
	n := 0
	for _, v := range q {
		if v == char {
			n++
		}
	}
	return n
}

func IDF(indexString string, N int) (float64, error) {
	// 获取`-`出现的次数
	numOfChar := GetNumOfChar(indexString, '-')
	if numOfChar > N {
		return 0, fmt.Errorf("numOfChar > N")
	}
	//	 计算IDF
	idf := math.Log10((float64(N-numOfChar) + 0.5) / (float64(numOfChar) + 0.5))
	return idf, nil
}

// R 计算 R
func R(f int, d int, avgDocLength float64, k1, b float64) float64 {
	K := k1 * ((1 - b) + b*(float64(d)/avgDocLength))
	return (k1 + 1) * float64(f) / (K + float64(f))
}

// BM25 计算 BM25
func BM25(q string) (map[string]float64, error) {
	// 初始化页面评分
	pageScoreMap := make(map[string]float64)

	// 分词
	queryList := fenci(q)

	// 获取页面总数
	N, err := models.GetDicDoneAboutCount()
	if err != nil {
		log.Println("获取页面总数失败:", err)
		return nil, err
	}
	if N <= 0 {
		return nil, fmt.Errorf("页面总数 N <= 0")
	}

	// 平均文档长度和 BM25 参数
	avgDocLength := 1000.0
	k1, b := 2.0, 0.75

	// 遍历每个分词
	for _, query := range queryList {
		// 计算 IDF
		idf, err := IDF(query, N)
		if err != nil {
			log.Printf("计算 IDF 失败 (%s): %v\n", query, err)
			continue
		}

		// 获取索引字符串
		indexString, err := models.GetIndexString(query)
		if err != nil {
			log.Printf("获取 indexString 失败 (%s): %v\n", query, err)
			continue
		}
		if indexString == "" {
			continue
		}

		// 分割索引字符串
		indexStringList := strings.Split(indexString, "-")
		for _, indexString := range indexStringList {
			// 分割索引项
			indexSplit := strings.Split(indexString, ",")
			if len(indexSplit) < 3 {
				log.Printf("索引项格式错误: %s\n", indexString)
				continue
			}

			// 提取文档名和频率
			name := indexSplit[0] + "," + indexSplit[1]
			f, err := strconv.Atoi(indexSplit[2])
			if err != nil {
				log.Printf("转换文档频率失败 (%s): %v\n", indexString, err)
				continue
			}

			// 计算 R 和 BM25
			d, err := strconv.Atoi(indexSplit[3])
			if err != nil {
				log.Printf("转换文档长度失败 (%s): %v\n", indexString, err)
				continue
			}
			r := R(f, d, avgDocLength, k1, b)
			bm25 := idf * r

			// 累加评分
			pageScoreMap[name] += bm25
		}
	}
	return pageScoreMap, nil
}

// GetSortedPageList 获取排序后的页面列表
func GetSortedPageList(q string, n int) []types.Pair {
	// 计算 BM25
	pageScoreMap, err := BM25(q)
	if err != nil {
		log.Println("计算 BM25 失败:", err)
		return nil
	}

	// 排序
	pairs := sortMapByValue(pageScoreMap)
	// 截取前 n 个
	if n > 0 && n < len(pairs) {
		pairs = pairs[:n]
	}
	return pairs
}
