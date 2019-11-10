package server

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/dhowden/tag"
	"github.com/gorilla/mux"
	"github.com/gravesm/blueshift/pkg/models"
	"github.com/gravesm/blueshift/pkg/services"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

type Server struct {
	collection models.Collection
	streamhdlr services.StreamHandler
	templates  map[string]*template.Template
}

func NewServer(c models.Collection, sh services.StreamHandler, tmpl string) *mux.Router {
	templates := loadTemplates(tmpl)
	s := &Server{collection: c, streamhdlr: sh, templates: templates}

	r := mux.NewRouter()
	r.HandleFunc("/tracks/", s.getTracks).Methods("GET")
	r.HandleFunc("/tracks/", s.addTrack).
		Methods("POST").Headers("Content-type", "application/json")
	r.HandleFunc("/tracks/{id:[0-9]+}", s.getTrack).Methods("GET")
	r.HandleFunc("/tracks/{id:[0-9]+}", s.editTrack).
		Methods("POST").Headers("Content-type", "application/json")
	r.HandleFunc("/tracks/{id:[0-9]+}/stream", s.stream).Methods("GET")
	r.HandleFunc("/tracks/upload", s.uploadTrack).Methods("POST")

	r.HandleFunc("/releases/", s.getReleases).Methods("GET")
	r.HandleFunc("/releases/", s.addRelease).
		Methods("POST").Headers("Content-type", "application/json")
	r.HandleFunc("/releases/{id:[0-9]+}", s.getRelease).Methods("GET")
	r.HandleFunc("/releases/{id:[0-9]+}", s.editRelease).
		Methods("POST").Headers("Content-type", "application/json")
	r.HandleFunc("/releases/upload", s.uploadRelease).Methods("POST")

	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return r
}

func (s Server) getTracks(w http.ResponseWriter, r *http.Request) {
	s.render("track/index", w, s.collection.Tracks(0, 10))
}

func (s Server) getTrack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	s.render("track/track", w, s.collection.GetTrack(id))
}

func (s Server) addTrack(w http.ResponseWriter, r *http.Request) {
	var t models.Track
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		log.Fatal(err)
	}
	s.collection.CreateTrack(&t)
}

func (s Server) editTrack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	t := s.collection.GetTrack(id)
	err = json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		log.Fatal(err)
	}
	s.collection.SaveTrack(t)
}

func (s Server) stream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	t := s.collection.GetTrack(id)
	strm := t.Streams[0]
	w.Header().Set("Content-type", strm.Format.Mimetype)
	http.ServeFile(w, r, strm.Path)
}

func (s Server) uploadTrack(w http.ResponseWriter, r *http.Request) {
	tmp, err := ioutil.TempFile("", "blueshift-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	io.Copy(tmp, r.Body)

	var t models.Track
	var strm models.Stream
	meta := services.FileMetadata(tmp)
	s.makeTrack(&t, meta)
	s.makeStream(&strm, meta, tmp)
	t.AddStream(strm)
	s.collection.CreateTrack(&t)

	encoder := json.NewEncoder(w)
	encoder.Encode(t)
}

func (s Server) getReleases(w http.ResponseWriter, r *http.Request) {
	s.render("release/index", w, s.collection.Releases(0, 10))
}

func (s Server) getRelease(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	s.render("release/release", w, s.collection.GetRelease(id))
}

func (s Server) addRelease(w http.ResponseWriter, r *http.Request) {
	var rel models.Release
	err := json.NewDecoder(r.Body).Decode(&rel)
	if err != nil {
		log.Fatal(err)
	}
	s.collection.CreateRelease(&rel)
}

func (s Server) editRelease(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	rel := s.collection.GetRelease(id)
	err = json.NewDecoder(r.Body).Decode(&rel)
	if err != nil {
		log.Fatal(err)
	}
	s.collection.SaveRelease(rel)
}

func (s Server) uploadRelease(w http.ResponseWriter, r *http.Request) {
	tmp, err := ioutil.TempFile("", "blueshift-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	io.Copy(tmp, r.Body)
	arxv, err := zip.OpenReader(tmp.Name())
	if err != nil {
		log.Fatal(err)
	}
	defer arxv.Close()
	var rel models.Release
	for _, f := range arxv.File {
		var t models.Track
		var strm models.Stream
		trk, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer trk.Close()
		ftmp, err := ioutil.TempFile("", "blueshift-")
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(ftmp.Name())
		io.Copy(ftmp, trk)
		meta := services.FileMetadata(ftmp)
		s.makeRelease(&rel, meta)
		s.makeTrack(&t, meta)
		s.makeStream(&strm, meta, ftmp)
		t.AddStream(strm)
		rel.AddTrack(t)
	}

	s.collection.CreateRelease(&rel)
}

func (s Server) makeTrack(t *models.Track, m tag.Metadata) {
	raw := m.Raw()
	p, _ := m.Track()
	d, _ := m.Disc()
	t.Title = m.Title()
	t.Position = p
	t.Disc = d
	t.MBID = fmt.Sprintf("%v", raw["musicbrainz_trackid"])
}

func (s Server) makeStream(strm *models.Stream, m tag.Metadata, f io.Reader) {
	path := s.streamhdlr.Store(f)
	strm.Path = path
	strm.Format = s.collection.GetFormat(string(m.FileType()))
}

func (s Server) makeRelease(r *models.Release, m tag.Metadata) {
	raw := m.Raw()
	r.Title = m.Album()
	r.MBID = fmt.Sprintf("%v", raw["musicbrainz_albumid"])
	r.Year, _ = strconv.Atoi(fmt.Sprintf("%v", raw["originalyear"]))
}

func (s Server) render(tmpl string, w http.ResponseWriter, ctx interface{}) {
	err := s.templates[tmpl].ExecuteTemplate(w, "base", ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func loadTemplates(root string) map[string]*template.Template {
	templates := make(map[string]*template.Template)
	base := template.Must(template.ParseGlob(path.Join(root, "base.html")))
	tmpls := []string{"release/index", "release/release", "track/index", "track/track"}
	for _, t := range tmpls {
		b, err := base.Clone()
		if err != nil {
			log.Fatal(err)
		}
		b.ParseFiles(path.Join(root, t+".html.tmpl"))
		templates[t] = b
	}
	return templates
}
