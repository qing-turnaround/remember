package model

type Job struct {
	ID int64 `gorm:"primary_key; auto_increment"`
	// 日期
	Date     string
	Category string
	Task     string
	Offset   int64
	CategoryRank int64
	TaskRank int64
}
