// Copyright 2016 The kingshard Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package router

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v2"

	"kingshard/config"
	"kingshard/proxy/database"
	"kingshard/sqlparser"
)

func NewRouterT(cfg *config.DatabaseConfig) *Router {
	rt := new(Router)
	rt.Rules = make(map[string]map[string]*Rule)
	rt.Databases = map[string]*database.Database{
		"upyun": &database.Database{Cfg: cfg},
	}
	return rt
}

func TestParseRule(t *testing.T) {
	var s = `
schema:
  databases:
    -
      db: upyun
      user: root1
      password: root1
      nodes: [node2,node3]
  shard:
    -
      db: upyun
      table: test_shard_hash
      key: id
      nodes: [node2, node3]
      locations: [16,16]
      type: hash
    -
      db: upyun
      table: test_shard_range
      key: id
      type: range
      nodes: [node2, node3]
      locations: [16,16]
      table_row_limit: 10000
`
	var cfg config.Config
	if err := yaml.Unmarshal([]byte(s), &cfg); err != nil {
		t.Fatal(err)
	}

	rt := NewRouterT(&cfg.Schema.Databases[0])
	err := rt.ParseRules(&cfg.Schema)
	if err != nil {
		t.Fatal(err)
	}

	hashRule := rt.GetRule("upyun", "test_shard_hash")
	if hashRule.Type != HashRuleType {
		t.Fatal(hashRule.Type)
	}

	if len(hashRule.Nodes) != 2 || hashRule.Nodes[0] != "node2" || hashRule.Nodes[1] != "node3" {
		t.Fatal("parse nodes not correct.")
	}

	if n, _ := hashRule.FindNode(uint64(11)); n != "node2" {
		t.Fatal(n)
	}

	rangeRule := rt.GetRule("upyun", "test_shard_range")
	if rangeRule.Type != RangeRuleType {
		t.Fatal(rangeRule.Type)
	}

	if n, _ := rangeRule.FindNode(10000 - 1); n != "node2" {
		t.Fatal(n)
	}
}

func newTestRouter() *Router {
	var s = `
schema :
  databases:
    -
      db: upyun
      user: root1
      password: root1
      nodes: [node1, node2,node3]
  shard:
    -
      db: upyun
      table: test1
      key: id
      nodes: [node1,node2,node3]
      locations: [4,4,4]
      type: hash

    -
      db: upyun
      table: test2
      key: id
      type: range
      nodes: [node1,node2,node3]
      locations: [4,4,4]
      table_row_limit: 10000
    -
      db: upyun
      table: test_shard_year
      key: date
      nodes: [node2, node3]
      date_range: [2012-2015,2016-2018]
      type: date_year
    -
      db: upyun
      table: test_shard_month
      key: date
      type: date_month
      nodes: [node2, node3]
      date_range: [201512-201603,201604-201608]
    -
      db: upyun
      table: test_shard_day
      key: date
      type: date_day
      nodes: [node2, node3]
      date_range: [20151201-20160122,20160202-20160308]
`

	cfg, err := config.ParseConfigData([]byte(s))
	if err != nil {
		println(err.Error())
		panic(err)
	}

	rt := NewRouterT(&cfg.Schema.Databases[0])

	err = rt.ParseRules(&cfg.Schema)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	return rt
}

//TODO YYYY-MM-DD HH:MM:SS,YYYY-MM-DD test
func TestParseDateRule(t *testing.T) {
	var s = `
schema:
  databases:
    -
      db: upyun
      user: root1
      password: root1
      nodes: [node1, node2,node3]
  shard:
    -
      db: upyun
      table: test_shard_year
      key: date
      nodes: [node2, node3]
      date_range: [2012-2015,2016-2018]
      type: date_year
    -
      db: upyun
      table: test_shard_month
      key: date
      type: date_month
      nodes: [node2, node3]
      date_range: [201512-201603,201604-201608]
    -
      db: upyun
      table: test_shard_day
      key: date
      type: date_day
      nodes: [node2, node3]
      date_range: [20151201-20160122,20160202-20160308]
`
	var cfg config.Config
	if err := yaml.Unmarshal([]byte(s), &cfg); err != nil {
		t.Fatal(err)
	}

	rt := NewRouterT(&cfg.Schema.Databases[0])
	err := rt.ParseRules(&cfg.Schema)
	if err != nil {
		t.Fatal(err)
	}
	yearRule := rt.GetRule("upyun", "test_shard_year")
	if yearRule.Type != DateYearRuleType {
		t.Fatal(yearRule.Type)
	}

	if len(yearRule.Nodes) != 2 || yearRule.Nodes[0] != "node2" || yearRule.Nodes[1] != "node3" {
		t.Fatal("parse nodes not correct.")
	}

	if n, _ := yearRule.FindNode(1457082679); n != "node3" {
		t.Fatal(n)
	}

	monthRule := rt.GetRule("upyun", "test_shard_month")
	if monthRule.Type != DateMonthRuleType {
		t.Fatal(monthRule.Type)
	}

	if n, _ := monthRule.FindNode(1457082679); n != "node2" {
		t.Fatal(n)
	}

	dayRule := rt.GetRule("upyun", "test_shard_day")
	if dayRule.Type != DateDayRuleType {
		t.Fatal(monthRule.Type)
	}

	if n, _ := dayRule.FindNode(1457082679); n != "node3" {
		t.Fatal(n)
	}

}

