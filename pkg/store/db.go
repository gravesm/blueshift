package store

import (
	"github.com/dhowden/tag"
	"github.com/gravesm/blueshift/pkg/models"
	"github.com/jinzhu/gorm"
	"log"
)

const (
	unknown = string(tag.UnknownFileType)
	ogg     = string(tag.OGG)
	mp3     = string(tag.MP3)
	flac    = string(tag.FLAC)
)

type DbCollection struct {
	handler *gorm.DB
}

func (db DbCollection) GetFormat(name string) models.Format {
	var f models.Format
	err := db.handler.Where("name = ?", name).First(&f).Error
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func (db DbCollection) CreateRelease(release *models.Release) {
	db.handler.Create(release)
}

func (db DbCollection) SaveRelease(release models.Release) {
	db.handler.Save(release)
}

func (db DbCollection) GetRelease(id int64) models.Release {
	var r models.Release
	db.handler.Preload("Tracks").First(&r, id)
	return r
}

func (db DbCollection) Releases(offset int, rows int) []models.Release {
	var releases []models.Release
	db.handler.Order("id desc").Offset(offset).Limit(rows).Find(&releases)
	return releases
}

func (db DbCollection) CreateTrack(t *models.Track) {
	db.handler.Create(t)
}

func (db DbCollection) SaveTrack(t models.Track) {
	db.handler.Save(t)
}

func (db DbCollection) GetTrack(id int64) models.Track {
	var t models.Track
	err := db.handler.Preload("Streams").First(&t, id).Error
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func (db DbCollection) Tracks(offset int, rows int) []models.Track {
	var tracks []models.Track
	db.handler.Order("id desc").Offset(offset).Limit(rows).Find(&tracks)
	return tracks
}

func (db DbCollection) CreateArtist(artist *models.Artist) {
}

func (db DbCollection) SaveArtist(artist models.Artist) {
}

func (db DbCollection) GetArtist(id int64) models.Artist {
	a := models.Artist{}
	return a
}

func (db DbCollection) Artists(offset int, rows int) []models.Artist {
	arts := []models.Artist{}
	return arts
}

func Migrate(db *gorm.DB) {
	db.AutoMigrate(&models.Track{}, &models.Stream{}, &models.Format{},
		&models.Release{})
}

func Initialize(db *gorm.DB) {
	Migrate(db)
	db.Create(&models.Format{Name: unknown})
	db.Create(&models.Format{Name: mp3})
	db.Create(&models.Format{Name: ogg})
	db.Create(&models.Format{Name: flac})
}

func NewDbCollection(handler *gorm.DB) models.Collection {
	return DbCollection{handler: handler}
}
