package repo

import (
	entity "go-cinema/entities"
	"sync"

	"github.com/misenkashari/goutils/repository"
	"gorm.io/gorm"
)

var (
	MovieRepository   *repository.GormRepository[entity.Movie, uint]
	SeriesRepository  *repository.GormRepository[entity.Series, uint]
	EpisodeRepository *repository.GormRepository[entity.Episode, uint]
	once              sync.Once
)

func InitRepositories(db *gorm.DB) {
	once.Do(func() {
		MovieRepository = repository.Gorm[entity.Movie, uint](db)
		SeriesRepository = repository.Gorm[entity.Series, uint](db)
		EpisodeRepository = repository.Gorm[entity.Episode, uint](db)
	})
}