func newTestDBRule() *Router {
	var s = `
schema :
  databases:
    -
      db: upyun
      user: root1
      password: root1
      nodes: [node1, node2,node3]
  shard:
    -
      db: upyun
      table: test1
      key: id
      nodes: [node1,node2,node3]
      locations: [1,2,3]
      type: hash

    -
      db: upyun
      table: test2
      key: id
      type: range
      nodes: [node1,node2,node3]
      locations: [8,8,8]
      table_row_limit: 100
`

	cfg, err := config.ParseConfigData([]byte(s))
	if err != nil {
		println(err.Error())
		panic(err)
	}

	r := NewRouterT(&cfg.Schema.Databases[0])
	err = r.ParseRules(&cfg.Schema)
	if err != nil {
		panic(err)
	}
	if err != nil {
		println(err.Error())
		panic(err)
	}

	return r
}

func TestBadUpdateExpr(t *testing.T) {
	var sql string
	var db string
	r := newTestDBRule()
	db = "upyun"
	sql = "insert into test1 (id) values (5) on duplicate key update  id = 10"

	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		t.Fatal(err.Error())
	}

	if _, err := r.BuildPlan(db, "node1", stmt); err == nil {
		t.Fatal("must err")
	}

	sql = "update test1 set id = 10 where id = 5"

	stmt, err = sqlparser.Parse(sql)
	if err != nil {
		t.Fatal(err.Error())
	}

	if _, err := r.BuildPlan(db, "node1", stmt); err == nil {
		t.Fatal("must err")
	}
}

func isListEqual(l1 []int, l2 []int) bool {
	var i, j int
	if len(l1) != len(l2) {
		return false
	}
	if len(l1) == 0 {
		return true
	}
	for i = 0; i < len(l1); i++ {
		for j = 0; j < len(l2); j++ {
			if l1[i] == l2[j] {
				break
			}
		}
		if j == len(l2) {
			return false
		}
	}
	return true
}

func checkPlan(t *testing.T, sql string, tableIndexs []int, nodeIndexs []int) {
	r := newTestRouter()
	db := "upyun"
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		t.Fatal(err.Error())
	}
	plan, err := r.BuildPlan(db, "node1", stmt)
	if err != nil {
		t.Fatal(err.Error())
	}

	if isListEqual(plan.RouteTableIndexs, tableIndexs) == false {
		err := fmt.Errorf("RouteTableIndexs=%v but tableIndexs=%v",
			plan.RouteTableIndexs, tableIndexs)
		t.Fatal(err.Error())
	}
	if isListEqual(plan.RouteNodeIndexs, nodeIndexs) == false {
		err := fmt.Errorf("RouteNodeIndexs=%v but nodeIndexs=%v",
			plan.RouteNodeIndexs, nodeIndexs)
		t.Fatal(err.Error())
	}
	t.Logf("rewritten_sql=%v", plan.RewrittenSqls)

}
func TestWhereInPartitionByTableIndex(t *testing.T) {
	var sql string
	t1 := makeList(0, 12)

	//2016-08-02 13:37:26
	sql = "select * from test1 where id in (1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22) "
	checkPlan(t, sql,
		t1,
		[]int{0, 1, 2},
	)
	// ensure no impact for or operator in where
	sql = "select * from test1 where id in (1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21) or name='test'"
	checkPlan(t, sql,
		t1,
		[]int{0, 1, 2},
	)

	// ensure no impact for not in
	sql = "select * from test1 where id not in (0,1,2,3,4,5,6,7)"
	checkPlan(t, sql,
		t1,
		[]int{0, 1, 2})

}

