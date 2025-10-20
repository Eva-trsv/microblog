package handlers

import "microblog/internal/storage"



var Store *storage.ObjectStorage

func InitHandlers(s *storage.ObjectStorage) {
	Store = s
}
