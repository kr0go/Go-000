package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/pkg/errors"

	_ "github.com/go-sql-driver/mysql"
)

//老师，我是开发萌新
//不知道server层api层是什么概念
//所以就接收消息操作数据库
//后面不必为了照顾萌新而出简单的作业
//不懂的概念我自己学

var connStr string = "xxxxx"
var sqlStr string = "select id from tab"

func conn() (*sql.DB, error) {
	return sql.Open("mysql", connStr)
}

func query() ([]int, error) {
	res := make([]int, 0)
	c, err := conn()
	defer c.Close()
	if err != nil {
		return res, errors.Wrap(err, connStr)
	}

	rows, err := c.Query(sqlStr)
	if err != nil {
		return res, errors.Wrap(err, sqlStr)
	}
	defer rows.Close()

	var i int
	for rows.Next() {
		err := rows.Scan(&i)
		if err != nil {
			return make([]int, 0), errors.Wrap(err, sqlStr)
		}
		res = append(res, i)
	}

	if err := rows.Err(); err != nil {
		return make([]int, 0), errors.Wrap(err, sqlStr)
	}

	return res, nil
}

func worker() (interface{}, error) {
	res, err := query()
	if errors.Is(err, sql.ErrNoRows) {
		return "NULL", errors.WithMessage(err, "return null")
	}
	return res, nil
}

func GO(x func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		x()
	}()
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Println(err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		GO(func() {
			buf := make([]byte, 64)
			i, err := conn.Read(buf)
			if err != nil {
				log.Println(err)
				return
			}
			if i > 0 {
				res, err := worker()
				if err != nil {
					log.Printf(`the worker get error: %v \n`, err)
				}
				log.Println("the worker get result:", res)
			}
			return
		})
	}
}
