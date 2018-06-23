package common

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var DB *sql.DB

func Config() map[string]map[string]string {
	content, err := ioutil.ReadFile(filepath.Join("..", "config.json"))
	if err != nil {
		log.Fatalln(err)
	}
	config := make(map[string]map[string]string)
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatalln(err)
	}
	return config
}

func db() *sql.DB {
	name := "mysql"
	c := Config()["database"]
	source := c["user"] + `:` + c["pass"] + `@tcp(` + c["host"] + `)/` + c["name"] + `?charset=utf8`
	db, err := sql.Open(name, source)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1000)
	return db
}

func init() {
	DB = db()
}

func FetchAll(rows *sql.Rows) (result []map[string]string) {
	defer rows.Close()
	cols, _ := rows.Columns()
	data := make([]interface{}, len(cols))
	buf := make([]interface{}, len(cols))

	for i := range buf {
		buf[i] = &data[i]
	}
	for rows.Next() {
		if err := rows.Scan(buf...); nil != err {
			log.Fatalln(err)
		}
		column := make(map[string]string)
		for k, v := range data {
			if nil == v {
				v = ""
			}
			column[cols[k]] = fmt.Sprintf("%s", v)
		}
		result = append(result, column)
	}
	return
}

func Fetch(rows *sql.Rows) (result map[string]string) {
	defer rows.Close()
	cols, _ := rows.Columns()
	data := make([]interface{}, len(cols))
	buf := make([]interface{}, len(cols))

	for i := range buf {
		buf[i] = &data[i]
	}
	for rows.Next() {
		if err := rows.Scan(buf...); nil != err {
			log.Fatalln(err)
		}
		column := make(map[string]string)
		for k, v := range data {
			if nil == v {
				v = ""
			}
			column[cols[k]] = fmt.Sprintf("%s", v)
		}
		result = column
		break
	}
	return
}

func FetchCol(rows *sql.Rows) (result string) {
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&result); nil != err {
			log.Fatalln(err)
		}
		break
	}
	return
}

var columnIndex = 1

type Select struct {
	table               string
	where, order        []string
	limit, offset       uint
	setLimit, setOffset bool
	columns             map[string]string
}

func (t *Select) From(from string) *Select {
	t.table = from
	return t
}

func (t *Select) Where(where string, params ...interface{}) *Select {
	for _, v := range params {
		m := ""
		switch v.(type) {
		case string:
			where = strings.Replace(where, "?", v.(string), 1)
			m = v.(string)
		case int, int8, int16, int32, int64:
			m = fmt.Sprintf("%d", v)
		case uint, uint8, uint16, uint32, uint64:
			m = fmt.Sprintf("%d", v)
		case float32, float64:
			m = fmt.Sprintf("%f", v)
		case []int:
			var values []string
			for _, n := range v.([]int) {
				values = append(values, strconv.Itoa(n))
			}
			m = strings.Join(values, ", ")
		case []string:
			var values []string
			for _, s := range v.([]string) {
				values = append(values, fmt.Sprintf(`"%s"`, s))
			}
			m = strings.Join(values, ", ")
		}
		if m != "" {
			where = strings.Replace(where, "?", m, 1)
		}
	}
	t.where = append(t.where, where)
	return t
}

func (t *Select) FilterWhere(where string, params ...interface{}) *Select {
	if len(params) != 0 {
		t.Where(where, params...)
	}
	return t
}

func (t *Select) Limit(limit uint) *Select {
	t.limit = limit
	t.setLimit = true
	return t
}

func (t *Select) Offset(offset uint) *Select {
	t.offset = offset
	t.setOffset = true
	return t
}

func (t *Select) Columns(columns interface{}) *Select {
	if t.columns == nil {
		t.columns = make(map[string]string)
	}
	switch columns.(type) {
	case string:
		t.columns[strconv.Itoa(columnIndex)] = columns.(string)
		columnIndex++
	case map[string]string:
		if v, ok := columns.(map[string]string); ok {
			t.columns = v
		}
	}
	return t
}

func (t *Select) Order(field string, orderType ...string) *Select {
	if len(orderType) == 0 {
		orderType = append(orderType, "ASC")
	}
	t.order = append(t.order, strings.Join([]string{field, orderType[0]}, " "))
	return t
}

func (t *Select) renderColumns() string {
	var field string
	var columns []string
	if t.columns == nil {
		t.Columns("*")
	}
	exp, _ := regexp.Compile(`^[\d|\s]+$`)
	for k, v := range t.columns {
		field = v
		if k != "" && !exp.Match([]byte(k)) {
			field = strings.Join([]string{k, v}, " ")
		}
		columns = append(columns, field)
	}
	return strings.Join(columns, ", ")
}

func (t *Select) renderTable() string {
	tables := []string{"FROM", t.table}
	return strings.Join(tables, " ")
}

func (t *Select) renderWhere() string {
	if len(t.where) == 0 {
		return ""
	}
	var wheres []string
	for _, where := range t.where {
		wheres = append(wheres, fmt.Sprintf(`(%s)`, where))
	}
	return fmt.Sprintf(`WHERE %s`, strings.Join(wheres, " AND "))
}

func (t *Select) renderGroup() string {
	return ""
}

func (t *Select) renderHaving() string {
	return ""
}

func (t *Select) renderOrder() string {
	if len(t.order) == 0 {
		return ""
	}
	var orders []string
	for _, v := range t.order {
		orders = append(orders, v)
	}
	return fmt.Sprintf(`ORDER BY %s`, strings.Join(orders, " "))
}

func (t *Select) renderLimit() string {
	if t.setLimit {
		return fmt.Sprintf("LIMIT %d", t.limit)
	}
	return ""
}

func (t *Select) renderOffset() string {
	if t.setOffset {
		return fmt.Sprintf("OFFSET %d", t.offset)
	}
	return ""
}

func (t *Select) String() string {
	sql := []string{"SELECT"}
	sql = append(sql, t.renderColumns())
	sql = append(sql, t.renderTable())

	where := t.renderWhere()
	if where != "" {
		sql = append(sql, where)
	}

	group := t.renderGroup()
	if group != "" {
		sql = append(sql, group)
	}

	having := t.renderHaving()
	if having != "" {
		sql = append(sql, having)
	}

	order := t.renderOrder()
	if order != "" {
		sql = append(sql, order)
	}

	limit := t.renderLimit()
	if limit != "" {
		sql = append(sql, limit)
	}

	offset := t.renderOffset()
	if offset != "" {
		sql = append(sql, offset)
	}
	return strings.Join(sql, " ")
}
