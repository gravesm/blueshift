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
	templates  string
}

func NewServer(c models.Collection, sh services.StreamHandler, tmpl string) *mux.Router {
	s := &Server{collection: c, streamhdlr: sh, templates: tmpl}

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
	trks := s.collection.Tracks(0, 10)
	tmpl := path.Join(s.templates, "tracks.html")
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		log.Fatal(err)
	}
	err = t.Execute(w, trks)
	if err != nil {
		log.Fatal(err)
	}
}

func (s Server) getTrack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	trk := s.collection.GetTrack(id)
	tmpl := path.Join(s.templates, "track.html")
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		log.Fatal(err)
	}
	err = t.Execute(w, trk)
	if err != nil {
		log.Fatal(err)
	}
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
	rels := s.collection.Releases(0, 10)
	tmpl := path.Join(s.templates, "releases.html")
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		log.Fatal(err)
	}
	err = t.Execute(w, rels)
	if err != nil {
		log.Fatal(err)
	}
}

func (s Server) getRelease(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	rel := s.collection.GetRelease(id)
	tmpl := path.Join(s.templates, "release.html")
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		log.Fatal(err)
	}
	err = t.Execute(w, rel)
	if err != nil {
		log.Fatal(err)
	}
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
