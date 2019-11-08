package server

import (
	"archive/zip"
	"encoding/json"
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
	r.HandleFunc("/tracks/upload", s.uploadTrack).Methods("POST")

	r.HandleFunc("/releases/", s.getReleases).Methods("GET")
	r.HandleFunc("/releases/", s.addRelease).
		Methods("POST").Headers("Content-type", "application/json")
	r.HandleFunc("/releases/{id:[0-9]+}", s.getRelease).Methods("GET")
	r.HandleFunc("/releases/{id:[0-9]+}", s.editRelease).
		Methods("POST").Headers("Content-type", "application/json")
	r.HandleFunc("/releases/upload", s.uploadRelease).Methods("POST")

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

func (s Server) uploadTrack(w http.ResponseWriter, r *http.Request) {
	tmp, err := ioutil.TempFile("", "blueshift-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	io.Copy(tmp, r.Body)

	var t models.Track
	s.makeTrack(&t, tmp)
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
		s.makeTrack(&t, ftmp)
		rel.AddTrack(t)
	}
	s.collection.CreateRelease(&rel)
}

func (s Server) makeTrack(t *models.Track, f io.ReadSeeker) {
	m := services.FileMetadata(f)
	p, _ := m.Track()
	d, _ := m.Disc()
	format := s.collection.GetFormat(string(m.FileType()))

	t.Title = m.Title()
	t.Position = p
	t.Disc = d

	path := s.streamhdlr.Store(f)
	t.AddStream(models.Stream{Path: path, Format: format})
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
