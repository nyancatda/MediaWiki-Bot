package GetWikiInfo

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/nyancatda/ShionBot/src/MediaWikiAPI"
	"github.com/nyancatda/ShionBot/src/Modular"
	"github.com/nyancatda/ShionBot/src/Struct"
	"github.com/nyancatda/ShionBot/src/utils"
	"github.com/nyancatda/ShionBot/src/utils/Language"
)

func Error(SNSName string, UserID string, WikiLink string, title string, LanguageMessage *Language.LanguageInfo) string {
	text := utils.StringVariable(LanguageMessage.GetWikiInfoError, []string{WikiLink, title})
	return text
}

//判断Wiki名字是否存在
func WikiNameExist(WikiName string, SNSName string, Messagejson Struct.WebHookJson) bool {
	//判断用户设置
	db := utils.SQLLiteLink()
	var user Struct.UserInfo
	UserID := Modular.GetSNSUserID(SNSName, Messagejson)
	db.Where("account = ? and sns_name = ?", UserID, SNSName).Find(&user)
	if user.Account == UserID {
		WikiInfo := user.WikiInfo
		WikiInfoData := []interface{}{}
		json.Unmarshal([]byte(WikiInfo), &WikiInfoData)
		for _, value := range WikiInfoData {
			WikiInfoName := value.(map[string]interface{})["WikiName"].(string)
			if find := strings.Contains(WikiName, WikiInfoName); find {
				return true
			}
		}
	}

	Config := utils.GetConfig
	var ConfigWikiName string
	for one := range Config.Wiki.([]interface{}) {
		ConfigWikiName = Config.Wiki.([]interface{})[one].(map[interface{}]interface{})["WikiName"].(string)
		if find := strings.Contains(WikiName, ConfigWikiName); find {
			return true
		}
	}
	return false
}

//获取主Wiki名字
func GeiMainWikiName(SNSName string, Messagejson Struct.WebHookJson) string {
	//获取用户设置
	db := utils.SQLLiteLink()
	var user Struct.UserInfo
	UserID := Modular.GetSNSUserID(SNSName, Messagejson)
	db.Where("account = ? and sns_name = ?", UserID, SNSName).Find(&user)
	if user.Account == UserID {
		WikiInfo := user.WikiInfo
		WikiInfoData := []interface{}{}
		json.Unmarshal([]byte(WikiInfo), &WikiInfoData)
		for _, value := range WikiInfoData {
			WikiInfoName := value.(map[string]interface{})["WikiName"].(string)
			return WikiInfoName
		}
	}

	Config := utils.GetConfig
	MainWikiName := Config.Wiki.([]interface{})[0].(map[interface{}]interface{})["WikiName"].(string)
	return MainWikiName
}

//搜索wiki
func SearchWiki(SNSName string, Messagejson Struct.WebHookJson, WikiName string, title string) string {
	SearchInfo, _ := MediaWikiAPI.Opensearch(WikiName, 10, title)
	if len(SearchInfo) != 0 {
		SearchList := SearchInfo[1].([]interface{})
		if len(SearchList) != 0 {
			var SearchPages strings.Builder
			for _, value := range SearchList {
				PagseName := "[" + value.(string) + "]"
				SearchPages.WriteString(PagseName)
				SearchPages.WriteString("\n")
			}
			return SearchPages.String()
		}
		return ""
	} else {
		return ""
	}
}

//为空处理
func NilProcessing(SNSName string, Messagejson Struct.WebHookJson, UserID string, WikiName string, title string, LanguageMessage *Language.LanguageInfo) string {
	SearchInfo := SearchWiki(SNSName, Messagejson, WikiName, title)
	if SearchInfo != "" {
		Info := utils.StringVariable(LanguageMessage.WikiInfoSearch, []string{SearchInfo, WikiName})
		return Info
	} else {
		WikiLink := utils.GetWikiLink(SNSName, Messagejson, WikiName)
		return Error(SNSName, UserID, WikiLink, title, LanguageMessage)
	}
}

