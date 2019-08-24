package barrage

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"strings"
	"unicode/utf8"
)

func Filter(barrage Barrage) Barrage {
	db, err := sql.Open("mysql", "root:123@tcp(127.0.0.1:3306)/barrage")
	CheckError(err, "Can't connect to mysql")
	defer db.Close()
	rows, err := db.Query("SELECT word FROM sensitive_word")
	columns, _ := rows.Columns()
	columnLength := len(columns)
	cache := make([]interface{}, columnLength)
	for index, _ := range cache {
		var a interface{}
		cache[index] = &a
	}
	var list []map[string]interface{}
	for rows.Next() {
		_ = rows.Scan(cache...)
		item := make(map[string]interface{})
		for i, data := range cache {
			item[columns[i]] = *data.(*interface{})
		}
		list = append(list, item)
	}
	rows.Close()
	res := barrage.Message
	for i := 0; i < len(list); i++ {
		word := string(list[i]["word"].([]uint8))
		length := utf8.RuneCountInString(word)
		new := ""
		for i := 0; i < length; i++ {
			new += "*"
		}
		res = strings.Replace(res, word, new, -1)
	}
	barrage.Message = res
	return barrage
}

func AddWord(c *gin.Context) {
	new := c.Request.FormValue("word")
	if new != "" {
		db, err := sql.Open("mysql", "root:123@tcp(127.0.0.1:3306)/barrage")
		CheckError(err, "Can't connect to mysql")
		defer db.Close()
		rows, err := db.Query("SELECT word FROM sensitive_word")
		columns, _ := rows.Columns()
		columnLength := len(columns)
		cache := make([]interface{}, columnLength)
		for index, _ := range cache {
			var a interface{}
			cache[index] = &a
		}
		var list []map[string]interface{}
		for rows.Next() {
			_ = rows.Scan(cache...)
			item := make(map[string]interface{})
			for i, data := range cache {
				item[columns[i]] = *data.(*interface{})
			}
			list = append(list, item)
		}
		rows.Close()
		for i := 0; i < len(list); i++ {
			word := string(list[i]["word"].([]uint8))
			if new == word {
				c.JSON(200, map[string]interface{}{"message": "敏感词已存在", "status": 600})
				goto End
			}
		}
		_, err = db.Exec("INSERT INTO sensitive_word(word)VALUES (?)", new)
		CheckError(err, "Can't add word")
		c.JSON(200, map[string]interface{}{"message": "添加敏感词成功", "status": 200})
	} else {
		c.JSON(200, map[string]interface{}{"message": "敏感词不能为空", "status": 400})
	}
End:
}
