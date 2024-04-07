package theatre

import (
	"log"
	"os"

	"gorm.io/gorm"
)

type Movie struct {
	gorm.Model
	Title       string `json:"Title" gorm:"not null"`
	Description string `json:"Description"`
	Path        string `json:"Path" gorm:"not null"`
	ResumeAt    string `json:"ResumeAt"`
}

type MovieRequest struct {
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

type Episode struct {
	gorm.Model
	Path         string `json:"Path" gorm:"not null"`
	ResumeAt     string `json:"ResumeAt"`
	EpisodeIndex int    `json:"EpisodeIndex" gorm:"not null"`
	SeriesID     uint   `json:"series_id" gorm:"not null"`
	Series       Series `json:"Series" gorm:"foreignKey:SeriesID;references:ID"`
}

type Series struct {
	gorm.Model
	Title        string    `json:"Title" gorm:"not null"`
	Description  string    `json:"Description"`
	BaseDir      string    `json:"BaseDir" gorm:"not null"`
	Episodes     []Episode `json:"Episodes" gorm:"foreignKey:SeriesID;references:ID"`
	CurrentIndex uint      `json:"CurrentIndex" gorm:"not null"`
}

func ServeVideo(name string) (*os.File, error) {
	file, err := os.Open(name)
	if err != nil {
		log.Println("Error opening video file", err)
		return nil, err
	}
	return file, nil
}

func GetFileSize(file *os.File) int64 {
	info, err := file.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}
