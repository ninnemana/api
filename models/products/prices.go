package products

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/curt-labs/GoAPI/helpers/database"
	"github.com/curt-labs/GoAPI/helpers/redis"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"time"
)

type Price struct {
	Id           int       `json:"id,omitempty" xml:"id,omitempty"`
	PartId       int       `json:"partId,omitempty" xml:"partId,omitempty"`
	Type         string    `json:"type,omitempty" xml:"type,omitempty"`
	Price        float64   `json:"price,omitempty" xml:"price,omitempty"`
	Enforced     bool      `json:"enforced,omitempty", xml:"enforced, omitempty"`
	DateModified time.Time `json:"dateModified,omitempty" xml:"dateModified,omitempty"`
}

var (
	partPriceStmt = `
		select priceType, price, enforced from Price
		where partID = ?
		order by priceType`
	getPrice    = `select priceID, partID, priceType, price, enforced, dateModified from Price where priceID = ?`
	createPrice = `INSERT INTO Price (partID, priceType, price, enforced) VALUES (?,?,?,?) `
	updatePrice = `UPDATE Price SET partID = ?, priceType = ?, price = ?, enforced = ? WHERE priceID = ?`
	deletePrice = `DELETE FROM Price WHERE priceID = ?`
)

func (p *Part) GetPricing() error {
	redis_key := fmt.Sprintf("part:%d:pricing", p.ID)

	data, err := redis.Get(redis_key)
	if err == nil && len(data) > 0 {
		if err = json.Unmarshal(data, &p.Pricing); err != nil {
			return nil
		}
	}

	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	qry, err := db.Prepare(partPriceStmt)
	if err != nil {
		return err
	}
	defer qry.Close()

	rows, err := qry.Query(p.ID)
	if err != nil || rows == nil {
		return err
	}

	for rows.Next() {
		var pr Price
		err = rows.Scan(&pr.Type, &pr.Price, &pr.Enforced)
		if err == nil {
			p.Pricing = append(p.Pricing, pr)
		}
	}

	go redis.Setex(redis_key, p.Pricing, 86400)

	return nil
}

//by priceId
func (p *Price) Get() error {
	redis_key := fmt.Sprintf("pricing:" + strconv.Itoa(p.Id))

	data, err := redis.Get(redis_key)
	if err == nil && len(data) > 0 {
		if err = json.Unmarshal(data, &p); err != nil {
			return nil
		}
	}

	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	qry, err := db.Prepare(getPrice)
	if err != nil {
		return err
	}
	defer qry.Close()

	err = qry.QueryRow(p.Id).Scan(&p.Id, &p.PartId, &p.Type, &p.Price, &p.Enforced, &p.DateModified)
	if err != nil {
		return err
	}

	go redis.Setex(redis_key, p, 86400)

	return nil
}
func (p *Price) Create() (err error) {
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin()
	stmt, err := tx.Prepare(createPrice)
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(p.PartId, p.Type, p.Price, p.Enforced)
	if err != nil {
		tx.Rollback()
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	p.Id = int(id)
	err = tx.Commit()
	return err
}

func (p *Price) Update() (err error) {
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin()
	stmt, err := tx.Prepare(updatePrice)
	if err != nil {
		return err
	}
	log.Print(p.Price)
	defer stmt.Close()
	_, err = stmt.Exec(p.PartId, p.Type, p.Price, p.Enforced, p.Id)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

func (p *Price) Delete() (err error) {
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin()
	stmt, err := tx.Prepare(deletePrice)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(p.Id)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}
