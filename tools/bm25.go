package tools

import (
	"OUCSearcher/models"
	"OUCSearcher/types"
	"container/heap"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// 定义最小堆
type MinHeap []types.Pair

// 实现 heap.Interface 接口
func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].Value < h[j].Value } // 小值在堆顶
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

// Push 和 Pop 方法用于堆操作
func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(types.Pair))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

// 排序并获取前 k 个最大值
func getTopK(m map[string]float64, k int) []types.Pair {
	// 初始化一个最小堆
	h := &MinHeap{}
	heap.Init(h)

	// 遍历 map，将元素添加到堆中
	for key, value := range m {
		// 如果堆的大小小于 k，直接加入
		if h.Len() < k {
			heap.Push(h, types.Pair{Key: key, Value: value})
		} else {
			// 如果堆的大小已经达到 k，且当前元素更大，替换堆顶元素
			if h.Len() > 0 && value > (*h)[0].Value { // 使用 (*h)[0] 来访问堆顶
				heap.Pop(h)                                      // 弹出堆顶
				heap.Push(h, types.Pair{Key: key, Value: value}) // 插入当前元素
			}
		}
	}

	// 将堆中的元素转换为切片返回
	result := make([]types.Pair, 0, k)
	for h.Len() > 0 {
		result = append(result, heap.Pop(h).(types.Pair))
	}

	// 直接使用 sort.Slice 排序，按降序排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Value > result[j].Value // 降序排序
	})

	return result
}

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

// 判断k是否远小于n
func isKFarLessThanN(k, n int) bool {
	r := 0.1
	return float64(k)/float64(n) < r
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
	resultChannel := make(chan map[string]float64)
	var wg sync.WaitGroup
	go func() {
		for result := range resultChannel {
			for k, v := range result {
				pageScoreMap[k] += v
			}
		}
	}()

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
		wg.Add(1)

		func(query string) {
			defer wg.Done()
			// 计算 IDF
			idf, err := IDF(query, N)
			if err != nil {
				log.Printf("计算 IDF 失败 (%s): %v\n", query, err)
				return
			}

			// 获取索引字符串
			indexString, err := models.GetIndexString(query)
			if err != nil {
				log.Printf("获取 indexString 失败 (%s): %v\n", query, err)
				return
			}
			if indexString == "" {
				return
			}

			// 分割索引字符串
			indexStringList := strings.Split(indexString, "-")
			for _, indexString := range indexStringList {
				wg.Add(1)
				go func(indexString string) {
					defer wg.Done()
					// 分割索引项
					indexSplit := strings.Split(indexString, ",")
					if len(indexSplit) < 3 {
						log.Printf("索引项格式错误: %s\n", indexString)
						return
					}

					// 提取文档名和频率
					name := indexSplit[0] + "," + indexSplit[1]
					f, err := strconv.Atoi(indexSplit[2])
					if err != nil {
						log.Printf("转换文档频率失败 (%s): %v\n", indexString, err)
						return
					}

					// 计算 R 和 BM25
					d, err := strconv.Atoi(indexSplit[3])
					if err != nil {
						log.Printf("转换文档长度失败 (%s): %v\n", indexString, err)
						return
					}
					r := R(f, d, avgDocLength, k1, b)
					bm25 := idf * r
					// 累加评分
					resultChannel <- map[string]float64{name: bm25}
				}(indexString)
			}
		}(query)
	}
	wg.Wait()
	close(resultChannel)
	return pageScoreMap, nil
}

// GetSortedPageList 获取排序后的页面列表
func GetSortedPageList(q string, k int) []types.Pair {
	// 计算 BM25
	pageScoreMap, err := BM25(q)
	if err != nil {
		log.Println("计算 BM25 失败:", err)
		return nil
	}
	// 所有页面个数
	n := len(pageScoreMap)
	if n <= k {
		return sortMapByValue(pageScoreMap)
	}
	// 如果 k 远小于 n，使用最小堆
	if isKFarLessThanN(k, n) {
		fmt.Println("使用最小堆")
		return getTopK(pageScoreMap, k)
	} else {
		// 排序
		pairs := sortMapByValue(pageScoreMap)
		// 截取前 k 个
		if k > 0 && k < len(pairs) {
			pairs = pairs[:k]
		}
		return pairs
	}
}
