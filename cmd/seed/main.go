package main

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	nsa "github.com/aopontann/nijisanji-songs-announcement"
)

func main() {
	dsn := "postgres://postgres:example@localhost:5432/test_db?sslmode=disable"
	// dsn := "unix://user:pass@dbname/var/run/postgresql/.s.PGSQL.5432"
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}
	sqldb := stdlib.OpenDB(*config)
	db := bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()
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

	query := db.NewCreateTable().Model((*nsa.Vtuber)(nil))
	rawQuery, err := query.AppendQuery(db.Formatter(), nil)
	if err != nil {
		panic(err)
	}
	s := string(rawQuery)
	fmt.Println(s)

	query = db.NewCreateTable().Model((*nsa.Video)(nil))
	rawQuery, err = query.AppendQuery(db.Formatter(), nil)
	if err != nil {
		panic(err)
	}
	s = string(rawQuery)
	fmt.Println(s)

	query = db.NewCreateTable().Model((*nsa.User)(nil))
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