func TestDatePlan(t *testing.T) {
	var sql string
	//2016-03-06 13:37:26
	sql = "select * from test_shard_year where date > 1457242646 "
	checkPlan(t, sql,
		[]int{2016, 2017, 2018},
		[]int{1},
	)

	//2012-03-06 13:37:26
	sql = "select * from test_shard_year where date < 1331012246 "
	checkPlan(t, sql,
		[]int{2012},
		[]int{0},
	)

	//2015-03-06 13:37:26
	sql = "select * from test_shard_year where date > '2015-03-06 13:37:26' "
	checkPlan(t, sql,
		[]int{2015, 2016, 2017, 2018},
		[]int{0, 1},
	)

	//2015-03-06 13:37:26
	sql = "select * from test_shard_year where date <= '2015-03-06' "
	checkPlan(t, sql,
		[]int{2012, 2013, 2014, 2015},
		[]int{0},
	)

	//2015-03-06 13:37:26
	sql = "select * from test_shard_month where date <= 1459921046 "
	checkPlan(t, sql,
		[]int{201512, 201601, 201602, 201603, 201604},
		[]int{0, 1},
	)

	//2015-3-6 13:37:26
	sql = "select * from test_shard_month where date > '2016-05-06' "
	checkPlan(t, sql,
		[]int{201605, 201606, 201607, 201608},
		[]int{1},
	)

	//2016-05-07 12:23:56
	sql = "select * from test_shard_month where date = '2016-05-07 12:23:56' "
	checkPlan(t, sql,
		[]int{201605},
		[]int{1},
	)

	//2016-03-07 12:23:56
	sql = "select * from test_shard_day where date = '2016-03-07 12:23:56' "
	checkPlan(t, sql,
		[]int{20160307},
		[]int{1},
	)

	//2016-03-07 12:23:56
	sql = "select * from test_shard_day where date > '2016-03-07' "
	checkPlan(t, sql,
		[]int{20160307, 20160308},
		[]int{1},
	)

	//2016-03-07 12:23:56
	sql = "select * from test_shard_day where date > 1457242646 "
	checkPlan(t, sql,
		[]int{20160306, 20160307, 20160308},
		[]int{1},
	)
}

func TestSelectPlan(t *testing.T) {
	var sql string
	t1 := makeList(0, 12)

	sql = "select/*master*/ * from test1 where id = 5"
	checkPlan(t, sql, []int{5}, []int{1}) //table_5 node1

	sql = "select * from test1 where id in (5, 8)"
	checkPlan(t, sql, []int{5, 8}, []int{1, 2})

	sql = "select * from test1 where id > 5"

	checkPlan(t, sql, t1, []int{0, 1, 2})

	sql = "select * from test1 where id in (5, 6) and id in (5, 6, 7)"
	checkPlan(t, sql, []int{5, 6}, []int{1})

	sql = "select * from test1 where id in (5, 6) or id in (5, 6, 7,8)"
	checkPlan(t, sql, []int{5, 6, 7, 8}, []int{1, 2})

	sql = "select * from test1 where id not in (5, 6) or id in (5, 6, 7,8)"
	checkPlan(t, sql, t1, []int{0, 1, 2})

	sql = "select * from test1 where id not in (5, 6)"
	checkPlan(t, sql, t1, []int{0, 1, 2})

	sql = "select * from test1 where id in (5, 6) or (id in (5, 6, 7,8) and id in (1,5,7))"
	checkPlan(t, sql, []int{5, 6, 7}, []int{1})

	sql = "select * from test2 where id = 10000"
	checkPlan(t, sql, []int{1}, []int{0})

	sql = "select * from test2 where id between 10000 and 20000"
	checkPlan(t, sql, []int{1, 2}, []int{0})

	sql = "select * from test2 where id not between 1000 and 100000"
	checkPlan(t, sql, []int{0, 10, 11}, []int{0, 2})

	sql = "select * from test2 where id > 10000"
	checkPlan(t, sql, makeList(1, 12), []int{0, 1, 2})

	sql = "select * from test2 where id >= 9999"
	checkPlan(t, sql, t1, []int{0, 1, 2})

	sql = "select * from test2 where id <= 10000"
	checkPlan(t, sql, []int{0, 1}, []int{0})

	sql = "select * from test2 where id < 10000"
	checkPlan(t, sql, []int{0}, []int{0})

	sql = "select * from test2 where id >= 10000 and id <= 30000"
	checkPlan(t, sql, []int{1, 2, 3}, []int{0})

	sql = "select * from test2 where (id >= 10000 and id <= 30000) or id < 100"
	checkPlan(t, sql, []int{0, 1, 2, 3}, []int{0})

	sql = "select * from test2 where id in (1, 10000)"
	checkPlan(t, sql, []int{0, 1}, []int{0})

	sql = "select * from test2 where id not in (1, 10000)"
	checkPlan(t, sql, makeList(0, 12), []int{0, 1, 2})
}

func TestValueSharding(t *testing.T) {
	var sql string

	sql = "insert into test1 (id) values (5)"
	checkPlan(t, sql, []int{5}, []int{1})

	sql = "insert into test2 (id) values (10000)"
	checkPlan(t, sql, []int{1}, []int{0})

	sql = "insert into test2 (id) values (20000)"
	checkPlan(t, sql, []int{2}, []int{0})

	sql = "update test1 set a =10 where id =12"
	checkPlan(t, sql, []int{0}, []int{0})

	sql = "update test2 set a =10 where id < 30000 and 10000< id"
	checkPlan(t, sql, []int{1, 2}, []int{0})

	sql = "delete from test2 where id < 30000 and 10000< id"
	checkPlan(t, sql, []int{1, 2}, []int{0})

	sql = "replace into test1(id) values(5)"
	checkPlan(t, sql, []int{5}, []int{1})
}
