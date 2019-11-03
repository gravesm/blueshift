package server

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gravesm/blueshift/pkg/models"
	"github.com/gravesm/blueshift/pkg/services"
	"github.com/gravesm/blueshift/pkg/store"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestServer(t *testing.T) {
	Convey("Test Server", t, func() {
		db, err := gorm.Open("sqlite3", ":memory:")
		if err != nil {
			panic(err)
		}
		defer db.Close()
		tmp, err := ioutil.TempDir("", "blueshift-")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(tmp)

		coll := store.NewDbCollection(db)
		store.Initialize(db)
		s := Server{collection: coll, streamhdlr: services.FileStreamHandler{tmp}, templates: "../../templates"}

		Convey("should list tracks", func() {
			coll.CreateTrack(&models.Track{Title: "Track 1"})
			coll.CreateTrack(&models.Track{Title: "Track 2"})
			req, _ := http.NewRequest("GET", "/tracks/", nil)
			rec := httptest.NewRecorder()
			hdlr := http.HandlerFunc(s.getTracks)
			hdlr.ServeHTTP(rec, req)
			body, _ := ioutil.ReadAll(rec.Body)
			html := string(body)
			So(rec.Code, ShouldEqual, http.StatusOK)
			So(strings.Contains(html, "Track 1"), ShouldBeTrue)
			So(strings.Contains(html, "Track 2"), ShouldBeTrue)
		})

		Convey("should show track", func() {
			t := models.Track{Title: "Track 1"}
			coll.CreateTrack(&t)
			req, _ := http.NewRequest("GET", "/tracks/", nil)
			req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(t.ID, 10)})
			rec := httptest.NewRecorder()
			hdlr := http.HandlerFunc(s.getTrack)
			hdlr.ServeHTTP(rec, req)
			body, _ := ioutil.ReadAll(rec.Body)
			So(rec.Code, ShouldEqual, http.StatusOK)
			So(strings.Contains(string(body), "Track 1"), ShouldBeTrue)
		})

		Convey("should add track", func() {
			post, _ := json.Marshal(&models.Track{Title: "Track 1", Position: 1})
			req, _ := http.NewRequest("POST", "/tracks/", bytes.NewReader(post))
			rec := httptest.NewRecorder()
			hdlr := http.HandlerFunc(s.addTrack)
			hdlr.ServeHTTP(rec, req)
			var t models.Track
			db.First(&t)
			So(rec.Code, ShouldEqual, http.StatusOK)
			So(t.Title, ShouldEqual, "Track 1")
			So(t.Position, ShouldEqual, 1)
			So(t.Disc, ShouldEqual, 0)
		})

		Convey("should edit track", func() {
			t := models.Track{Title: "Track 1", Position: 1}
			t.AddStream(models.Stream{Path: "foo/bar"})
			coll.CreateTrack(&t)
			post := `{"title": "Track 3", "position": 3}`
			req, _ := http.NewRequest("POST", "/tracks/", strings.NewReader(post))
			req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(t.ID, 10)})
			rec := httptest.NewRecorder()
			hdlr := http.HandlerFunc(s.editTrack)
			hdlr.ServeHTTP(rec, req)
			var trk models.Track
			db.Preload("Streams").First(&trk)
			So(rec.Code, ShouldEqual, http.StatusOK)
			So(trk.Title, ShouldEqual, "Track 3")
			So(trk.Position, ShouldEqual, 3)
			So(trk.Disc, ShouldEqual, 0)
			So(len(trk.Streams), ShouldEqual, 1)
		})

		Convey("should add track from upload", func() {
			f, err := os.Open("../testdata/magic_flute.ogg")
			if err != nil {
				panic(err)
			}
			req, _ := http.NewRequest("POST", "/tracks/upload", f)
			rec := httptest.NewRecorder()
			hdlr := http.HandlerFunc(s.uploadTrack)
			hdlr.ServeHTTP(rec, req)
			var t models.Track
			db.First(&t)
			So(rec.Code, ShouldEqual, http.StatusOK)
			So(t.Title, ShouldEqual, "Der Hölle Rache kocht in meinem Herzen")
			So(t.Position, ShouldEqual, 18)
			So(t.Disc, ShouldEqual, 0)
		})

		Convey("should list releases", func() {
			coll.CreateRelease(&models.Release{Title: "Release 1"})
			coll.CreateRelease(&models.Release{Title: "Release 2"})
			req, _ := http.NewRequest("POST", "/releases/", nil)
			rec := httptest.NewRecorder()
			hdlr := http.HandlerFunc(s.getReleases)
			hdlr.ServeHTTP(rec, req)
			body, _ := ioutil.ReadAll(rec.Body)
			html := string(body)
			So(rec.Code, ShouldEqual, http.StatusOK)
			So(strings.Contains(html, "Release 1"), ShouldBeTrue)
			So(strings.Contains(html, "Release 2"), ShouldBeTrue)
		})

		Convey("should show release", func() {
			r := models.Release{Title: "Release 1"}
			coll.CreateRelease(&r)
			req, _ := http.NewRequest("POST", "/releases/", nil)
			req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(r.ID, 10)})
			rec := httptest.NewRecorder()
			hdlr := http.HandlerFunc(s.getRelease)
			hdlr.ServeHTTP(rec, req)
			body, _ := ioutil.ReadAll(rec.Body)
			So(rec.Code, ShouldEqual, http.StatusOK)
			So(strings.Contains(string(body), "Release 1"), ShouldBeTrue)
		})

		Convey("should add release", func() {
			post, _ := json.Marshal(&models.Release{Title: "Release 1"})
			req, _ := http.NewRequest("POST", "/releases/", bytes.NewReader(post))
			rec := httptest.NewRecorder()
			hdlr := http.HandlerFunc(s.addRelease)
			hdlr.ServeHTTP(rec, req)
			var r models.Release
			db.First(&r)
			So(rec.Code, ShouldEqual, http.StatusOK)
			So(r.Title, ShouldEqual, "Release 1")
		})

		Convey("should edit release", func() {
			r := models.Release{Title: "Release 1"}
			r.AddTrack(models.Track{Title: "Title 1"})
			coll.CreateRelease(&r)
			post := `{"title": "Release 2"}`
			req, _ := http.NewRequest("POST", "/releases/", strings.NewReader(post))
			req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(r.ID, 10)})
			rec := httptest.NewRecorder()
			hdlr := http.HandlerFunc(s.editRelease)
			hdlr.ServeHTTP(rec, req)
			var rel models.Release
			db.Preload("Tracks").First(&rel)
			So(rec.Code, ShouldEqual, http.StatusOK)
			So(rel.Title, ShouldEqual, "Release 2")
			So(len(rel.Tracks), ShouldEqual, 1)
		})

		SkipConvey("should add release from upload", func() {})
	})
}
