package controllers

import (
	"background/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"math"
	"path"
	"time"

	"bytes"
	"encoding/gob"
	"github.com/gomodule/redigo/redis"
)

type ArticleController struct {
	beego.Controller
}

//展示文章列表页
func (this *ArticleController) ShowArticleList() {
	userName := this.GetSession("userName")
	if userName == nil {
		this.Redirect("/login", 302)
		return
	}
	this.Data["userName"] = userName.(string)
	//查询数据库，拿出数据，传递给视图
	//获取orm对象
	o := orm.NewOrm()
	//获取查询对象
	var articles []models.GoodsSKU
	//查询
	//queryseter  高级查询使用的数据类型
	qs := o.QueryTable("GoodsSKU")
	//查询所有的文章
	//qs.All(&articles)//select * from article

	//实现分页
	//获取总记录数和总页数
	typeName := this.GetString("select")
	var count int64
	if typeName == "" {
		count, _ = qs.RelatedSel("GoodsType").Count()
	} else {
		count, _ = qs.RelatedSel("GoodsType").Filter("GoodsType__TypeName", typeName).Count()
	}

	pageSize := int64(2)

	pageCount := float64(count) / float64(pageSize)

	pageCount = math.Ceil(pageCount)

	//向上取整
	//把数据传递给视图
	this.Data["count"] = count
	this.Data["pageCount"] = pageCount

	//获取首页末页数据
	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	//获取分页的数据
	start := pageSize * (int64(pageIndex) - 1)
	//RelatedSel 一对多关系表查询中，用来指定另外一张表的函数
	//relatedSel指定表之后，查询的内容都是有这个属性值的数据
	//qs.Limit(pageSize,start).RelatedSel("ArticleType").All(&articles)  //queryseter

	//根据传递的类型获取相应的文章
	//获取数据

	this.Data["typeName"] = typeName
	if typeName == "" {
		qs.Limit(pageSize, start).RelatedSel("GoodsType").All(&articles) //queryseter
	} else {
		qs.Limit(pageSize, start).RelatedSel("GoodsType").Filter("GoodsType__TypeName", typeName).All(&articles) //queryseter
	}

	//获取所有类型
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		beego.Error("redis链接错误", err)
	}
	defer conn.Close()
	reply, err := conn.Do("get", "types")
	byteType, err := redis.Bytes(reply, err)
	if err != nil {
		beego.Error("获取数据错误", err)
	}
	//解码
	var articleTypes []models.GoodsType
	if len(byteType) == 0 {
		o.QueryTable("GoodsType").All(&articleTypes)

		//存储操作
		//编码成字节流
		//编码器
		//定义一个容器，来存储编码之后的字节流
		var buffer bytes.Buffer
		enc := gob.NewEncoder(&buffer)
		enc.Encode(articleTypes)
		beego.Info("从myqsl中获取数据")

		conn.Do("set", "types", buffer.Bytes())
	} else {
		dec := gob.NewDecoder(bytes.NewReader(byteType))
		dec.Decode(&articleTypes)
	}

	this.Data["articleTypes"] = articleTypes
	this.Data["pageIndex"] = pageIndex
	this.Data["articles"] = articles

	this.Layout = "layout.html"
	this.TplName = "index.html"
}

//展示添加文章页面
func (this *ArticleController) ShowAddArticle() {
	//获取所有类型并传递给视图
	//获取orm对象
	//o := orm.NewOrm()
	//
	//var articleTypes []models.GoodsType
	//
	//o.QueryTable("ArticleType").All(&articleTypes)
	//
	////返回给视图
	//this.Data["articleTypes"] = articleTypes
	//this.Layout = "layout.html"
	//this.TplName = "add.html"
}

