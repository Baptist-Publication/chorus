package testdb

import (
	"database/sql"
	"fmt"
	"strings"
	// 引入数据库驱动注册及初始化
	_ "github.com/go-sql-driver/mysql"
)

const (
	userName = "root"
	password = "pass123456"
	ip       = "127.0.0.1"
	port     = "3306"
	dbName   = "testp2p"
)

var DB *sql.DB

func InitDB() (err error) {
	//构建连接："用户名:密码@tcp(IP:端口)/数据库?charset=utf8"
	path := strings.Join([]string{userName, ":", password, "@tcp(", ip, ":", port, ")/", dbName, "?charset=utf8"}, "")

	DB, err = sql.Open("mysql", path)
	if err != nil {
		return
	}
	//设置数据库最大连接数
	DB.SetConnMaxLifetime(100)
	//设置上数据库最大闲置连接数
	DB.SetMaxIdleConns(10)
	//验证连接
	if err = DB.Ping(); err != nil {
		err = fmt.Errorf("opon database fail %v", err)
		return
	}
	stmt, err := DB.Prepare("create table if not exists message(id int AUTO_INCREMENT PRIMARY KEY," +
		"insert_time datetime,chID char(24),msg_hashed text,size int,peer_remote char(128),peer_listen char(128) )")
	if err != nil {
		return
	}
	if stmt != nil {
		_, err = stmt.Exec()
		if err != nil {
			return err
		}
		err = stmt.Close()
	}
	return err
}

//SaveP2pMessage
func SaveP2pMessage(insertTime, chID, msg_hashed string, size int, peerRemote, peerListen string) (err error) {
	//开启事务
	tx, err := DB.Begin()
	if err != nil {
		err = fmt.Errorf("tx fail")
		return
	}
	//准备sql语句
	sql_cmd := `INSERT INTO message (insert_time, chID, msg_hashed,size, peer_remote, peer_listen) VALUES (?, ?, ?, ?, ?, ?)`
	stmt, err := tx.Prepare(sql_cmd)
	if err != nil {
		err = fmt.Errorf("Prepare fail %v", err)
		return
	}

	//将参数传递到sql语句中并且执行,
	res, err := stmt.Exec(insertTime, chID, msg_hashed, size, peerRemote, peerListen)
	if err != nil {
		err = fmt.Errorf("Exec fail")
		return
	}
	//将事务提交
	tx.Commit()
	_, err = res.LastInsertId()
	if err != nil {
		err = fmt.Errorf("fetch last insert id failed:", err.Error())
		return
	}
	return
}
