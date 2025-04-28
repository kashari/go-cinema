package entity

import (
	"os"

	"github.com/kashari/golog"
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

type Series struct {
	gorm.Model
	Title        string    `json:"Title"`
	Description  string    `json:"Description"`
	BaseDir      string    `json:"BaseDir" gorm:"not null"`
	Episodes     []Episode `json:"Episodes"`
	CurrentIndex uint      `json:"CurrentIndex" gorm:"not null"`
}

type Episode struct {
	gorm.Model
	Path         string `json:"Path" gorm:"not null"`
	ResumeAt     string `json:"ResumeAt"`
	EpisodeIndex int    `json:"EpisodeIndex" gorm:"not null"`
	SeriesID     uint   `json:"series_id"`
}

type SeriesRequest struct {
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

func ServeVideo(name string) (*os.File, error) {
	file, err := os.Open(name)
	if err != nil {
		golog.Error("Error opening video file", err)
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

func OpenNBytes(file *os.File, offset int64, n int64) ([]byte, error) {
	buf := make([]byte, n)
	_, err := file.ReadAt(buf, offset)
	if err != nil {
		golog.Error("Error reading file", err)
		return nil, err
	}
	return buf, nil
}

func CopyNBytes(dst *os.File, src *os.File, n int64) (int64, error) {
	buf := make([]byte, n)
	nr, err := src.Read(buf)
	if err != nil {
		golog.Error("Error reading file", err)
		return 0, err
	}
	nw, err := dst.Write(buf[:nr])
	if err != nil {
		golog.Error("Error writing file", err)
		return 0, err
	}
	return int64(nw), nil
}

func FileSeekNBytes(file *os.File, offset int64, whence int) (int64, error) {
	n, err := file.Seek(offset, whence)
	if err != nil {
		golog.Error("Error seeking file", err)
		return 0, err
	}
	return n, nil
}

func FileClose(file *os.File) error {
	err := file.Close()
	if err != nil {
		golog.Error("Error closing file", err)
		return err
	}
	return nil
}

type Option[T any] struct {
	Some T
	None bool
}

func NewOption[T any](value T) Option[T] {
	return Option[T]{Some: value, None: false}
}

func NewNone[T any]() Option[T] {
	return Option[T]{None: true}
}

func (o Option[T]) IsSome() bool {
	return !o.None
}

func (o Option[T]) IsNone() bool {
	return o.None
}

func (o Option[T]) Unwrap() T {
	if o.None {
		golog.Info("Unwrap called on None option")
		return *new(T)
	}
	return o.Some
}

func (o Option[T]) UnwrapOr(defaultValue T) T {
	if o.None {
		return defaultValue
	}
	return o.Some
}

func (o Option[T]) UnwrapOrElse(f func() T) T {
	if o.None {
		return f()
	}
	return o.Some
}

func (o Option[T]) Map(f func(T) T) Option[T] {
	if o.None {
		return o
	}
	return NewOption(f(o.Some))
}

func (o Option[T]) MapOr(defaultValue T, f func(T) T) T {
	if o.None {
		return defaultValue
	}
	return f(o.Some)
}

func (o Option[T]) MapOrElse(defaultValue func() T, f func(T) T) T {
	if o.None {
		return defaultValue()
	}
	return f(o.Some)
}
