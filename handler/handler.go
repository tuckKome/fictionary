package handler

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/postgres" //postgresql を使うためのライブラリ

	"github.com/line/line-bot-sdk-go/linebot"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"

	"github.com/tuckKome/fictionary/data"
	"github.com/tuckKome/fictionary/db"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func lineInit(id string, typeOfSource string) data.Line {
	var newLine data.Line
	newLine.TalkID = id
	newLine.Type = typeOfSource
	var now = time.Now()
	newLine.CreatedAt = now
	newLine.UpdatedAt = now
	return newLine
}

func isNill(id string) error {
	if id == "" {
		return errors.New("ERROR : Cannnot get game ID")
	}
	return nil
}

//LineConnect : LINE bot 接続
func LineConnect() *linebot.Client {
	channelSecret := getEnv("CHANNEL_SECRET", "")
	channelToken := getEnv("CHANNEL_ACCESS_TOKEN", "")

	bot, err := linebot.New(channelSecret, channelToken)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return bot
}

//TwitterConnect : twitter bot 接続
func TwitterConnect() *twitter.Client {
	consumerKey := getEnv("TWITTER_CONSUMER_KEY", "")
	consumerSecret := getEnv("TWITTE_CONSUMER_SECRET", "")
	accessToken := getEnv("TWITTER_ACCESS_TOKEN", "")
	accessSecret := getEnv("TWITTER_ACCESS_SECRET", "")

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)
	return client
}

func shuffle(a []data.Kaitou) {
	//[0,1,2,...,k-1]を用意
	k := len(a)
	arr := make([]int, k)
	for i := 0; i < k; i++ {
		arr[i] = i
	}

	//Fisher–Yates シャッフル
	rand.Seed(time.Now().UnixNano())
	for i := k - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		tmp := arr[i]
		arr[i] = arr[j]
		arr[j] = tmp
	}

	//シャッフルされたarr を[]Kaitou に入れる
	for i := range a {
		a[i].Base = arr[i]
	}
}

//Index ははじめのページを取得
func Index(c *gin.Context) {
	var h = db.GetGames()
	c.HTML(200, "index.html", gin.H{"History": h})
}

//Error : エラーページ取得
func Error(c *gin.Context) {
	c.HTML(200, "error.html", gin.H{})
}

//GetNewGame 新しいゲームを作るページを取得
func GetNewGame(c *gin.Context) {
	c.HTML(200, "new_game.html", gin.H{})
}

//GetKaitou は回答フォームを取得する
func GetKaitou(c *gin.Context) {
	//idをint型に変換
	n := c.Param("id")
	err := isNill(n)
	if err != nil {
		log.Fatal(err)
		c.Redirect(302, "/error")
	}
	id, err := strconv.Atoi(n)
	if err != nil {
		panic(err)
	}

	game := db.GetGame(id)
	uri := "/games/" + n + "/new"
	c.HTML(200, "phase21.html", gin.H{"odai": game.Odai, "uri": uri})
}

//GetAccepted はAcceptedページを取得
func GetAccepted(c *gin.Context) {
	//idを取得
	n := c.Param("id")
	err := isNill(n)
	if err != nil {
		log.Fatal(err)
		c.Redirect(302, "/error")
	}
	id, err := strconv.Atoi(n)
	if err != nil {
		panic(err)
	}

	game := db.GetGame(id)
	uri := "/games/" + n
	uri2 := "/games/" + n + "/new"
	c.HTML(200, "phase22.html", gin.H{"odai": game.Odai, "uri": uri, "uri2": uri2})
}

func GetVerify(c *gin.Context) {
	//idを取得
	k := c.Param("id")
	n := "secret-" + k
	m := c.PostForm(n)
	err := isNill(k)
	if err != nil {
		log.Fatal(err)
		c.Redirect(302, "/error")
	}
	id, err := strconv.Atoi(k)
	if err != nil {
		panic(err)
	}

	//Game を取得
	l := db.GetGame(id)
	//合言葉は正しい？
	if l.Secret != m {
		c.Redirect(302, "/error")
	}

	uri := "/games/" + k + "/check-in-adv"
	c.Redirect(302, uri)
}

