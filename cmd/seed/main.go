package main

import (
	"fmt"
	"os"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

type Vtuber struct {
	bun.BaseModel `bun:"table:vtubers"`

	ID        string    `bun:"id,type:varchar(24),pk"`
	Name      string    `bun:"name,notnull,type:varchar"`
	ItemCount int64     `bun:"item_count,default:0,type:integer"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp()"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp() ON UPDATE current_timestamp()"`
}

type Video struct {
	bun.BaseModel `bun:"table:videos"`

	ID        string    `bun:"id,type:varchar(11),pk"`
	Title     string    `bun:"title,notnull,type:varchar"`
	Duration  string    `bun:"duration,notnull,type:varchar"` //int型に変換したほうがいいか？
	Viewers   int64     `bun:"viewers,notnull,type:integer"`
	Content   string    `bun:"content,notnull,type:varchar"`
	Announced bool      `bun:"announced,default:false,type:boolean"`
	StartTime time.Time `bun:"scheduled_start_time,type:timestamp"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp()"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp() ON UPDATE current_timestamp()"`
}

func main() {
	sqldb, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		panic(err)
	}

	// ctx := context.Background()
	db := bun.NewDB(sqldb, mysqldialect.New())

	// var ids []string
	// var itemCount []int64
	// err = db.NewSelect().Model((*Vtubers)(nil)).Column("id", "item_count").Scan(ctx, &ids, &itemCount)
	// if err != nil {
	// 	panic(err)
	// }
	// for i := range ids {
	// 	ids[i] = strings.Replace(ids[i], "UC", "UU", 1)
	// }
	// fmt.Println(ids)
	// fmt.Println(itemCount)

	query := db.NewCreateTable().Model((*Vtuber)(nil))
	rawQuery, err := query.AppendQuery(db.Formatter(), nil)
	if err != nil {
		panic(err)
	}
	s := string(rawQuery)
	fmt.Println(s)

	query = db.NewCreateTable().Model((*Video)(nil))
	rawQuery, err = query.AppendQuery(db.Formatter(), nil)
	if err != nil {
		panic(err)
	}
	s = string(rawQuery)
	fmt.Println(s)

	// ctx := context.Background()

	// ids := []string{"UC_4tXjqecqox5Uc05ncxpxg", "UC_82HBGtvwN1hcGeOGHzUBQ"}
	// cntList := []int64{1584, 590}
	// query := db.NewUpdate().Model((*Vtuber)(nil)).
	// 	Set("item_count = ELT(FIELD(id, ?), ?)", bun.In(ids), bun.In(cntList)).
	// 	Where("id IN (?)", bun.In(ids))
	// rawQuery, err := query.AppendQuery(db.Formatter(), nil)
	// if err != nil {
	// 	panic(err)
	// }
	// s := string(rawQuery)
	// fmt.Println(s)

	// _, err = query.Exec(ctx)
	// if err != nil {
	// 	slog.Error("update-itemCount",
	// 		slog.String("severity", "ERROR"),
	// 		slog.String("message", err.Error()),
	// 	)
	// 	return
	// }

	// var vList []Vtubers
	// f, err := os.Open("cmd/seed/list.csv")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer f.Close()

	// r := csv.NewReader(f)
	// for {
	// 	record, err := r.Read()
	// 	if err == io.EOF {
	// 		break
	// 	}
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Println(record)
	// 	vList = append(vList, Vtubers{
	// 		ID: record[0],
	// 		Name: record[1],
	// 	})
	// }

	// _, err = db.NewInsert().Model(&vList).Exec(ctx)
	// if err != nil {
	// 	panic(err)
	// }
}
