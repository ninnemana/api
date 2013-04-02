package models

import (
	"../helpers/database"
	"../helpers/redis"
	"../helpers/rest"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	basicsStmt = `select p.status, p.dateAdded, p.dateModified, p.shortDesc, p.partID, p.priceCode, pc.class
				from Part as p
				left join Class as pc on p.classID = pc.classID
				where p.partID = ? && p.status in (800,900) limit 1`

	partAttrStmt = `select field, value from PartAttribute where partID = ?`

	partPriceStmt = `select priceType, price, enforced from Price where partID = ?`

	relatedPartStmt = `select distinct relatedID from RelatedPart
				where partID = ?
				order by relatedID`

	partContentStmt = `select ct.type, con.text
				from Content as con
				join ContentBridge as cb on con.contentID = cb.contentID
				join ContentType as ct on con.cTypeID = ct.cTypeID
				where cb.partID = ?
				order by ct.type`

	partInstallSheetStmt = `select c.text from ContentBridge as cb
					join Content as c on cb.contentID = c.contentID
					join ContentType as ct on c.cTypeID = ct.cTypeID
					where partID = ? && ct.type = 'InstallationSheet'
					limit 1`
)

type Part struct {
	PartId, Status, PriceCode, RelatedCount int
	AverageReview                           float64
	DateModified, DateAdded                 time.Time
	ShortDesc, PartClass                    string
	InstallSheet                            *url.URL
	Attributes                              []Attribute
	VehicleAttributes                       []string
	Content                                 []Content
	Pricing                                 []Pricing
	Reviews                                 []Review
	Images                                  []Image
	Related                                 []int
	Categories                              []ExtendedCategory
	Videos                                  []Video
	Packages                                []Package
	Customer                                CustomerPart
}

type PagedParts struct {
	Parts  []Part
	Paging []Paging
}

type Paging struct {
	CurrentIndex int
	PageCount    int
}

type CustomerPart struct {
	Price         float64
	CartReference int
}

type Attribute struct {
	Key, Value string
}

type Content struct {
	Key, Value string
}

type Pricing struct {
	Type     string
	Price    float64
	Enforced bool
}

func (p *Part) FromDatabase() error {

	var errs []string

	basicChan := make(chan int)
	attrChan := make(chan int)
	priceChan := make(chan int)
	reviewChan := make(chan int)
	imageChan := make(chan int)
	videoChan := make(chan int)
	relatedChan := make(chan int)
	packageChan := make(chan int)
	categoryChan := make(chan int)
	contentChan := make(chan int)

	go func() {
		basicErr := p.Basics()
		if basicErr != nil {
			errs = append(errs, basicErr.Error())
		}
		basicChan <- 1
	}()

	go func() {
		attrErr := p.GetAttributes()
		if attrErr != nil {
			errs = append(errs, attrErr.Error())
		}
		attrChan <- 1
	}()

	go func() {
		priceErr := p.GetPricing()
		if priceErr != nil {
			errs = append(errs, priceErr.Error())
		}
		priceChan <- 1
	}()

	go func() {
		reviewErr := p.GetReviews()
		if reviewErr != nil {
			errs = append(errs, reviewErr.Error())
		}
		reviewChan <- 1
	}()

	go func() {
		imgErr := p.GetImages()
		if imgErr != nil {
			errs = append(errs, imgErr.Error())
		}
		imageChan <- 1
	}()

	go func() {
		vidErr := p.GetVideos()
		if vidErr != nil {
			errs = append(errs, vidErr.Error())
		}
		videoChan <- 1
	}()

	go func() {
		relErr := p.GetRelated()
		if relErr != nil {
			errs = append(errs, relErr.Error())
		}
		relatedChan <- 1
	}()

	go func() {
		pkgErr := p.GetPartPackaging()
		if pkgErr != nil {
			errs = append(errs, pkgErr.Error())
		}
		packageChan <- 1
	}()

	go func() {
		catErr := p.PartBreadcrumbs()
		if catErr != nil {
			errs = append(errs, catErr.Error())
		}
		categoryChan <- 1
	}()

	go func() {
		conErr := p.GetContent()
		if conErr != nil {
			errs = append(errs, conErr.Error())
		}
		contentChan <- 1
	}()

	<-basicChan
	<-attrChan
	<-priceChan
	<-reviewChan
	<-imageChan
	<-videoChan
	<-relatedChan
	<-packageChan
	<-categoryChan
	<-contentChan

	if len(errs) > 0 {
		return errors.New("Error: " + strings.Join(errs, ", "))
	}

	if part_bytes, err := json.Marshal(p); err == nil {
		part_key := "part:" + strconv.Itoa(p.PartId)
		redis.RedisClient.Set(part_key, part_bytes)
		redis.RedisClient.Expire(part_key, 86400)
	}

	return nil
}

func (p *Part) FromCache() error {

	part_bytes, err := redis.RedisClient.Get("part:" + strconv.Itoa(p.PartId))
	if err != nil {
		return err
	} else if len(part_bytes) == 0 {
		return errors.New("Part does not exist in cache")
	}

	err = json.Unmarshal(part_bytes, &p)

	return err
}

func (p *Part) Get(key string) error {

	partChan := make(chan int)
	customerChan := make(chan int)

	var err error

	go func() {
		if err = p.FromCache(); err != nil {
			err = p.FromDatabase()
		}
		partChan <- 1
	}()

	go func(api_key string) {
		err = p.BindCustomer(api_key)
		customerChan <- 1
	}(key)

	<-partChan
	<-customerChan

	return err
}

