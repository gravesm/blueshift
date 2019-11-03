package store

import (
	"github.com/gravesm/blueshift/pkg/models"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDb(t *testing.T) {
	Convey("Test DB CollectionStore", t, func() {
		db, err := gorm.Open("sqlite3", ":memory:")
		if err != nil {
			panic(err)
		}
		defer db.Close()

		store := NewDbCollection(db)
		Initialize(db)

		Convey("should return format", func() {
			f := store.GetFormat(flac)
			So(f.Name, ShouldEqual, flac)
		})

		Convey("should create release", func() {
			r := models.Release{Title: "New release"}
			r.AddTrack(models.Track{Title: "Track 1"})
			r.AddTrack(models.Track{Title: "Track 2"})
			store.CreateRelease(&r)
			var release models.Release
			db.Preload("Tracks").First(&release)
			So(release.Title, ShouldEqual, "New release")
			So(len(release.Tracks), ShouldEqual, 2)
		})

		Convey("should save release", func() {
			r := models.Release{Title: "New release"}
			store.CreateRelease(&r)
			r.Title = "Old release"
			store.SaveRelease(r)
			var release models.Release
			db.First(&release)
			So(release.Title, ShouldEqual, "Old release")
		})

		Convey("should retrieve release", func() {
			r := models.Release{Title: "New release"}
			r.AddTrack(models.Track{Title: "Track 1"})
			store.CreateRelease(&r)
			release := store.GetRelease(r.ID)
			So(release.Title, ShouldEqual, "New release")
			So(len(release.Tracks), ShouldEqual, 1)
		})

		Convey("should retrieve releases", func() {
			store.CreateRelease(&models.Release{Title: "Release 1"})
			store.CreateRelease(&models.Release{Title: "Release 2"})
			store.CreateRelease(&models.Release{Title: "Release 3"})
			releases := store.Releases(0, 2)
			So(len(releases), ShouldEqual, 2)
			So(releases[1].Title, ShouldEqual, "Release 2")
		})

		Convey("should create track", func() {
			var format models.Format
			db.Where("name = ?", ogg).First(&format)
			t := models.Track{
				Title:    "Foobar",
				Position: 1,
				Disc:     1,
			}
			t.AddStream(models.Stream{Path: "foo/bar", Format: format})
			store.CreateTrack(&t)
			var track models.Track
			db.Preload("Streams").First(&track, t.ID)
			So(track.Title, ShouldEqual, "Foobar")
			So(len(track.Streams), ShouldEqual, 1)
		})

		Convey("should save track", func() {
			t := models.Track{Title: "Track 1"}
			t.AddStream(models.Stream{Path: "foo/bar"})
			store.CreateTrack(&t)
			t.Title = "Track 2"
			store.SaveTrack(t)
			var trk models.Track
			db.Preload("Streams").First(&trk)
			So(trk.Title, ShouldEqual, "Track 2")
			So(len(trk.Streams), ShouldEqual, 1)
		})

		Convey("should retrieve track", func() {
			t := models.Track{Title: "Track 1"}
			t.AddStream(models.Stream{Path: "foo/bar"})
			store.CreateTrack(&t)
			trk := store.GetTrack(t.ID)
			So(trk.Title, ShouldEqual, "Track 1")
			So(len(trk.Streams), ShouldEqual, 1)
		})

		Convey("should retrieve tracks", func() {
			store.CreateTrack(&models.Track{Title: "Track 1"})
			store.CreateTrack(&models.Track{Title: "Track 2"})
			store.CreateTrack(&models.Track{Title: "Track 3"})
			trks := store.Tracks(1, 10)
			So(len(trks), ShouldEqual, 2)
			So(trks[0].Title, ShouldEqual, "Track 2")
		})
	})

}
