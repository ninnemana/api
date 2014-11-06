package testimonials

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTestimonials(t *testing.T) {
	var test Testimonial
	var err error
	Convey("Testing Create Testimonial", t, func() {
		test.Content = "Test Content"
		err = test.Create()
		So(err, ShouldBeNil)
	})
	Convey("Update", t, func() {
		test.Content = "New Content"
		test.Active = true
		test.Approved = true
		err = test.Update()
		So(err, ShouldBeNil)

	})

	Convey("GetAll - No paging", t, func() {
		err = test.Get()
		So(err, ShouldBeNil)
	})
	Convey("GetAll - No paging", t, func() {
		ts, err := GetAllTestimonials(0, 1, false)
		So(err, ShouldBeNil)
		So(len(ts), ShouldBeGreaterThan, 0)

	})

	Convey("GetAll - Paged", t, func() {
		ts, err := GetAllTestimonials(0, 1, false)

		So(err, ShouldBeNil)
		So(len(ts), ShouldBeGreaterThan, 0)

	})

	Convey("GetAll - randomized", t, func() {
		ts, err := GetAllTestimonials(0, 1, true)
		So(err, ShouldBeNil)
		So(len(ts), ShouldBeGreaterThan, 0)

	})
	Convey("Delete", t, func() {
		err = test.Delete()
		So(err, ShouldBeNil)

	})

}