func (p *Part) GetWithVehicle(vehicle *Vehicle, api_key string) error {

	var errs []string

	superChan := make(chan int)
	noteChan := make(chan int)
	go func(key string) {
		p.Get(key)
		superChan <- 1
	}(api_key)
	go func() {
		notes, nErr := vehicle.GetNotes(p.PartId)
		if nErr != nil && notes != nil {
			errs = append(errs, nErr.Error())
			p.VehicleAttributes = []string{}
		} else {
			p.VehicleAttributes = notes
		}
		noteChan <- 1
	}()

	<-superChan
	<-noteChan

	if len(errs) > 0 {
		return errors.New("Error: " + strings.Join(errs, ", "))
	}
	return nil
}

func (p *Part) GetById(id int, key string) {
	p.PartId = id

	p.Get(key)
}

func (p *Part) GetAttributes() (err error) {
	qry, err := database.Db.Prepare(partAttrStmt)
	if err != nil {
		return
	}

	rows, _, err := qry.Exec(p.PartId)
	if database.MysqlError(err) {
		return err
	} else if rows == nil {
		return
	}

	var attrs []Attribute
	for _, row := range rows {
		attr := Attribute{
			Key:   row.Str(0),
			Value: row.Str(1),
		}
		attrs = append(attrs, attr)
	}
	p.Attributes = attrs

	return
}

func (p *Part) Basics() error {
	qry, err := database.Db.Prepare(basicsStmt)
	if err != nil {
		return err
	}

	row, res, err := qry.ExecFirst(p.PartId)
	if database.MysqlError(err) {
		return err
	} else if row == nil {
		return errors.New("No Part Found for:" + string(p.PartId))
	}
	status := res.Map("status")
	dateAdded := res.Map("dateAdded")
	dateModified := res.Map("dateModified")
	shortDesc := res.Map("shortDesc")
	partID := res.Map("partID")
	priceCode := res.Map("priceCode")
	class := res.Map("class")

	p.PartId = row.Int(partID)
	date_add, _ := time.Parse("2006-01-02 15:04:15", row.Str(dateAdded))
	p.DateAdded = date_add

	date_mod, _ := time.Parse("2006-01-02 15:04:15", row.Str(dateModified))
	p.DateModified = date_mod

	p.ShortDesc = row.Str(shortDesc)
	p.PriceCode = row.Int(priceCode)
	p.PartClass = row.Str(class)
	p.Status = row.Int(status)

	return nil
}

func (p *Part) GetPricing() error {
	qry, err := database.Db.Prepare(partPriceStmt)
	if err != nil {
		return err
	}

	rows, res, err := qry.Exec(p.PartId)
	if database.MysqlError(err) {
		return err
	} else if rows == nil {
		return errors.New("No pricing found for part: " + string(p.PartId))
	}

	typ := res.Map("priceType")
	price := res.Map("price")
	enforced := res.Map("enforced")

	var prices []Pricing
	for _, row := range rows {
		pr := Pricing{
			row.Str(typ),
			row.Float(price),
			row.ForceBool(enforced),
		}

		if pr.Type == "Map" {
			pr.Enforced = true
		}
		prices = append(prices, pr)
	}

	p.Pricing = prices

	return nil
}

func (p *Part) GetRelated() error {
	qry, err := database.Db.Prepare(relatedPartStmt)
	if err != nil {
		return err
	}

	rows, _, err := qry.Exec(p.PartId)
	if database.MysqlError(err) {
		return err
	} else if rows == nil {
		return errors.New("No related found for part: " + string(p.PartId))
	}

	var related []int
	for _, row := range rows {
		related = append(related, row.Int(0))
	}
	p.Related = related
	p.RelatedCount = len(related)
	return nil
}

func (p *Part) GetContent() error {
	qry, err := database.Db.Prepare(partContentStmt)
	if err != nil {
		return err
	}

	rows, _, err := qry.Exec(p.PartId)
	if database.MysqlError(err) {
		return err
	} else if rows == nil {
		return errors.New("No content found for part: " + string(p.PartId))
	}

	var content []Content
	for _, row := range rows {
		con := Content{
			Key:   row.Str(0),
			Value: row.Str(1),
		}

		if strings.Contains(strings.ToLower(con.Key), "install") {
			sheetUrl, _ := url.Parse(con.Value)
			p.InstallSheet = sheetUrl
		} else {
			content = append(content, con)
		}
	}
	p.Content = content
	return nil
}

func (p *Part) BindCustomer(key string) error {
	price, err := GetCustomerPrice(key, p.PartId)
	if err != nil {
		return err
	}

	ref, err := GetCustomerCartReference(key, p.PartId)
	if err != nil {
		return err
	}

	cust := CustomerPart{
		Price:         price,
		CartReference: ref,
	}
	p.Customer = cust
	return nil
}

func (p *Part) GetInstallSheet(r *http.Request) (data []byte, err error) {
	qry, err := database.Db.Prepare(partInstallSheetStmt)
	if err != nil {
		return
	}

	row, _, err := qry.ExecFirst(p.PartId)
	if database.MysqlError(err) || row == nil {
		return
	}

	data, err = rest.GetPDF(row.Str(0), r)

	return
}