//处理添加文章业务
func (this *ArticleController) HandeAddArticle() {
	//接受数据
	artileName := this.GetString("articleName")
	content := this.GetString("content")
	//校验数据
	if artileName == "" || content == "" {
		this.Data["errmsg"] = "文章标题或内容不能为空"
		this.TplName = "add.html"
		return
	}

	typeName := this.GetString("select")

	//接收图片
	file, head, err := this.GetFile("uploadname")
	if err != nil {
		this.Data["errmsg"] = "获取文件失败"
		this.TplName = "add.html"
		return
	}
	defer file.Close()
	//1.判断文件大小
	if head.Size > 500000 {
		this.Data["errmsg"] = "文件太大，上传失败"
		this.TplName = "add.html"
		return
	}

	//2.判断图片格式
	//1.jpg
	fileExt := path.Ext(head.Filename)
	if fileExt != ".jpg" && fileExt != ".png" && fileExt != ".jpeg" {
		this.Data["errmsg"] = "文件格式不正确，请重新上传"
		this.TplName = "add.html"
		return
	}

	//3.文件名防止重复
	fileName := time.Now().Format("2006-01-02-15-04-05") + fileExt
	this.SaveToFile("uploadname", "./static/image/"+fileName)

	//处理数据
	//数据库的插入操作
	//获取orm对象
	o := orm.NewOrm()
	//获取插入对象
	var article models.GoodsSKU
	//给插入对象赋值
	article.Name = artileName
	article.Desc = content
	article.Image = "/static/image/" + fileName

	//根据类型名称获取类型对象
	var articleType models.GoodsType
	articleType.Name = typeName
	o.Read(&articleType, "TypeName")
	article.GoodsType = &articleType

	//插入
	_, err = o.Insert(&article)
	if err != nil {
		this.Data["errmsg"] = "添加文章失败，请重新添加"
		this.TplName = "add.html"
		return
	}
	//返回页面
	this.Redirect("/article/articleList", 302)
}

//展示文章详情页
func (this *ArticleController) ShowArticleDetail() {
	////获取数据
	//articleId, err := this.GetInt("id")
	////校验数据
	//if err != nil {
	//	this.Data["errmsg"] = "请求路径错误"
	//	this.TplName = "index.html"
	//	return
	//}
	//
	////处理数据
	////查询数据
	////获取orm对象
	//o := orm.NewOrm()
	////获取查询对象
	//var article models.Article
	////给查询条件赋值
	//article.Id = articleId
	////查询
	//err = o.Read(&article)
	//if err != nil {
	//	this.Data["errmsg"] = "请求路径错误"
	//	this.TplName = "index.html"
	//	return
	//}
	//
	////获取article对象   知道向哪里插入数据
	//
	////获取多对多操作对象   知道插入到对象的哪个字段
	//m2m := o.QueryM2M(&article, "Users")
	////第三步,获取要插入的数据   知道插入什么数据
	//var user models.User
	//userName := this.GetSession("userName")
	//user.UserName = userName.(string)
	//o.Read(&user, "UserName")
	//
	////插入多对多关系
	//m2m.Add(user)
	//
	////第一种多对多查询
	//o.LoadRelated(&article, "Users")
	//
	//////第二种多对多关系查询   正向插入，反向查询
	//////filter  过滤器  指定查询条件，进行过滤查找
	//var users []models.User
	//
	////select * from user                     where article.Id == articleId
	//o.QueryTable("User").Filter("Articles__Article__Id", articleId).Distinct().All(&users)
	//this.Data["users"] = users
	////返回数据
	//this.Data["article"] = article
	//this.TplName = "content.html"
}

//展示编辑文章页面
func (this *ArticleController) ShowUpdateArticle() {
	////获取数据
	//articleId, err := this.GetInt("id")
	//
	//errmsg := this.GetString("errmsg")
	//if errmsg != "" {
	//	this.Data["errmsg"] = errmsg
	//}
	////校验数据
	//if err != nil {
	//	beego.Error("请求路径错误")
	//	this.Redirect("/article/articleList?errmsg", 302)
	//	return
	//}
	////数据处理
	////查询操作
	////获取orm对象
	//o := orm.NewOrm()
	////获取查询对象
	//var article models.GoodsSKU
	////给查询条件赋值
	//article.Id = articleId
	////查询
	//o.Read(&article)
	////返回数据
	//this.Data["article"] = article
	//this.TplName = "update.html"
}

//文件上传函数
func UploadFile(this *ArticleController, filePath string) string {
	//接收图片
	file, head, err := this.GetFile(filePath)
	if err != nil {
		this.Data["errmsg"] = "获取文件失败"
		this.TplName = "add.html"
		return ""
	}
	defer file.Close()
	//1.判断文件大小
	if head.Size > 500000 {
		this.Data["errmsg"] = "文件太大，上传失败"
		this.TplName = "add.html"
		return ""
	}

	//2.判断图片格式
	//1.jpg
	fileExt := path.Ext(head.Filename)
	if fileExt != ".jpg" && fileExt != ".png" && fileExt != ".jpeg" {
		this.Data["errmsg"] = "文件格式不正确，请重新上传"
		this.TplName = "add.html"
		return ""
	}

	//3.文件名防止重复
	fileName := time.Now().Format("2006-01-02-15-04-05") + fileExt
	this.SaveToFile(filePath, "./static/image/"+fileName)
	return "/static/image/" + fileName

}

