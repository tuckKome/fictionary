package db

import (
	"fmt"
	"os"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //postgresqlを使うためのライブラリ
	"github.com/tuckKome/fictionary/data"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func argInit() string {
	host := getEnv("FICTIONARY_DATABASE_HOST", "127.0.0.1")
	port := getEnv("FICTIONARY_PORT", "5432")
	user := getEnv("FICTIONARY_USER", "tahoiya")
	dbname := getEnv("FICTIONARY_DB_NAME", "dbtahoiya")
	password := getEnv("FICTIONARY_DB_PASS", "password")
	sslmode := getEnv("FICTIONARY_SSLMODE", "disable")

	dbinfo := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		user,
		password,
		host,
		port,
		dbname,
		sslmode,
	)
	return dbinfo
}

//Init : DB初期化
func Init() {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&data.Kaitou{}, &data.Game{}, &data.Line{}, &data.Vote{}, &data.Donation{})
	defer db.Close()
}

func GetAllLines() []data.Line {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.GetAllLines)")
	}
	defer db.Close()

	var ls []data.Line
	db.Find(&ls)
	return ls
}

//GetGame : DBから一つ取り出す：回答ページで使用
func GetGame(id int) data.Game {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.GetGame)")
	}
	defer db.Close()

	var game data.Game
	db.First(&game, id)
	return game
}

//GetKaitou : 回答 ID で回答を取り出す
func GetKaitou(id int) data.Kaitou {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(sb.GetKaitou)")
	}
	defer db.Close()

	var k data.Kaitou
	db.First(&k, id)
	return k
}

//GetKaitous : DBから[]Kaitouを取り出す
func GetKaitous(g data.Game) []data.Kaitou {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.GetKaitous)")
	}
	defer db.Close()

	var kaitous []data.Kaitou
	// db.Where("game_id = ?", id).Find(&kaitous)
	db.Model(&g).Association("Kaitous").Find(&kaitous)
	return kaitous
}

//GetGames はDBからゲーム一覧を取得
func GetGames() []data.Game {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.GetGames)")
	}
	defer db.Close()

	var games []data.Game
	db.Find(&games)
	return games
}

func GetGamesPhaseIs(st string) []data.Game {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.GetGamesPhaseIs)")
	}
	defer db.Close()

	var gs []data.Game
	db.Where("phase = ?", st).Find(&gs)
	return gs
}

//GetVotes はひとつのKaitou に対する Votes を取得
func GetVotes(b data.Kaitou) []data.Vote {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.GetVotes)")
	}
	defer db.Close()

	var votes []data.Vote
	db.Model(&b).Related(&votes)
	return votes
}

func GetAllDonation() []data.Donation {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.GetDonation)")
	}
	defer db.Close()

	var d []data.Donation
	db.Find(&d)
	return d
}

//InsertKaitou : DBに新しいkaitouを追加
func InsertKaitou(g data.Game, k data.Kaitou) {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.InsertKaitou)")
	}
	defer db.Close()

	db.Model(&g).Association("Kaitous").Append(&k)
	// db.Create(&kaitou)
}

func InsertGame(g data.Game) data.Game {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.InserGame)")
	}
	defer db.Close()

	db.Create(&g)
	var gm data.Game
	db.Last(&gm)
	return gm
}

func InsertDonation(d data.Donation) {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.InsertDonation)")
	}
	defer db.Close()

	db.Create(&d)
}

//UpdateKaitous は解答リストをupdate する
func UpdateKaitous(ks []data.Kaitou) {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.UpdateKaitous)")
	}
	defer db.Close()

	var wg sync.WaitGroup
	for i := range ks {
		wg.Add(1)
		b := ks[i].Base
		go func(num int) {
			defer wg.Done()
			db.Model(&ks[num]).Update("base", b)
		}(i)
	}
	wg.Wait()
}

//UpdateGame は Game を丸ごと更新する
func UpdateGame(g data.Game) data.Game {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.UpdateGame)")
	}
	defer db.Close()

	db.Save(&g)
	db.First(&g, g.ID)
	return g
}

//InsertLine : DBに新しいlineを追加
func InsertLine(line data.Line) {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(InsertLine)")
	}
	defer db.Close()

	db.Create(&line)
}

//DeleteLine はDBから該当するlineを削除。「退出」「アンフォロー」で使用
func DeleteLine(line data.Line) {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.DeleteLine)")
	}
	defer db.Close()

	//TalkIDが一致するものを削除
	db.Where("talk_id = ?", line.TalkID).Delete(&line)
}

//VoteTo は Kaitou に Vote を紐付ける
func VoteTo(k data.Kaitou, v data.Vote) {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic("データベース開ず(db.VoteTo)")
	}
	defer db.Close()

	db.First(&k) //これ必要？？？？？
	db.Model(&k).Association("Votes").Append(&v)
}
