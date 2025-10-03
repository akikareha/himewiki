package action

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/data"
	"github.com/akikareha/himewiki/internal/filter"
	"github.com/akikareha/himewiki/internal/util"
)

func ViewImage(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	_, image, err := data.LoadImage(params.DbName)
	if err != nil {
		http.Redirect(w, r, "/"+url.PathEscape(params.Name)+"?b=upload", http.StatusFound)
		return
	}

	mimeType := mime.TypeByExtension("." + params.Ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimeType)

	w.Write(image)
}

func fileToBytes(fh *multipart.FileHeader) ([]byte, error) {
	file, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}
	return data, nil
}

func Upload(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	name := ""
	if r.Method == http.MethodPost {
		err := r.ParseMultipartForm(10 << 20) // max 10 MiB
		if err != nil {
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}

		name = r.FormValue("name")
		if name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "failed to read file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		image, err := fileToBytes(header)
		if err != nil {
			http.Error(w, "Failed to load image", http.StatusInternalServerError)
			return
		}

		filtered, err := filter.ImageApply(cfg, name, image)
		if err != nil {
			http.Error(w, "Failed to filter image", http.StatusInternalServerError)
			return
		}

		if err := data.SaveImage(cfg, name, filtered); err != nil {
			http.Error(w, "Failed to save", http.StatusInternalServerError)
			return
		}
	}

	tmpl := util.NewTemplate("upload.html")
	tmpl.Execute(w, struct {
		Base string
		SiteName string
		Name string
	}{
		Base: cfg.Site.Base,
		SiteName: cfg.Site.Name,
		Name: name,
	})
}

const imagesPerPage = 200

func AllImages(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	pageStr := r.URL.Query().Get("p")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	images, err := data.LoadAllImages(page, imagesPerPage)
	if err != nil {
		http.Error(w, "Failed to load images", http.StatusInternalServerError)
		return
	}

	tmpl := util.NewTemplate("allimgs.html")
	tmpl.Execute(w, struct {
		Base string
		SiteName string
		Images []string
		NextPage int
	}{
		Base: cfg.Site.Base,
		SiteName: cfg.Site.Name,
		Images: images,
		NextPage: page + 1,
	})
}