//处理编辑文章业务
func (this *ArticleController) HandleUpdateArticle() {
	////获取数据
	//articleName := this.GetString("articleName")
	//content := this.GetString("content")
	//fileName := UploadFile(this, "uploadname")
	//articleId, err2 := this.GetInt("id")
	////校验数据
	//if articleName == "" || content == "" || fileName == "" || err2 != nil {
	//	errmsg := "内容不能为空"
	//	this.Redirect("/article/updateArticle?id="+strconv.Itoa(articleId)+"&errmsg="+errmsg, 302)
	//	return
	//}
	////处理数据
	////update  更新操作
	////获取orm对象
	//o := orm.NewOrm()
	////获取更新对象
	//var article models.Article
	////给更新对象赋值
	//article.Id = articleId
	//err := o.Read(&article)
	//if err != nil {
	//	errmsg := "更新文章不存在"
	//	this.Redirect("/article/updateArticle?id="+strconv.Itoa(articleId)+"&errmsg="+errmsg, 302)
	//	return
	//}
	////给更新字段赋新值
	//article.Title = articleName
	//article.Content = content
	//article.Image = fileName
	////更新
	//o.Update(&article)
	//
	////返回数据
	//this.Redirect("/article/articleList", 302)
}

//删除业务处理
func (this *ArticleController) DeleteArticle() {
	////获取数据
	//articleId, err := this.GetInt("id")
	////校验数据
	//if err != nil {
	//	beego.Error("路径错误")
	//	this.Redirect("/article/articleList", 302)
	//	return
	//}
	////处理数据
	////删除操作
	////获取orm对象
	//o := orm.NewOrm()
	////获取删除对象
	//var article models.Article
	////给删除对象赋值
	//article.Id = articleId
	////删除
	//_, err = o.Delete(&article)
	//if err != nil {
	//	beego.Error("删除失败")
	//	this.Redirect("/article/articleList", 302)
	//	return
	//}
	//
	////返回数据
	//this.Redirect("/article/articleList", 302)
}

//展示添加类型界面
func (this *ArticleController) ShowAddType() {
	////获取所有类型数据，并展示
	////获取orm对象
	//o := orm.NewOrm()
	////查询容器
	//var articleTypes []models.ArticleType
	////指定查询表
	//qs := o.QueryTable("ArticleType")
	//qs.All(&articleTypes)
	//
	////返回数据给视图
	//this.Data["articleTypes"] = articleTypes
	this.Layout = "layout.html"
	this.TplName = "addType.html"
}

//处理类型添加业务
func (this *ArticleController) HandleAddType() {
	//获取数据
	typeName := this.GetString("typeName")
	// 添加商品类型logo
	logoPath := UploadFile(this, "uploadlogo")
	// 添加商品类型图片
	typePath := UploadFile(this, "uploadTypeImage")

	//校验数据
	if typeName == "" || logoPath == "" || typePath == "" {
		this.Data["errmsg"] = "类型名不能为空"
		this.Redirect("/article/addType", 302)
		return
	}

	//处理数据
	//插入操作
	//获取orm对象
	o := orm.NewOrm()
	//获取插入对象
	var articleType models.GoodsType
	//给插入对象赋值
	articleType.Name = typeName
	articleType.Logo = logoPath
	articleType.Image = typePath
	//插入
	_, err := o.Insert(&articleType)
	if err != nil {
		this.Data["errmsg"] = "文章类型添加失败"
		this.TplName = "addType.html"
		return
	}

	//返回数据
	this.Redirect("/article/addType", 302)
}

//删除类型
func (this *ArticleController) DeleteType() {
	////获取数据
	//typeId, err := this.GetInt("id")
	////校验数据
	//if err != nil {
	//	beego.Error("删除失败")
	//	this.Redirect("/article/addType", 302)
	//	return
	//}
	//
	////处理数据u
	////删除操作
	////获取orm对象
	//o := orm.NewOrm()
	////获取删除对象
	//var articleType models.ArticleType
	////给产出对象赋值
	//articleType.Id = typeId
	////删除
	//_, err = o.Delete(&articleType)
	//if err != nil {
	//	beego.Error("删除失败")
	//	this.Redirect("/article/addType", 302)
	//	return
	//}
	//
	////返回数据
	//this.Redirect("/article/addType", 302)
}