//GetListInAdv は出題者が事前確認するページ

func GetListInAdv(c *gin.Context) {
	//idを取得
	n := c.Param("id")
	err := isNill(n)
	if err != nil {
		log.Fatal(err)
		c.Redirect(302, "/error")
	}
	id, err := strconv.Atoi(n)
	if err != nil {
		panic(err)
	}

	game := db.GetGame(id)
	a := db.GetKaitous(game) //回答一覧を取得

	//Kaitou.Base　で並び替える
	sort.SliceStable(a, func(i, j int) bool { return a[i].Base < a[j].Base })

	k := len(a) //coutOfUsers のため

	uri := "/games/" + n + "/to-playing" //uri のため

	c.HTML(200, "check-in-adv.html", gin.H{
		"odai":         game.Odai,
		"who":          game.CreatedBy,
		"countOfUsers": k,
		"kaitous":      a,
		"uri":          uri,
	})
}

//GetList は回答一覧を取得
func GetList(c *gin.Context) {
	//idを取得
	n := c.Param("id")
	err := isNill(n)
	if err != nil {
		log.Fatal(err)
		c.Redirect(302, "/error")
	}
	id, err := strconv.Atoi(n)
	if err != nil {
		log.Fatal(err)
	}

	game := db.GetGame(id)
	a := db.GetKaitous(game) //回答一覧を取得
	//has many relation を取得
	for i := range a {
		a[i].Votes = db.GetVotes(a[i])
	}

	//Kaitou.Base　で並び替える
	sort.SliceStable(a, func(i, j int) bool { return a[i].Base < a[j].Base })

	k := len(a) //coutOfUsers のため

	uri := "/games/" + n //uri のため

	c.HTML(200, "phase3.html", gin.H{
		"odai":         game.Odai,
		"who":          game.CreatedBy,
		"countOfUsers": k,
		"kaitous":      a,
		"uri":          uri,
	})
}

func getNotNill(a string, b string, c string) string {
	if a != "" {
		return a
	} else if b != "" {
		return b
	} else {
		return c
	}
}

//CreateGame は「*linebot.Client型を引数にした」新しいゲームを作る関数
func CreateGame(bot *linebot.Client, twitterClient *twitter.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := c.PostForm("odai")
		lineUse := c.PostForm("check-line")
		twitterUse := c.PostForm("check-twitter")
		n := c.PostForm("creator-name")
		s := c.PostForm("secret")
		g := data.Game{Odai: t, Phase: "accepting", CreatedBy: n, Secret: s}

		db.InsertGame(g)

		id := strconv.Itoa(int(g.ID))
		uri := "/games/" + id + "/new"

		//LINE bot・twitterに投げる
		if getEnv("GIN_MODE", "debug") == "release" {
			url := getEnv("HOST_ADDRESS", "localhost:8080") + uri
			message := fmt.Sprintf("お題は「%s」\nこのURLから回答してね\n%s", t, url)
			if lineUse == "on" {
				var lines []data.Line
				lines = db.GetAllLines()

				for i := range lines {
					to := lines[i].TalkID
					if _, err := bot.PushMessage(to, linebot.NewTextMessage(message)).Do(); err != nil {
						log.Fatal(err)
					}
				}
			}
			if twitterUse == "on" {
				// Send a Tweet
				_, _, err := twitterClient.Statuses.Update(message, nil)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		c.Redirect(302, uri)
	}
}

//CreateLine は「*linebot.Clientを引数にする」ユーザー・グループIDをDBに登録するhandler
func CreateLine(bot *linebot.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		events, err := bot.ParseRequest(c.Request)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(events) //jsonを確認したい
		for _, event := range events {
			userID := event.Source.UserID
			groupID := event.Source.GroupID
			roomID := event.Source.RoomID

			var line data.Line
			d := getNotNill(userID, groupID, roomID)
			if d == userID {
				line = lineInit(d, "user")
			} else if d == groupID {
				line = lineInit(d, "group")
			} else {
				line = lineInit(d, "room")
			}
			switch event.Type {
			case linebot.EventTypeJoin:
				db.InsertLine(line) //DBにLINEからの情報が登録された
				fmt.Printf("New Line created. ID %s \n", string(line.ID))
			case linebot.EventTypeLeave:
				db.DeleteLine(line)
				fmt.Printf("ID %s is deleted \n", string(line.ID))
			case linebot.EventTypeUnfollow:
				db.DeleteLine(line)
				fmt.Printf("ID %s is deleted \n", string(line.ID))
			}

		}
	}
}

