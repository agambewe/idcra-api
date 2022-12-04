package handler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/kerti/idcra-api/model"
	"github.com/kerti/idcra-api/service"
	uuid "github.com/satori/go.uuid"
)

func SurveyReport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// w.Header().Set("Content-type", "application/octet-stream")
		w.Header().Set("Content-type", "application/pdf")

		ctx := r.Context()
		filename := strings.TrimPrefix(r.URL.Path, "/reports/surveys/")
		idString := strings.Replace(filename, ".pdf", "", -1)
		id, err := uuid.FromString(idString)
		if err != nil {
			response := &model.Response{
				Code:  http.StatusBadRequest,
				Error: err.Error(),
			}
			writeResponse(w, response, response.Code)
			return
		}

		reportData, err := ctx.Value("reportService").(*service.ReportService).GenerateSurveyPDF(id)
		if err != nil {
			response := &model.Response{
				Code:  http.StatusInternalServerError,
				Error: err.Error(),
			}
			writeResponse(w, response, response.Code)
			return
		}

		reportBytes := bytes.NewReader(reportData.Bytes())
		io.Copy(w, reportBytes)
	})
}

func SchoolReport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// w.Header().Set("Content-type", "application/octet-stream")
		w.Header().Set("Content-type", "application/json")

		ctx := r.Context()
		schoolId := strings.TrimPrefix(r.URL.Path, "/reports/school/")

		os.RemoveAll("./reports/")
		os.MkdirAll("./reports/", os.ModePerm)

		err := ctx.Value("reportService").(*service.ReportService).GenerateSchoolReport(schoolId)
		if err != nil {
			response := &model.Response{
				Code:  http.StatusInternalServerError,
				Error: err.Error(),
			}
			writeResponse(w, response, response.Code)
			return
		}

		school, err := ctx.Value("schoolService").(*service.SchoolService).FindByID(schoolId)
		if err != nil {
			response := &model.Response{
				Code:  http.StatusInternalServerError,
				Error: err.Error(),
			}
			writeResponse(w, response, response.Code)
			return
		}

		err = os.Rename("./schoolreports.zip", fmt.Sprintf("./reports/%s.zip", school.Name))
		if err != nil {
			response := &model.Response{
				Code:  http.StatusInternalServerError,
				Error: err.Error(),
			}
			writeResponse(w, response, response.Code)
			return
		}

		os.RemoveAll("./tmp/")
		os.MkdirAll("./tmp/", os.ModePerm)

		response := &model.Response{
			Code: http.StatusOK,
		}

		writeResponse(w, response, response.Code)
		return
	})
}

func DownloadZip() ([]byte, error) {
	r, w := io.Pipe()

	defer r.Close()
	defer w.Close()

	zip, err := os.Stat("./schoolreports.zip")
	if err != nil {
		return nil, err
	}

	go func() {

		f, err := os.Open(zip.Name())
		if err != nil {
			return

		}

		buf := make([]byte, 1024)
		for {
			chunk, err := f.Read(buf)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if chunk == 0 {
				break
			}

			if _, err := w.Write(buf[:chunk]); err != nil {
				return
			}

		}

		w.Close()
	}()

	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return body, nil

}
