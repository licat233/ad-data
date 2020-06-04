package router

import (
	"ad-data/addata"
	"ad-data/config"
	"ad-data/webstatistic"

	//"ad-data/mywebsocket"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()
	// 创建基于cookie的存储引擎，secret 参数是用于加密的密钥
	store := cookie.NewStore([]byte("secret"))
	// 设置session中间件，参数mysession，指的是session的名字，也是cookie的名字
	// store是前面创建的存储引擎，我们可以替换成其他存储引擎
	r.Use(sessions.Sessions("addatasession", store))

	r.Delims("{[{", "}]}")
	r.Static("/static", "./static")
	r.LoadHTMLGlob("./view/*")
	r.GET("/", func(c *gin.Context) {
		if err := LoginVerify(c); err != nil {
			return
		}
		c.HTML(http.StatusOK, "index.html", nil)
		//mywebsocket.WsHandler(c.Writer,c.Request)
	})

	api := r.Group("/api")
	{
		api.POST("/projectdata", func(c *gin.Context) { projectdataapi(c) })
		api.POST("/login", func(c *gin.Context) { loginapi(c) })
		api.POST("/todaydata", func(c *gin.Context) { todaydata(c) })
		api.POST("/PJdatabyID", func(c *gin.Context) { PJdatabyID(c) })
		api.POST("/GetAllPJdata", func(c *gin.Context) { GetAllPJdata(c) })
		api.POST("/addproject", func(c *gin.Context) { addproject(c) })
		api.POST("/instructTiming", func(c *gin.Context) { instructTiming(c) })
		api.POST("/modifyprojectStatus", func(c *gin.Context) { modifyprojectStatus(c) })
	}

	r.GET("/ws/onlineServer", func(c *gin.Context) {
		if c.Query("weburl") == "" {
			c.JSON(200, gin.H{
				"msg": "滚！！！",
			})
			return
		}
		webstatistic.WebSocketHandler(c.Query("weburl"), c.Writer, c.Request)
	})

	r.GET("/login", func(context *gin.Context) {
		context.HTML(http.StatusOK, "login.html", nil)
	})
	return r
}

func instructTiming(c *gin.Context) {
	if err := LoginVerify(c); err != nil {
		return
	}
	type Param2 struct {
		Instruct string `form:"instruct" json:"instruct" xml:"instruct" binding:"required"`
	}
	var param Param2
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "param": param})
		return
	}
	addata.InstructTiming(param.Instruct)
	c.JSON(http.StatusOK, gin.H{
		"msg":    "监控状态修改成功",
		"status": 200,
	})
	return
}

func modifyprojectStatus(c *gin.Context) {
	if err := LoginVerify(c); err != nil {
		return
	}
	type Param1 struct {
		Projectid string `form:"projectid" json:"projectid" xml:"projectid" binding:"required"`
		//Status bool `form:"status" json:"status" xml:"status" binding:"required"`
		Statuscode string `form:"statuscode" json:"statuscode" xml:"statuscode" binding:"required"`
	}
	var param Param1
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "param": param})
		return
	}
	Status, _ := strconv.Atoi(param.Statuscode)
	if err := addata.MdfProjectStatus(param.Projectid, Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "msg": "状态修改失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":    "状态修改成功",
		"status": 200,
	})
	return
}

func addproject(c *gin.Context) {
	if err := LoginVerify(c); err != nil {
		return
	}
	var param addata.Param
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := addata.AddProject(param); err != nil {
		c.JSON(400, gin.H{
			"msg":   "项目添加失败",
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":    "项目注册成功",
		"status": 200,
	})
	return
}

func GetAllPJdata(c *gin.Context) {
	if err := LoginVerify(c); err != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"jsondata":     addata.GetAllProjectData(),
		"timingstatus": addata.TimingStatus,
	})
	return
}

func PJdatabyID(c *gin.Context) {
	if err := LoginVerify(c); err != nil {
		return
	}
	type Param struct {
		Projectid string `form:"projectid" json:"projectid" xml:"projectid" binding:"required"`
	}
	var param Param
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"projectid": param.Projectid,
		"jsondata":  addata.APdatabyid(param.Projectid),
	})
	return
}

func todaydata(c *gin.Context) {
	if err := LoginVerify(c); err != nil {
		return
	}
	type Param struct {
		Projectid string `form:"projectid" json:"projectid" xml:"projectid" binding:"required"`
	}
	var param Param
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err1 := addata.Todaydata(param.Projectid)
	if err1 != nil {
		c.JSON(http.StatusOK, gin.H{
			"Projectid": param.Projectid,
			"msg":       err1,
			"jsondata":  res,
			"status":    400,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Projectid": param.Projectid,
		"jsondata":  res,
	})
	return
}

func LoginVerify(c *gin.Context) error {
	// 初始化session对象
	session := sessions.Default(c)
	// 通过session.Get读取session值
	// session是键值对格式数据，因此需要通过key查询数据
	if session.Get("islogin") != "yes" {
		c.JSON(http.StatusOK, gin.H{
			"msg":    "Sorry, you are not logged in",
			"status": 400,
			"data":   "",
		})
		return errors.New("Sorry, you are not logged in")
	}
	return nil
}

func projectdataapi(c *gin.Context) {
	if err := LoginVerify(c); err != nil {
		return
	}
	type Param struct {
		Projectid string `form:"projectid" json:"projectid" xml:"projectid" binding:"required"`
	}
	var param Param
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"projectid": param.Projectid,
		"jsondata":  addata.Projectdata(param.Projectid),
	})
	return
}

func loginapi(c *gin.Context) {
	// 绑定为json
	type Login struct {
		User     string `form:"username" json:"username" xml:"username"  binding:"required"`
		Password string `form:"password" json:"password" xml:"password" binding:"required"`
	}
	var json Login
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if json.User != config.ADname || json.Password != config.ADPwd {
		c.JSON(http.StatusOK, gin.H{
			"msg":    "login fail",
			"status": 400,
		})
		return
	}
	// 初始化session对象
	session := sessions.Default(c)
	// 设置session数据
	session.Set("islogin", "yes")
	// 保存session数据
	_ = session.Save()

	c.JSON(http.StatusOK, gin.H{"msg": "login successfully", "status": 200})
}
