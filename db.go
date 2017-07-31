package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

func initDB() (DataBase DB, err error) {
	DataBase.db, err = sql.Open("sqlite3", "./twicli.db")

	if err != nil {
		return DataBase, err
	}

	sqlStmt := `
		create table accesstoken (
        access_token text primary key
		);
		`
	_, _ = DataBase.db.Exec(sqlStmt)

	return DataBase, nil
}

func (DataBase *DB) InsertAccessToken(accessToken string) (err error) {
	selectRow, err := DataBase.SelectAccessToken()
	if err != nil {
		return err
	}
	if selectRow.accessToken != "" {
		err = DataBase.DeleteAccessToken()
		if err != nil {
			return err
		}
	}

	tx, err := DataBase.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("insert into accesstoken(access_token) values (?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(accessToken)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

type rowAccessToken struct {
	accessToken string
}

func (DataBase *DB) SelectAccessToken() (selectRow rowAccessToken, err error) {
	row, err := DataBase.db.Query("select access_token from accesstoken")
	if err != nil {
		return selectRow, err
	}
	defer row.Close()

	if !row.Next() {
		return selectRow, nil
	}
	var accessToken string
	err = row.Scan(&accessToken)
	if err != nil {
		return selectRow, err
	}
	selectRow.accessToken = accessToken

	return selectRow, nil
}

func (DataBase *DB) DeleteAccessToken() (err error) {
	_, err = DataBase.db.Exec("delete from accesstoken")
	if err != nil {
		return err
	}
	return nil
}
