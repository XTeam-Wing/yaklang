package standard_parser

import (
	"github.com/yaklang/yaklang/common/utils"
)

type fuzztagPos struct {
	tagType    *TagDefine
	start, end int
	subs       []*fuzztagPos
}

func Parse(raw string, tagTypes ...*TagDefine) ([]Node, error) {
	// 去重、提取标签边界符
	utagTypes := []*TagDefine{}
	tagBoundaryMap := map[string]rune{} // 转义符映射
	var tagBoundarys []string
	typeMap := map[string]struct{}{}
	for _, tagType := range tagTypes {
		if _, ok := typeMap[tagType.name]; !ok {
			l := []string{tagType.start, tagType.end}
			if tagType.start == tagType.end {
				l = []string{tagType.start}
			}
			for _, t := range l {
				if _, ok := tagBoundaryMap[t]; !ok {
					tagBoundarys = append(tagBoundarys, t)
				} else {
					return nil, utils.Errorf("tag boundary of tag `%s` conflicts with other tags", tagType.name)
				}
			}
			utagTypes = append(utagTypes, tagType)
		}
	}

	// 查找所有标签位置信息
	tagBoundaryPostions1 := IndexAllSubstrings(raw, tagBoundarys...)
	pre := [2]int{-1, -1}
	tagBoundaryPostions := [][2]int{}
	for _, postion := range tagBoundaryPostions1 {
		if pre[0] != -1 {
			if pre[1] == postion[1] { // 当多个字符串之前存在包含关系时，只保留长的字符串
				//tagBoundaryPostions = append(tagBoundaryPostions, postion)

			} else {
				tagBoundaryPostions = append(tagBoundaryPostions, pre)
			}
		}
		pre = postion
	}
	tagBoundaryPostions = append(tagBoundaryPostions, pre)
	escapeSymbol := `\`
	stack := utils.NewStack[*fuzztagPos]()
	rootTags := []*fuzztagPos{}
	isEscapeMode := false
	nextStart := 0
	for _, pos := range tagBoundaryPostions {
		if pos[1] < nextStart {
			continue
		}
		if isEscapeMode && pos[1] >= len(escapeSymbol) { // 跳过被转义的边界符
			if raw[pos[1]-len(escapeSymbol):pos[1]] == escapeSymbol {
				nextStart = pos[1] + len(tagBoundarys[pos[0]])
				continue
			}
		}
		tagIndex := pos[0] / 2
		isleft := pos[0]%2 == 0
		typ := tagTypes[tagIndex]
		if isleft {
			if stack.Size() != 0 && stack.Peek().tagType.raw {
				continue
			}
			tag := &fuzztagPos{start: pos[1] + len(typ.start), tagType: typ}
			if stack.Size() != 0 {
				top := stack.Peek()
				top.subs = append(top.subs, tag)
			} else {
				rootTags = append(rootTags, tag)
				isEscapeMode = true
			}
			stack.Push(tag)
		} else {
			if stack.Size() != 0 {
				if stack.Peek().tagType.name == typ.name {
					top := stack.Pop()
					top.end = pos[1]
					if stack.Size() == 0 { // 根标签闭合
						isEscapeMode = false
					}
				}
			}
		}
	}
	// 过滤未闭合的标签
	var filterValidTag func(rootTags []*fuzztagPos) (result []*fuzztagPos)
	filterValidTag = func(rootTags []*fuzztagPos) (result []*fuzztagPos) {
		for _, tag := range rootTags {
			if tag.end == 0 { // 无效tag
				result = append(result, filterValidTag(tag.subs)...)
			} else {
				result = append(result, tag)
			}
		}

		return
	}
	escaper := NewDefaultEscaper(escapeSymbol, "{{", "}}")
	// 根据标签位位置信息解析tag
	var newDatasFromPos func(start, end int, poss []*fuzztagPos, deep int) []Node
	newDatasFromPos = func(start, end int, poss []*fuzztagPos, deep int) []Node {
		pre := start
		res := []Node{}
		var addRes func(s Node)
		if deep > 0 {
			addRes = func(s Node) {
				if v, ok := s.(StringNode); ok {
					v1, _ := escaper.Unescape(string(v))
					res = append(res, StringNode(v1))
				} else {
					res = append(res, s)
				}
			}
		} else {
			addRes = func(s Node) {
				res = append(res, s)
			}
		}
		for _, pos := range poss {
			if pos.start < start || pos.end > end { // 不解析参数外的fuzztag
				continue
			}
			tag := pos.tagType.NewTag()
			tag.AddData(newDatasFromPos(pos.start, pos.end, pos.subs, deep+1)...)
			s := raw[pre : pos.start-len(pos.tagType.start)]
			if len(s) != 0 {
				addRes(StringNode(s))
			}
			addRes(tag)
			pre = pos.end + len(pos.tagType.end)
		}
		if pre < end {
			addRes(StringNode(raw[pre:end]))
		}
		return res
	}
	return newDatasFromPos(0, len(raw), filterValidTag(rootTags), 0), nil
}
