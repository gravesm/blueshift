package models

type Collection interface {
	GetFormat(name string) Format

	CreateRelease(release *Release)
	SaveRelease(release Release)
	GetRelease(id int64) Release
	Releases(offset int, rows int) []Release

	CreateTrack(track *Track)
	SaveTrack(track Track)
	GetTrack(id int64) Track
	Tracks(offset int, rows int) []Track

	CreateArtist(artist *Artist)
	SaveArtist(artist Artist)
	GetArtist(id int64) Artist
	Artists(offset int, rows int) []Artist
}

type Format struct {
	ID   int64
	Name string
}

type Release struct {
	ID      int64
	MBID    string
	Title   string
	Year    int
	Tracks  []Track
	Artists []Artist `gorm:"many2many:user_languages;"`
}

type Track struct {
	ID        int64
	MBID      string
	Title     string
	Position  int
	Disc      int
	Artists   []Artist `gorm:"many2many:user_languages;"`
	Streams   []Stream
	ReleaseID int64
}

type Stream struct {
	ID       int64
	Path     string
	Format   Format `gorm:"association_autoupdate:false"`
	FormatID int64
	TrackID  int64
}

type Artist struct {
	ID   int64
	Name string
}

func (r *Release) AddTrack(track Track) {
	r.Tracks = append(r.Tracks, track)
}

func (r *Release) AddArtist(artist Artist) {
	r.Artists = append(r.Artists, artist)
}

func (t *Track) AddArtist(artist Artist) {
	t.Artists = append(t.Artists, artist)
}

func (t *Track) AddStream(s Stream) {
	t.Streams = append(t.Streams, s)
}
