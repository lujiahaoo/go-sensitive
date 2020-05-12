package go_sensitive

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	INVALID_WORDS = " ,~,!,@,#,$,%,^,&,*,(,),_,-,+,=,?,<,>,.,—,，,。,/,\\,|,《,》,？,;,:,：,',‘,；,“,！,。,；,：,’,{,},【,】,[,],、"
	SENSITIVE_CHILDRED_SIZE = 128

	LEXICON_PATH = "./lexicon" //todo:根据项目文件结构来修改该词库目录路径
)

//匹配程度
type MATCHTYPE int

const (
	SINGLE MATCHTYPE = iota
	ALL
)

var InvalidWords = make(map[string]interface{})
var SensitiveWords = make([]string, 20000)
var Util  *DFAUtil

func Setup() {
	//加载无效词汇
	inValidArr := strings.Split(INVALID_WORDS, ",")
	for _, v := range inValidArr {
		InvalidWords[v] = nil
	}

	//加载敏感词文件
	var fileList []string
	dir, err := ioutil.ReadDir(LEXICON_PATH)
	if err != nil {
		panic(err)
	}

	for _, fi := range dir {
		if fi.IsDir() {
			continue
		} else {
			fileList = append(fileList, filepath.Join(LEXICON_PATH, fi.Name()))
		}
	}

	if len(fileList) == 0 {
		panic("请添加敏感词文件")
	}

	for _, fileName := range fileList {
		r, _ := os.Open(fileName)
		defer r.Close()

		s := bufio.NewScanner(r)
		for s.Scan() {
			SensitiveWords = append(SensitiveWords, s.Text())
		}
	}

	//装填敏感词
	dfaUtil := &DFAUtil{
		root: newSensitiveNode(),
	}

	for _, word := range SensitiveWords {
		wordRuneList := []rune(word)
		//是词语才加入
		if len(wordRuneList) > 1 {
			dfaUtil.addWord(wordRuneList)
		}
	}


	Util = dfaUtil
}

type sensitiveNode struct {
	isEnd bool

	children map[rune]*sensitiveNode
}

//初始化Trie树
func newSensitiveNode() *sensitiveNode {
	return &sensitiveNode{
		isEnd:    false,
		children: make(map[rune]*sensitiveNode, SENSITIVE_CHILDRED_SIZE),
	}
}

type DFAUtil struct {
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

func NewDFAUtil(wordList []string) *DFAUtil {
	return dfaUtil
}

//添加敏感词汇
func (dfaUtil *DFAUtil) AddWord(words []rune) {
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
func (dfaUtil *DFAUtil) Contains(txt string) bool {
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
			//记录敏感词第一个字的位置
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
				i = start
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
func (dfaUtil *DFAUtil) SearchSensitive(txt string, matchType MATCHTYPE) (matchIndexList []*matchIndex) {
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
				i = start
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
func (dfaUtil *DFAUtil) Cover(txt string, mask rune) (string, bool) {
	matchIndexList := dfaUtil.searchSensitive(txt, ALL)
	if len(matchIndexList) == 0 {
		return txt, false
	}

	txtRune := []rune(txt)
	for _, matchIndexStruct := range matchIndexList {
		for index := matchIndexStruct.start; index <= matchIndexStruct.end; index++ {
			txtRune[index] = mask
		}
	}

	return string(txtRune), true
}



