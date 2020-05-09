package go_sensitive

import (
	"testing"
)

func Test_newSensitiveNode(t *testing.T) {
	words := []string{"草泥马", "狗日", "傻逼"}
	dfaUtil := newDfaUtil(words)
	input := "狗好的草!泥.马hodisho傻逼啊哈哈狗.日的，收到会动手oh配合的菩萨都拍好i啊哦哦爱红i，打死u大概 撒旦啊，傻、逼大撒比u是"

	if dfaUtil.contains(input) {
		t.Log("存在")
	} else {
		t.Log("不存在")
	}

	var matchList []*matchIndex
   	matchList = dfaUtil.searchSensitive(input, ALL)
   	for _, v := range matchList {
   		t.Log(v)
	}

	t.Log(dfaUtil.cover(input, '*'))
}