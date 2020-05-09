package go_sensitive

import (
	"strings"
	"sync"
)

type MATCHTYPE int

const (
	INVALID_WORDS = " ,~,!,@,#,$,%,^,&,*,(,),_,-,+,=,?,<,>,.,—,，,。,/,\\,|,《,》,？,;,:,：,',‘,；,“,！,。,；,：,’,{,},【,】,[,],、"
	SENSITIVE_CHILDRED_SIZE = 128

	SINGLE MATCHTYPE = iota
	ALL
)

var InvalidWords = make(map[string]interface{})

func init() {
	words := strings.Split(INVALID_WORDS, ",")
	for _, v := range words {
		InvalidWords[v] = nil
	}
}

type sensitiveNode struct {
	isEnd bool

	children map[rune]*sensitiveNode
}

func newSensitiveNode() *sensitiveNode {
	return &sensitiveNode{
		isEnd:    false,
		children: make(map[rune]*sensitiveNode, SENSITIVE_CHILDRED_SIZE),
	}
}

type DfaUtil struct {
	root *sensitiveNode

	mu sync.Mutex
}

type matchIndex struct {
	start int
	end int
}

func newMatchIndex(start, end int) *matchIndex {
	return &matchIndex{
		start: start,
		end:   end,
	}
}

//初始化Trie树
func newDfaUtil(wordList []string) *DfaUtil {
	dfaUtil := &DfaUtil{
		root: newSensitiveNode(),
	}

	for _, word := range wordList {
		wordRuneList := []rune(word)
		//是词语才加入
		if len(wordRuneList) > 1 {
			dfaUtil.addWord(wordRuneList)
		}
	}

	return dfaUtil
}

//添加敏感词汇
func (dfaUtil *DfaUtil) addWord(words []rune) {
	dfaUtil.mu.Lock()
	defer dfaUtil.mu.Unlock()

	currNode := dfaUtil.root
	for _, word := range words {
		if tagetNode, exists := currNode.children[word]; !exists {
			tagetNode = newSensitiveNode()
			//tagetNode.isEnd = false 默认就是false了
			//因为是之前没有出现过的分支，所以接下来会先将该分支加入到树中，然后再在这条新分支中进行操作
			currNode.children[word] = tagetNode
			currNode = tagetNode
		} else {
			//之前出现过这个分支，所以接下来会进入这个旧的分支进行操作
			currNode = tagetNode
		}
	}

	//添加完毕
	currNode.isEnd = true
}

//查看是否存在敏感词
func (dfaUtil *DfaUtil) contains(txt string) bool {
	var flag = false
	words := []rune(txt)
	currNode := dfaUtil.root
	var matchFlag = 0
	start := -1
	tag := -1

	for i := 0; i < len(words); i++ {
		if _, exists := InvalidWords[string(words[i])]; exists {
			continue
		}

		if targetNode, exists := currNode.children[words[i]]; exists {
			//记录敏感词第一个文字的位置
			tag++
			if tag == 0 {
				start = i
			}
			matchFlag++
			currNode = targetNode
			if currNode.isEnd == true {
				flag = true
				break
			}
		} else {
			//敏感词不全匹配，终止此敏感词查找。从开始位置的第二个文字继续判断
			if start != -1 {
				i = start + 1
			}
			//重置
			currNode = dfaUtil.root
			tag = -1
			start = -1
		}
	}

	//是词语才返回
	if matchFlag <2 || !flag {
		return false
	}

	return true
}

//查找敏感词索引
func (dfaUtil *DfaUtil) searchSensitive(txt string, matchType MATCHTYPE) (matchIndexList []*matchIndex) {
	words := []rune(txt)
	currNode := dfaUtil.root
	start := -1
	tag := -1

	for i := 0; i < len(words); i++ {
		if _, exists := InvalidWords[string(words[i])]; exists {
			continue
		}

		if targetNode, exists := currNode.children[words[i]]; exists {
			//记录敏感词第一个字的位置
			tag++
			if tag == 0 {
				start = i
			}
			currNode = targetNode
			if currNode.isEnd == true {
				matchIndexList = append(matchIndexList, newMatchIndex(start, i))
				if matchType == SINGLE {
					return matchIndexList
				}
				//重置,查找下一个敏感词
				currNode = dfaUtil.root
				tag = -1
				start = -1
			}
		} else {
			//敏感词不全匹配，终止此敏感词查找。从开始位置的第二个文字继续判断
			if start != -1 {
				i = start + 1
			}
			//重置
			currNode = dfaUtil.root
			tag = -1
			start = -1
		}
	}

	return matchIndexList
}

//替换敏感词
func (dfaUtil *DfaUtil) cover(txt string, mask rune) string {
	matchIndexList := dfaUtil.searchSensitive(txt, ALL)
	if len(matchIndexList) == 0 {
		return ""
	}

	txtRune := []rune(txt)
	for _, matchIndexStruct := range matchIndexList {
		for index := matchIndexStruct.start; index <= matchIndexStruct.end; index++ {
			txtRune[index] = mask
		}
	}

	return string(txtRune)
}



