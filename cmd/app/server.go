package app

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/asusg74/http/pkg/banners"
)

type Server struct {
	mux        *http.ServeMux
	bannersSvc *banners.Service
}

func NewServer(mux *http.ServeMux, bannersSvc *banners.Service) *Server {
	return &Server{mux: mux, bannersSvc: bannersSvc}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) Init() {
	s.mux.HandleFunc("/banners.getAll", s.handleGetAllBanners)
	s.mux.HandleFunc("/banners.getById", s.handleGetBannerByID)
	s.mux.HandleFunc("/banners.save", s.handleSaveBanner)
	s.mux.HandleFunc("/banners.removeById", s.handleRemoveByID)
}

func (s *Server) handleGetBannerByID(writer http.ResponseWriter, request *http.Request) {
	idParam := request.URL.Query().Get("id")
	log.Print("getting by id=" + idParam)

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err := s.bannersSvc.ByID(request.Context(), id)

	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleRemoveByID(writer http.ResponseWriter, request *http.Request) {
	idParam := request.URL.Query().Get("id")
	log.Print("delete by id = " + idParam)

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err := s.bannersSvc.RemoveByID(request.Context(), id)

	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleGetAllBanners(writer http.ResponseWriter, request *http.Request) {
	log.Print("looking for all banners")
	item, err := s.bannersSvc.All(request.Context())

	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleSaveBanner(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()

	if err != nil {
		log.Print(err)
	}

	err = request.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		log.Print(err)
	}

	fileReader, fileHeaders, err := request.FormFile("image")
	fileSent := 1
	if err != nil {
		fileSent = 0
		log.Print(err)
	}

	id, err := strconv.ParseInt(request.PostFormValue("id"), 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	result := &banners.Banner{
		ID:      s.bannersSvc.MaxID,
		Title:   request.PostFormValue("title"),
		Content: request.PostFormValue("content"),
		Button:  request.PostFormValue("button"),
		Link:    request.PostFormValue("link"),
	}
	var image []byte
	var imageName string
	if fileSent == 1 {
		image, err = ioutil.ReadAll(fileReader)
		if err != nil {
			log.Print(err)
			return
		}
		if id == 0 {
			imageName = strconv.Itoa(int(s.bannersSvc.MaxID)) + filepath.Ext(fileHeaders.Filename)
		} else {
			item, err := s.bannersSvc.ByID(request.Context(), id)
			if err != nil {
				log.Print(err)
				http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			imageName = item.Image
		}
		result.Image = imageName
		err = ioutil.WriteFile("web/banners/"+imageName, image, 0777)
		if err != nil {
			log.Print(err)
			return
		}
	}

	if id == 0 {
		s.bannersSvc.MaxID++
		item, err := s.bannersSvc.Save(request.Context(), result)
		result = item
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	} else {
		item, err := s.bannersSvc.ByID(request.Context(), id)
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		item.Title = request.PostFormValue("title")
		item.Content = request.PostFormValue("content")
		item.Button = request.PostFormValue("button")
		item.Link = request.PostFormValue("link")
		result = item
	}

	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}
