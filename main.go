package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"xyz.nyan/MediaWiki-Bot/src/InformationProcessing"
	"xyz.nyan/MediaWiki-Bot/src/MessagePushAPI/SNSAPI/QQAPI"
	"xyz.nyan/MediaWiki-Bot/src/Struct"
	"xyz.nyan/MediaWiki-Bot/src/utils"
	"xyz.nyan/MediaWiki-Bot/src/utils/Language"
	"xyz.nyan/MediaWiki-Bot/src/utils/ReleaseFile"
)

func Error() {
	fmt.Printf(Language.Message("", "").MainErrorTips)
	key := make([]byte, 1)
	os.Stdin.Read(key)
	os.Exit(1)
}

func main() {
	//释放资源文件
	ReleaseFile.ReleaseFile()

	//建立数据储存文件夹
	_, err := os.Stat("./data")
	if err != nil {
		os.MkdirAll("./data", 0777)
		db := utils.SQLLiteLink()
		db.AutoMigrate(&Struct.UserInfo{})
	}

	//读取配置文件
	Config := utils.ReadConfig()

	//判断是否需要初始化QQ部分
	if Config.SNS.QQ.Switch {
		QQAPI.StartQQAPI()
	}

	//启动WebHook接收
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	Port := Config.Run.WebHookPort
	fmt.Println(Language.StringVariable(1, Language.Message("", "").RunOK, Port, ""))
	WebHookKey := Config.Run.WebHookKey
	r.POST("/"+WebHookKey, func(c *gin.Context) {
		var json Struct.WebHookJson
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			fmt.Println(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		InformationProcessing.InformationProcessing(json)
	})
	r.Run(":" + Port)
}
