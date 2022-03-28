package mysql

import (
	"fmt"
	"remember/model"
	"remember/settings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

// Init Mysql数据库配置加载
func Init() (err error) {
	// 初始化 mysql
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		settings.Config.User,
		settings.Config.Password,
		settings.Config.Host,
		settings.Config.Port,
		settings.Config.Dbname,
	)
	db, err = gorm.Open("mysql", dsn)
	if err != nil {
		return err
	}
	// 让grom转义struct名字的时候不用加上s
	db.SingularTable(true)
	// 创建表，只需要一次
	// if err = db.CreateTable(&model.Job{}).Error; err != nil {
	// 	fmt.Println("Error creating table", err.Error())
	// }

	return nil
}

func FindJobByDateAndCategory(date string, category string) *model.Job {
	job := &model.Job{}
	db.Where(model.Job{Date: date, Category: category}).First(job)
	return job
}

func InsertJob(job *model.Job) error {
	return db.Create(job).Error
}

func FindMaxDateOffset(date string) *model.Job {
	job := &model.Job{}
	db.Order("offset DESC").
		Where("date = ?", date).Limit(1).Find(job)
	return job
}

func FindMaxDateCategoryOffset(date string, category string) *model.Job {
	job := &model.Job{}
	db.Order("offset DESC").
		Where("date = ? and category = ?", date, category).Limit(1).Find(job)
	return job
}

func FindNotCategoryDownJob(date string, categoryRank int64) []model.Job {
	var allJob []model.Job
	db.Where("date = ? and category_rank > ?", date, categoryRank).
		Find(&allJob)
	return allJob
}

func UpdateJob(job model.Job) error {
	return db.Model(job).Update(job).Error
}

func Close() {
	_ = db.Close()
}
