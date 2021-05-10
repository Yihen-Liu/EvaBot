/**
 * Description:
 * Author: Yihen.Liu
 * Create: 2021-05-10
 */
package newcoin

import (
	"github.com/riversgo007/EvaBot/common/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB // it is no need to mind closing action.
)

type Group struct {
	ID     		int    	`json:"id"`
	GroupID 	string 	`json:"group_id"`
	UserAmount 	int 	`json:"user_amount"`
	GroupName 	string  `json:"group_name"`
	BotName 	string 	`json:"bot_name"`
	IsKick 		int8 	`json:"is_kick"`
	CreateTime  int64   `json:"create_time"`
	UpdateTime 	int64 	`json:"update_time"`
}

func initMysql() (err error) {
	//初始化数据库连接
	dns := "eva:evanetwork.org@(localhost:3306)/telegram?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dns), &gorm.Config{})
	if err != nil {
		return
	} else {
		log.Info("create mysql connection successed.")
	}
	return
}
