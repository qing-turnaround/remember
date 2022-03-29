package markdown

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"remember/model"
	"remember/mysql"
	"strings"
	"sync"
	"time"
)

const fileHeader = `---
date: xxxx-xx-xx
description: "you can"
image: "images/about/积跬步以至千里.jpg"
title: "持续行动"
author: 诸葛青
authorEmoji: 😃
pinned: false
tags:
- 
series:
-
---

`

const headerOffset = len(fileHeader)

var (
	tempFile = FilePrefix + "temp" + FileSuffix
	// 一天的时间
	oneDay = time.Hour * 24
	// 回车换行符
	char = "\r\n"
	taskChar = "* "
	categoryChar = "## "
	// 文件路径前缀
	FilePrefix = "D:/gozhugeqing/src/blog/demo/content/en/write/"
	// FilePrefix = "E:/记忆曲线/"
	// 文件名称后缀
	FileSuffix = ".md"
	// 记忆时间，分别是当天，一天，三天后，一个星期后，一个月后
	logicNums = []int{0, 1, 3, 7, 31}
	reader = bufio.NewReader(os.Stdin)
	// 并发同步
	once = sync.WaitGroup{}
	mutex = sync.Mutex{}
)

func FileIsExist(fileName string, date string) bool {
	// 判断文件是否存在
	io, err := os.Open(fileName)
	defer io.Close()
	if err != nil {
		// 如果文件不存在就创建
		CreateFile(fileName, strings.Replace(fileHeader,
			"xxxx-xx-xx", date, 1))
		return false
	}
	return true
}


// writeFile 将任务输入进文件（\r\n为两个字符）
func write(date, fileName, category, task string, fileIsExist bool) {
	defer once.Done() // 减少一次
	// 重新打开文件
	io, _ := os.OpenFile(fileName, os.O_RDWR, 7777)
	defer io.Close()
	newCategory :=  char + categoryChar + category + char
	newTask := taskChar + task + char + char
	// 文件刚刚创建
	if !fileIsExist {
		// 设置 offset
		io.Seek(int64(headerOffset), os.SEEK_SET)
		// 写入文件
		WriteTofile(io, newCategory, newTask)
		// 设置 offset
		io.Seek(0, os.SEEK_SET)
		s, _ := ioutil.ReadAll(io)
		job := &model.Job{
			Date:         date,
			Category:     category,
			Task:         task,
			Offset:       int64(len(s)),
			CategoryRank: 1,
			TaskRank:     1,
		}
		// 写入数据库
		if err := mysql.InsertJob(job); err != nil {
			fmt.Println("向数据库中写入任务错误", err.Error())
		}
	} else {
		// 查询是否存在该分类
		findJob := mysql.FindJobByDateAndCategory(date, category)
		if findJob.ID == 0 {
			// 不存在改分类
			// 查询当天的最大的offset
			oldJob := mysql.FindMaxDateOffset(date)
			// 设置 offset
			io.Seek(oldJob.Offset, os.SEEK_SET)
			// 写入
			WriteTofile(io, newCategory, newTask)
			// 设置 offset
			io.Seek(0, os.SEEK_SET)
			s, _ := ioutil.ReadAll(io)
			job := &model.Job{
				Date:         date,
				Category:     category,
				Task:         task,
				Offset:       int64(len(s)),
				CategoryRank: oldJob.CategoryRank+1,
				TaskRank:     1,
			}
			// 写入数据库
			if err := mysql.InsertJob(job); err != nil {
				fmt.Println("向数据库中写入任务错误", err.Error())
			}
		} else {
			// 已经存在该分类
			// 查看当天该分类最大offset
			oldJob := mysql.FindMaxDateCategoryOffset(date, category)
			// 插入操作
			WriteInsert(io, oldJob.Offset, newTask, fileName)

			addOffset := int64(len(newTask))
			job := &model.Job{
				Date:         date,
				Category:     category,
				Task:         task,
				Offset:       oldJob.Offset + addOffset,
				CategoryRank: oldJob.CategoryRank,
				TaskRank:     oldJob.TaskRank+1,
			}
			// 写入数据库
			if err := mysql.InsertJob(job); err != nil {
				fmt.Println("向数据库中写入任务错误", err.Error())
			}
			// 更新 其他在该分类之下的 offset
			allJob := mysql.FindNotCategoryDownJob(date, oldJob.CategoryRank)
			for _, v := range allJob {
				v.Offset += addOffset
				mysql.UpdateJob(v)
			}
		}
	}

	fmt.Println(fileName)
}


// Logic 如何复习的逻辑
func Logic(local time.Time) {

	// 一直处理用户输入，除非用户主动退出
	for true {
		once.Add(len(logicNums))
		var category, task string
		fmt.Printf("请您输入要记忆任务的类别：")
		category, _ = reader.ReadString('\n')  // 使用 *bufio.Reader 来读取空格句子
		fmt.Printf("请您输入要记忆任务：")
		task , _ = reader.ReadString('\n')  // 使用 *bufio.Reader 来读取空格句子
		for _, v := range logicNums {
			timeString := timeToString(local.Add(time.Duration(v) * oneDay))
			fileName := FilePrefix + timeString + FileSuffix
			fileIsExist := FileIsExist(fileName, timeString)
			go write(timeString, fileName, category, task, fileIsExist)
		}
		once.Wait() // 等待全部执行完成
	}
}



func CreateFile(fileName string, content string){
	io, _ := os.Create(fileName)
	// 添加 fileHeader
	io.WriteString(content)
	defer io.Close()
}

// 向文件中写入
func WriteTofile(io *os.File, s ...string) {
	for i := range s {
		io.WriteString(s[i])
	}
}

// 时间转化为字符串
func timeToString(t time.Time) string {
	return t.String()[:10]
}

func WriteInsert(f1 *os.File, offset int64, content, fileName string) {
	mutex.Lock()
	defer mutex.Unlock()
	// 读取文件的buf
	f1.Seek(0, 0)
	buf := make([]byte, offset)
	f1.Read(buf)
	// 读取之后所有的内容
	oldContent, _ := ioutil.ReadAll(f1)
	f1.Close()

	appendString := string(buf) + content + string(oldContent)
	// fmt.Println(appendString)
	// 创建临时文件（并写入内容）
	CreateFile(tempFile, appendString)

	// 删除原始文件
	if err := os.Remove(fileName); err != nil {
		fmt.Println("os remove file err", err)
	}
	// 将temp.md 命名为原文件
	os.Rename(tempFile, fileName)
}