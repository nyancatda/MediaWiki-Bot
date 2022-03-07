/*
 * @Author: NyanCatda
 * @Date: 2021-10-03 04:14:10
 * @LastEditTime: 2022-01-24 19:54:41
 * @LastEditors: NyanCatda
 * @Description: 读取配置文件
 * @FilePath: \ShionBot\src\Utils\ReadConfig\ReadConfig.go
 */
package ReadConfig

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/nyancatda/ShionBot/Controller/MessageProcessing/Struct"
	"github.com/nyancatda/ShionBot/Modular"
	"github.com/nyancatda/ShionBot/Utils/SQLDB"
	"gopkg.in/yaml.v2"
)

var (
	GetConfig  *Config
	ConfigPath string
)

//加载配置文件
func LoadConfig() error {
	//检查配置文件是否存在
	_, err := os.Lstat(ConfigPath)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(ConfigPath)
	if err != nil {
		return err
	}
	newStu := &Config{}
	err = yaml.Unmarshal(content, &newStu)
	if err != nil {
		return err
	}
	GetConfig = newStu

	if err := GetConfig.CheckConfig(); err != nil {
		return err
	}

	return nil
}

/**
 * @description: 检查配置文件字段是否为空
 * @param {*}
 * @return {error}
 */
func (value *Config) CheckConfig() error {
	val := reflect.ValueOf(value).Elem() //获取字段值
	typ := reflect.TypeOf(value).Elem()  //获取字段类型
	//遍历struct中的字段
	for i := 0; i < typ.NumField(); i++ {
		//当字段出现空时，输出错误
		if val.Field(i).IsZero() {
			return errors.New(typ.Field(i).Name + "字段为空，请按照说明填写配置文件")
		}
	}
	return nil
}

/**
 * @description: 读取Wiki名字对应的Wiki链接
 * @param {string} SNSName 聊天软件名字
 * @param {Struct.WebHookJson} Messagejson 消息Json
 * @param {string} WikiName Wiki名字
 * @return {*}
 */
func GetWikiLink(SNSName string, Messagejson Struct.WebHookJson, WikiName string) string {
	//获取用户配置
	db := SQLDB.DB
	var user SQLDB.UserInfo
	UserID := Modular.GetSNSUserID(SNSName, Messagejson)
	db.Where("account = ? and sns_name = ?", UserID, SNSName).Find(&user)
	if user.Account == UserID {
		WikiInfo := user.WikiInfo
		WikiInfoData := []interface{}{}
		json.Unmarshal([]byte(WikiInfo), &WikiInfoData)
		for _, value := range WikiInfoData {
			WikiInfoName := value.(map[string]interface{})["WikiName"].(string)
			if WikiName == WikiInfoName {
				return "https://" + value.(map[string]interface{})["WikiLink"].(string)
			}
		}
	}

	Config := GetConfig
	var ConfigWikiName string
	for one := range Config.Wiki.([]interface{}) {
		ConfigWikiName = Config.Wiki.([]interface{})[one].(map[interface{}]interface{})["WikiName"].(string)
		if ConfigWikiName == WikiName {
			return Config.Wiki.([]interface{})[one].(map[interface{}]interface{})["WikiLink"].(string)
		}
	}
	return ""
}