//获取Wiki页面标题，过滤后缀
func GetUrlTitle(SNSName string, Messagejson Struct.WebHookJson, WikiName string, PageName string) string {
	WikiLink := utils.GetWikiLink(SNSName, Messagejson, WikiName)
	doc, err := htmlquery.LoadURL(WikiLink + "/" + PageName)
	if err != nil {
		fmt.Println(err)
	}
	for _, n := range htmlquery.Find(doc, "/html/head/title") {
		PageTitle := htmlquery.OutputHTML(n, false)
		countSplit := strings.SplitN(PageTitle, " - ", 2)
		Title := countSplit[0]
		return Title
	}
	return ""
}

//查询页面是否存在重定向
func QueryRedirects(SNSName string, Messagejson Struct.WebHookJson, WikiName string, title string) (whether bool, to string, from string, err error) {
	WikiLink := utils.GetWikiLink(SNSName, Messagejson, WikiName)
	info, err := MediaWikiAPI.QueryRedirects(WikiLink, title)

	for _, value := range info.Query.Pages {
		if value.Title != "" {
			if len(info.Query.Normalized) != 0 {
				return true, info.Query.Normalized[0].To, info.Query.Normalized[0].From, err
			} else {
				PageTitleInfo := GetUrlTitle(SNSName, Messagejson, WikiName, title)
				if PageTitleInfo != title {
					ToTitle := PageTitleInfo
					return true, ToTitle, title, err
				}
			}
			return false, "", "", err
		}
	}

	return false, "", "", err
}

//获取Wiki页面信息
func GetWikiInfo(SNSName string, Messagejson Struct.WebHookJson, UserID string, WikiName string, title string, language string) (string, error) {
	var LanguageMessage *Language.LanguageInfo
	if language != "" {
		LanguageMessage = Language.DesignateLanguageMessage(language)
	} else {
		LanguageMessage = Language.Message(SNSName, UserID)
	}
	var err error
	RedirectsState, ToTitle, FromTitle, _ := QueryRedirects(SNSName, Messagejson, WikiName, title)
	var info MediaWikiAPI.QueryExtractsJson
	WikiLink := utils.GetWikiLink(SNSName, Messagejson, WikiName)
	if RedirectsState {
		info, err = MediaWikiAPI.QueryExtracts(WikiLink, 100, ToTitle)
	} else {
		info, err = MediaWikiAPI.QueryExtracts(WikiLink, 100, title)
	}

	if len(info.Query.Pages) == 0 {
		return NilProcessing(SNSName, Messagejson, UserID, WikiName, title, LanguageMessage), err
	}

	var PageId string
	for one := range info.Query.Pages {
		PageId = one
	}

	if PageId != "-1" {
		PagesExtract := info.Query.Pages[PageId].Extract
		var returnText string
		if RedirectsState {
			WikiPageInfo, err := QueryWikiInfo(SNSName, Messagejson, WikiName, ToTitle)
			if err != nil {
				log.Println(err)
			}
			WikiPageLink := WikiPageInfo.(map[string]interface{})["fullurl"].(string)
			info := utils.StringVariable(LanguageMessage.WikiInfoRedirect, []string{FromTitle, ToTitle})
			returnText = WikiPageLink + info + PagesExtract.(string)
		} else {
			WikiPageInfo, err := QueryWikiInfo(SNSName, Messagejson, WikiName, title)
			if err != nil {
				log.Println(err)
			}
			WikiPageLink := WikiPageInfo.(map[string]interface{})["fullurl"].(string)
			returnText = WikiPageLink + "\n[" + title + "]\n" + PagesExtract.(string)
		}
		return returnText, err
	} else {
		return NilProcessing(SNSName, Messagejson, UserID, WikiName, title, LanguageMessage), err
	}
}
