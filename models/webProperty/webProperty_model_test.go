package webProperty_model

import (
	"github.com/curt-labs/GoAPI/helpers/api"
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestGetWebProperties(t *testing.T) {
	Convey("Testing Gets", t, func() {
		Convey("Testing GetAll()", func() {
			var ws WebProperties
			ws, err := GetAll()
			So(ws, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(len(ws), ShouldNotBeNil)
			if len(ws) > 0 {
				Convey("Testing Get()", func() {
					var w WebProperty
					w = ws[api_helpers.RandGenerator(len(ws)-1)]
					err = w.Get()
					So(w, ShouldNotBeNil)
					So(err, ShouldBeNil)
					So(w.Name, ShouldNotEqual, "")
					So(w.CustID, ShouldNotEqual, 0)
					So(w.WebPropertyType.Type, ShouldNotEqual, "")
					t.Logf("Web Property Notes: %v", w.WebPropertyNotes)
					var ns WebPropertyNotes
					So(w.WebPropertyNotes, ShouldHaveSameTypeAs, ns)
					So(len(w.WebPropertyRequirements), ShouldNotEqual, 0)
				})
				Convey("Testing Get(); focus on dates", func() {
					var w WebProperty
					w = ws[api_helpers.RandGenerator(len(ws)-1)]
					err = w.Get()
					So(w, ShouldNotBeNil)
					var t time.Time
					So(w.IsEnabledDate, ShouldHaveSameTypeAs, t)
					So(w.RequestedDate, ShouldHaveSameTypeAs, t)

					So(w.AddedDate, ShouldHaveSameTypeAs, t)
					So(err, ShouldBeNil)
				})
			}

		})

		Convey("Testing GetAllWebPropertyTypes()", func() {
			qs, err := GetAllWebPropertyTypes()
			So(qs, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("Testing GetNote()", func() {
			var n WebPropertyNote
			n.ID = 1
			err := n.Get()
			So(n.Text, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("Testing GetWebPropertyType()", func() {
			var n WebPropertyType
			n.ID = 1
			err := n.Get()
			So(n.Type, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("Testing GetWebPropertyRequirement()", func() {
			var n WebPropertyRequirement
			n.RequirementID = 1
			err := n.Get()
			So(n.Requirement, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("Testing GetWebPropertyRequirementCheck()", func() {
			var n WebPropertyRequirement
			n.ID = 16
			err := n.GetJoin()
			So(n.WebPropID, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("Testing GetAllWebPropertyNotes()", func() {
			qs, err := GetAllWebPropertyNotes()
			So(qs, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("Testing GetAllWebPropertyRequirements()", func() {
			qs, err := GetAllWebPropertyRequirements()
			So(qs, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("Testing Search()", func() {
			as, err := Search("test", "", "", "", "", "", "", "", "", "", "", "", "1", "0")
			So(as, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(as.Pagination.Page, ShouldEqual, 1)
			So(as.Pagination.ReturnedCount, ShouldNotBeNil)
			So(as.Pagination.PerPage, ShouldNotBeNil)
			So(as.Pagination.PerPage, ShouldEqual, len(as.Objects))
		})

	})
	Convey("Testing CUD", t, func() {
		Convey("Testing Create(), Update(), Delete()", func() {
			var f WebProperty
			var err error
			f.Name = "testTitle"
			f.CustID = 12345
			f.BadgeID = strconv.Itoa(rand.Int())
			f.IsEnabledDate, err = time.Parse(timeFormat, "2004-03-03 9:15:00")
			f.Url = "www.test.com"
			f.IsEnabled = true
			f.SellerID = "test"
			f.WebPropertyType.ID = 2
			f.IsFinalApproved = true
			f.IsDenied = false
			f.RequestedDate, err = time.Parse(timeFormat, "2004-03-03 9:15:00")

			f.Create()
			So(f.ID, ShouldNotBeNil)
			f.Get()
			So(f, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(f.Name, ShouldEqual, "testTitle")
			So(f.CustID, ShouldEqual, 12345)
			So(f.BadgeID, ShouldNotBeNil)
			So(f.Url, ShouldEqual, "www.test.com")
			So(f.SellerID, ShouldEqual, "test")
			So(f.WebPropertyType.ID, ShouldEqual, 2)
			So(f.IsFinalApproved, ShouldBeTrue)
			So(f.IsDenied, ShouldBeFalse)
			t, err := time.Parse(timeFormat, "2004-03-03 09:15:00")
			So(f.IsEnabledDate, ShouldResemble, t)
			u, err := time.Parse(timeFormat, "2004-03-03 09:15:00")
			So(f.RequestedDate, ShouldResemble, u)
			f.Name = "testTitle2"
			f.CustID = 123452
			f.BadgeID = strconv.Itoa(rand.Int())
			f.IsEnabledDate, err = time.Parse(timeFormat, "2004-03-03 9:15:22")
			f.Url = "www.test.com2"
			f.IsEnabled = false
			f.SellerID = "test2"
			f.WebPropertyType.ID = 22
			f.IsFinalApproved = false
			f.IsDenied = true
			f.RequestedDate, err = time.Parse(timeFormat, "2004-03-03 9:15:22")
			f.Update()
			So(f.ID, ShouldNotBeNil)
			f.Get()
			So(f, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(f.Name, ShouldEqual, "testTitle2")
			So(f.CustID, ShouldEqual, 123452)
			So(f.BadgeID, ShouldNotBeNil)
			So(f.Url, ShouldEqual, "www.test.com2")
			So(f.SellerID, ShouldEqual, "test2")
			So(f.WebPropertyType.ID, ShouldEqual, 22)
			So(f.IsFinalApproved, ShouldBeFalse)
			So(f.IsDenied, ShouldBeTrue)
			t, err = time.Parse(timeFormat, "2004-03-03 09:15:22")
			So(f.IsEnabledDate, ShouldHaveSameTypeAs, t)
			u, err = time.Parse(timeFormat, "2004-03-03 09:15:22")
			So(f.RequestedDate, ShouldHaveSameTypeAs, u)
			err = f.Delete()
			So(err, ShouldBeNil)
		})
		Convey("Testing Create(), Update(), Delete() WebPropNotes", func() {
			var n WebPropertyNote
			var f WebProperty
			f.ID = 248
			n.WebPropID = 248
			n.Text = "test note"
			c := make(chan int)
			go func() {
				n.Create()
				c <- 1
			}()
			<-c
			f.Get()
			So(f.Name, ShouldHaveSameTypeAs, "")
			var ns WebPropertyNotes
			So(f.WebPropertyNotes, ShouldHaveSameTypeAs, ns)

			n.Text = "Funk"
			err := n.Update()
			So(err, ShouldBeNil)

			err = n.Delete()
			So(err, ShouldBeNil)

		})
		Convey("Testing Create(), Update(), Delete() WebProperyRequirementsCheck", func() {
			var r WebPropertyRequirement
			var w WebProperty
			var err error
			w.ID = 248
			r.WebPropID = 248
			r.RequirementID = 1
			r.Compliance = true
			c := make(chan int)
			go func() {
				r.CreateJoin()
				c <- 1
			}()
			<-c
			err = w.Get()
			var rs WebPropertyRequirements
			So(w.WebPropertyRequirements, ShouldHaveSameTypeAs, rs)
			So(err, ShouldBeNil)

			r.Compliance = true
			err = r.UpdateJoin()
			So(err, ShouldBeNil)

			err = r.DeleteJoin()
			So(err, ShouldBeNil)
		})

		Convey("Testing Create(), Update(), Delete() Requirement", func() {
			var n WebPropertyRequirement
			n.ReqType = "Approved"
			n.Requirement = "TEST"
			err := n.Create()
			So(err, ShouldBeNil)
			n.Requirement = "booger"
			err = n.Update()
			So(err, ShouldBeNil)
			err = n.Delete()
			So(err, ShouldBeNil)
		})

		Convey("Testing Create(), Update(), Delete() Type", func() {
			var n WebPropertyType
			n.TypeID = 77
			n.Type = "TEST"
			err := n.Create()
			So(err, ShouldBeNil)
			n.Type = "booger"
			err = n.Update()
			So(err, ShouldBeNil)
			err = n.Delete()
			So(err, ShouldBeNil)

		})
	})
}
