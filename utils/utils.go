package utils

func IsVideoFormat(filename string) bool {
	videoFormats := []string{".mp4", ".avi", ".mkv", ".mov", ".flv", ".wmv"}
	for _, format := range videoFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func IsAudioFormat(filename string) bool {
	audioFormats := []string{".mp3", ".wav", ".aac", ".flac", ".ogg", ".wma"}
	for _, format := range audioFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func IsImageFormat(filename string) bool {
	imageFormats := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff"}
	for _, format := range imageFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func IsDocumentFormat(filename string) bool {
	documentFormats := []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx"}
	for _, format := range documentFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func IsArchiveFormat(filename string) bool {
	archiveFormats := []string{".zip", ".rar", ".tar", ".gz", ".7z"}
	for _, format := range archiveFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func IsExecutableFormat(filename string) bool {
	executableFormats := []string{".exe", ".bat", ".sh", ".bin"}
	for _, format := range executableFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func IsFontFormat(filename string) bool {
	fontFormats := []string{".ttf", ".otf", ".woff", ".woff2", ".eot"}
	for _, format := range fontFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func IsCodeFormat(filename string) bool {
	codeFormats := []string{".go", ".py", ".js", ".java", ".cpp", ".c", ".html", ".css"}
	for _, format := range codeFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func IsTextFormat(filename string) bool {
	textFormats := []string{".txt", ".csv", ".log", ".md", ".xml", ".json"}
	for _, format := range textFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func IsSpreadsheetFormat(filename string) bool {
	spreadsheetFormats := []string{".xls", ".xlsx", ".ods"}
	for _, format := range spreadsheetFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func IsPresentationFormat(filename string) bool {
	presentationFormats := []string{".ppt", ".pptx", ".odp"}
	for _, format := range presentationFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func IsDatabaseFormat(filename string) bool {
	databaseFormats := []string{".db", ".sql", ".sqlite", ".mdb"}
	for _, format := range databaseFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func IsMarkupFormat(filename string) bool {
	markupFormats := []string{".html", ".xml", ".svg", ".xhtml"}
	for _, format := range markupFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func IsWebFormat(filename string) bool {
	webFormats := []string{".html", ".css", ".js", ".json", ".xml"}
	for _, format := range webFormats {
		if format == filename[len(filename)-len(format):] {
			return true
		}
	}
	return false
}

func CanBeOpenedWith(filename string) bool {
	if IsVideoFormat(filename) || IsAudioFormat(filename) || IsImageFormat(filename) ||
		IsDocumentFormat(filename) || IsArchiveFormat(filename) || IsExecutableFormat(filename) ||
		IsFontFormat(filename) || IsCodeFormat(filename) || IsTextFormat(filename) ||
		IsSpreadsheetFormat(filename) || IsPresentationFormat(filename) || IsDatabaseFormat(filename) ||
		IsMarkupFormat(filename) || IsWebFormat(filename) {
		return true
	}
	return false
}

func ConcurrentFileTypeCheck(filenames []string) map[string]bool {
	results := make(map[string]bool)
	for _, filename := range filenames {
		results[filename] = CanBeOpenedWith(filename)
	}
	return results
}
