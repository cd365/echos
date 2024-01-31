package echos

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Int64JsonString int64

func (s *Int64JsonString) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%v"`, s)), nil
}

func (s *Int64JsonString) UnmarshalJSON(b []byte) error {
	val, err := strconv.ParseInt(string(b[1:len(b)-1]), 10, 64)
	if err != nil {
		return err
	}
	*s = Int64JsonString(val)
	return nil
}

func (s *Int64JsonString) Int64() int64 {
	return int64(*s)
}

// QueryIds id list
type QueryIds struct {
	Ids []*Int64JsonString `json:"ids" query:"ids" form:"ids" validate:"omitempty,min=0,max=500"` // id列表,如: ["0", "2023"]
}

func (s QueryIds) GetIdsInt64() []int64 {
	length := len(s.Ids)
	result := make([]int64, length)
	for i := 0; i < length; i++ {
		result[i] = s.Ids[i].Int64()
	}
	return result
}

func NewInt64JsonString(i int64) *Int64JsonString {
	tmp := Int64JsonString(i)
	return &tmp
}

func ToInt64JsonString(s []int64) []*Int64JsonString {
	length := len(s)
	result := make([]*Int64JsonString, length)
	for i := 0; i < length; i++ {
		result[i] = NewInt64JsonString(s[i])
	}
	return result
}

// QueryKeyword 搜索关键字
type QueryKeyword struct {
	Keyword *string `json:"keyword" query:"keyword" form:"keyword" validate:"omitempty,min=1,max=32"` // 检索关键字
}

func (s QueryKeyword) GetKeyword() string {
	return fmt.Sprintf("%%%s%%", *s.Keyword)
}

var (
	// regexpQueryOrder 校验GET请求的query参数order
	regexpQueryOrder = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9_]{0,29}):([ad])$`)
)

// QueryOrder 数据列表排序
type QueryOrder struct {
	Order string `json:"order" query:"order" form:"order" validate:"omitempty,order,max=256"` // 排序表达式,例: weight:a,height:d,age:a
}

// GetOrder 传入可以排序的字段列表,根据客户端传入的排序参数值,构建排序参数列表
// columns value like: []string{"a.uid", "username"}, return value like: a.uid:d,username:a
func (s QueryOrder) GetOrder(columns ...string) string {
	cm := make(map[string]string)
	for _, v := range columns {
		start := strings.LastIndex(v, ".")
		if start == -1 {
			cm[v] = v
			continue
		}
		cm[v[start+1:]] = v
	}
	orders := strings.Split(s.Order, ",")
	result := make([]string, 0, len(orders))
	for _, v := range orders {
		match := regexpQueryOrder.FindAllStringSubmatch(v, -1)
		length := len(match)
		if length != 1 {
			continue
		}
		matched := match[0]
		length = len(matched)
		if length != 3 {
			continue
		}
		column, ok := cm[matched[1]]
		if !ok {
			continue
		}
		// matched[0] == v
		if matched[length-1][0] == 97 || matched[length-1][0] == 100 {
			result = append(result, fmt.Sprintf("%s:%s", column, matched[length-1]))
		}
	}
	return strings.Join(result, ",")
}

// QueryLimit 控制数据列表返回的数据;如果数据超过offset的最大值,应该考虑换个方式查询分页数据如: SELECT id, name FROM ( SELECT id, name FROM account WHERE ( id < 5000000 ) ORDER BY id DESC )
type QueryLimit struct {
	Limit  int64 `json:"limit" query:"limit" form:"limit" validate:"omitempty,min=1,max=1000"`      // 数据条数
	Offset int64 `json:"offset" query:"offset" form:"offset" validate:"omitempty,min=0,max=100000"` // 数据偏移量
}

func (s QueryLimit) GetLimit() int64 {
	if s.Limit <= 0 {
		return 1
	}
	return s.Limit
}

func (s QueryLimit) GetOffset() int64 {
	if s.Offset < 0 {
		return 0
	}
	return s.Offset
}

// QueryMaxId 最大索引值,用于查询(表中数据条数超过QueryLimit.Offset值时的查询) 根据索引值倒序查询
// SELECT id, name FROM ( SELECT id, name FROM account WHERE ( id < 5000000 ) ORDER BY id DESC )
type QueryMaxId struct {
	MaxId *int64 `json:"max_id,string" query:"max_id" form:"max_id" validate:"omitempty"` // 最大id值,不包含当前值
}

// QueryMinId 最小索引值,用于查询(表中数据条数超过QueryLimit.Offset值时的查询) 根据索引值顺序查询
// SELECT id, name FROM ( SELECT id, name FROM account WHERE ( id > 5000000 ) ORDER BY id DESC )
type QueryMinId struct {
	MinId *int64 `json:"min_id,string" query:"min_id" form:"mid_id" validate:"omitempty"` // 最小id值,不包含当前值
}
