package main

import (
	_ "background/models"
	_ "background/routers"
	"github.com/astaxie/beego"
)

func main() {
	//在beego.Run之前把两个函数对应起来
	beego.AddFuncMap("ShowPrePage", PrePageIndex)
	beego.AddFuncMap("ShowNextPage", NextPage)
	beego.Run()
}

//第二部，在代码里面定义一个函数
func PrePageIndex(pageIndex int) int {
	prePage := pageIndex - 1
	if prePage < 1 {
		prePage = 1
	}
	return prePage
}

//定义一个函数
func NextPage(pageIndex int) int {

	return pageIndex + 1
}