//CreateKaitou は回答を作る
func CreateKaitou(c *gin.Context) {
	//idをuint型に変換
	a := c.Param("id")
	err := isNill(a)
	if err != nil {
		log.Fatal(err)
		c.Redirect(302, "/error")
	}
	id, err := strconv.Atoi(a)
	if err != nil {
		panic(err)
	}
	b := uint(id)
	g := db.GetGame(id)

	h := c.PostForm("name")
	d := c.PostForm("answer")

	e := data.Kaitou{User: h, Answer: d, GameID: b}
	//INSERT
	db.InsertKaitou(g, e)

	f := db.GetKaitous(g)
	shuffle(f)
	db.UpdateKaitous(f)

	uri := "/games/" + a + "/accepted"
	c.Redirect(302, uri)

}

//CreateVote は１つ投票を insert して、 Kaitou に紐づける
func CreateVote(c *gin.Context) {
	//使用してる変数：a, b, d, e, f, g, h, j, k,
	//idを取得
	a := c.Param("id")
	err := isNill(a)
	if err != nil {
		log.Fatal(err)
		c.Redirect(302, "/error")
	}

	uri := "/games/" + a

	l, err := strconv.Atoi(a)
	if err != nil {
		log.Fatal(err)
		c.Redirect(302, "/error")
	}

	b := c.PostForm("slct")
	err = isNill(b)
	if err != nil {
		log.Fatal(err)
		c.Redirect(302, "/error")
	}

	d, err := strconv.Atoi(b)
	if err != nil {
		log.Fatal(err)
		c.Redirect(302, "/error")
	}
	g := db.GetKaitou(d)

	e := c.PostForm("playerName")
	// playerName が重複していないかチェック
	//ゲームの全ての Vote を取得
	h := db.GetGame(l)
	j := db.GetKaitous(h)
	var k []data.Vote
	for i := range j {
		k = append(k, db.GetVotes(j[i])...)
	}
	//重複がないかチェック
	if contains(e, k) {
		c.Redirect(302, uri)
		return
	}

	f := data.Vote{CreatedBy: e, KaitouID: d}
	f.CreatedBy = e

	db.VoteTo(g, f) //Kaitou に Vote を紐つける

	c.Redirect(302, uri)

}

//UpdatePhaseToPlaying は Game の Phase を playing に変更する
func UpdatePhaseToPlaying(c *gin.Context) {
	//id を取得
	k := c.Param("id")
	err := isNill(k)
	if err != nil {
		log.Fatal(err)
		c.Redirect(302, "/error")
	}

	uri := "/games/" + k

	l, err := strconv.Atoi(k)
	if err != nil {
		log.Fatal(err)
		c.Redirect(302, "/error")
	}

	m := db.GetGame(l)
	m.Phase = "playing"
	db.UpdateGame(m)

	c.Redirect(302, uri)
}

func contains(a string, v []data.Vote) bool {
	for i := range v {
		if a == v[i].CreatedBy {
			return true
		}
	}
	return false
}
