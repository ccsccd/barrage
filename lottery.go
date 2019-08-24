package barrage

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func LotteryEnter(c *gin.Context) {
	c.JSON(200, map[string]interface{}{"luckyUser": lottery(), "status": 200})
}

func RedPaccketEnter(c *gin.Context) {
	word := c.Request.FormValue("word")
	c.JSON(200, map[string]interface{}{"luckyUser": redPacket(word), "status": 200})
}
func lottery() string {
	db, err := sql.Open("mysql", "root:123@tcp(127.0.0.1:3306)/barrage")
	CheckError(err, "Can't connect to mysql")
	defer db.Close()
	rows, err := db.Query("SELECT * FROM barrage AS t1 JOIN (SELECT ROUND(RAND() * ((SELECT MAX(id) FROM barrage)-(SELECT MIN(id) FROM barrage))+(SELECT MIN(id) FROM barrage)) AS id) AS t2 WHERE t1.id >= t2.id ORDER BY t1.id LIMIT 1")
	var id,id2 int
	var message, color, userId string
	for rows.Next() {
		err := rows.Scan(&id, &message, &color, &userId,&id2)
		CheckError(err, "Can't find")
		fmt.Println("LuckyID:",id)
	}
	rows.Close()
	return userId
}

func redPacket(word string) string {
	db, err := sql.Open("mysql", "root:123@tcp(127.0.0.1:3306)/barrage")
	CheckError(err, "Can't connect to mysql")
	defer db.Close()
	rows, err := db.Query("SELECT * FROM (SELECT * FROM barrage WHERE message LIKE ?)t1 WHERE t1.id >=(SELECT FLOOR(((MAX(id)-MIN(id)) * RAND()) + MIN(id)) FROM (SELECT * FROM barrage WHERE message LIKE ?)t2) ORDER BY t1.id LIMIT 1",word,word)
	var id int
	var message, color, userId string
	for rows.Next() {
		err := rows.Scan(&id, &message, &color, &userId)
		CheckError(err, "Can't find")
		fmt.Println("LuckyID:",id)
	}
	rows.Close()
	return userId
}
