package service

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	"github.com/kerti/idcra-api/model"
	"github.com/op/go-logging"
	uuid "github.com/satori/go.uuid"
	"github.com/tealeg/xlsx"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

type ReportService struct {
	db  *sqlx.DB
	log *logging.Logger
}

func NewReportService(db *sqlx.DB, log *logging.Logger) *ReportService {
	return &ReportService{db: db, log: log}
}

func (s *ReportService) CostBreakdownBySchoolAndDateRange(schoolID string, startDate string, endDate string) ([]*model.CostReport, error) {
	results := make([]*model.CostReport, 0)

	reportSQL := `
	select
		d.action description,
		sum(d.unit_cost) cost
	from
		cases c
		left join surveys s on c.survey_id = s.id
		left join students st on s.student_id = st.id
		left join diagnosis_and_actions d on c.diagnosis_and_action_id = d.id
	where
		st.school_id = ?
		and s.date >= ?
		and s.date < ?
	group by
		d.action;`

	err := s.db.Select(&results, reportSQL, schoolID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var totalCost float64
	for _, r := range results {
		totalCost += r.Cost
	}

	summary := &model.CostReport{
		Description: "Total",
		Cost:        totalCost,
	}

	results = append(results, summary)

	return results, nil
}

func (s *ReportService) GenerateSchoolReport(schoolId string) (err error) {
	models := []model.SchoolReports{}
	reportSQL := `select s.id, students.name, s.date from students join surveys s on students.id = s.student_id where school_id = ?;`

	err = s.db.Select(&models, reportSQL, schoolId)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, val := range models {
			id, err := uuid.FromString(val.ID)
			if err != nil {
				return
			}

			reportByte, err := s.GenerateSurveyPDF(id)
			if err != nil {
				return
			}

			err = os.Mkdir(fmt.Sprintf("./tmp/%s", val.Name), os.ModePerm)
			f, err := os.Create(fmt.Sprintf("./tmp/%s/%s.pdf", val.Name, val.Date.Format("2006-01-02")))

			defer f.Close()
			if err != nil {
				return
			}

			_, err = f.Write(reportByte.Bytes())
			if err != nil {
				return
			}

		}
	}()

	wg.Wait()

	err = zipSource("./tmp/", "schoolreports.zip")
	return err
}

func (s *ReportService) GenerateSurveyPDF(surveyID uuid.UUID) (reportData bytes.Buffer, err error) {
	models := []model.SurveyReport{}
	reportSQL := `
		select
			student.name studentname,
			school.name schoolname,
			s.date dateofsurvey,
			s.subjective_score scapercentage,
			s.upper_d dvalue,
			s.upper_m mvalue,
			s.upper_f fvalue
		from
			surveys s
			left join students student
				on s.student_id = student.id
			left join schools school
				on student.school_id = school.id
		where
			s.id = ?;`

	err = s.db.Select(&models, reportSQL, surveyID)
	if err != nil {
		return *bytes.NewBufferString(""), err
	}
	println(fmt.Sprintf("MODEL: %v", models))

	if len(models) != 1 {
		return *bytes.NewBufferString(""), err
	}

	model := models[0]
	model.Setup()
	reportData, err = getReport(model)
	return
}

func (s *ReportService) GenerateSurveyCSV(schoolId string, data *model.QuestionData) (err error) {
	models := []model.SurveyReportCSV{}

	reportSQL := `
		select
			student.name studentname, school.name schoolname,
			s.s1q1,s.s1q2,s.s1q3,s.s1q4,s.s1q5,s.s1q6,s.s1q7,
			s.s2q1,s.s2q2,s.s2q3,s.s2q4,s.s2q5,s.s2q6,s.s2q6,s.s2q7,s.s2q8,s.s2q9,
			s.lower_d,s.lower_e,s.lower_f,
			s.upper_d,s.upper_m,s.upper_f,
			s.subjective_score,s.created_at
		from
			surveys s
			left join students student
				on s.student_id = student.id
			left join schools school
				on student.school_id = school.id
		where
			school.id = ?;`

	err = s.db.Select(&models, reportSQL, schoolId)
	if err != nil {
		return err
	}
	println(fmt.Sprintf("MODEL: %v", models))

	// Create a new CSV file
	f, err := os.Create(fmt.Sprintf("./tmp/%s.csv", models[0].SchoolName))
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the data to the file
	w := csv.NewWriter(f)
	var dataTemp [][]string

	var tmp = data.Questions

	dataTemp = append(dataTemp, []string{
		"student name",
		tmp[0], tmp[1], tmp[2], tmp[3], tmp[4], tmp[5], tmp[6],
		tmp[7], tmp[8], tmp[9], tmp[10], tmp[11], tmp[12], tmp[13], tmp[14], tmp[15],
		"lower_d", "lower_e", "lower_f",
		"upper_d", "upper_m", "upper_f",
		"subjective score", "date of survey",
	})

	for _, v := range models {
		dataTemp = append(dataTemp, []string{
			v.StudentName,
			v.S1Q1, v.S1Q2, v.S1Q3, v.S1Q4, v.S1Q5, v.S1Q6, v.S1Q7,
			v.S2Q1, v.S2Q2, v.S2Q3, v.S2Q4, v.S2Q5, v.S2Q6, v.S2Q7, v.S2Q8, v.S2Q9,
			v.LowerD, v.LowerE, v.LowerF,
			v.UpperD, v.UpperM, v.UpperF,
			v.SubjectiveScore, v.DateOfSurvey.Format("2006-01-02 15:04:05"),
		})
	}

	w.WriteAll(dataTemp)
	w.Flush()

	err = generateXLSXFromCSV(fmt.Sprintf("./tmp/%s.csv", models[0].SchoolName), fmt.Sprintf("./tmp/%s.xlsx", models[0].SchoolName), ",")

	err = zipSource("./tmp/", "surveyreports.zip")
	return err
}

func getReport(reportModel model.SurveyReport) (reportData bytes.Buffer, err error) {
	begin := time.Now()

	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	m.SetPageMargins(15, 15, 10)

	m.RegisterHeader(func() {
		// meh
	})

	m.RegisterFooter(func() {
		// meh
	})

	// REPORT TITLE
	m.Row(9, func() {
		m.Col(12, func() {
			m.Text("Laporan Survey IDCRA", props.Text{
				Size:  16,
				Top:   0,
				Style: consts.Bold,
				Align: consts.Center,
			})
		})
	})

	// REPORT IDENTITY
	m.Row(6, func() {
		m.Col(12, func() {
			m.Text("Identitas", props.Text{
				Size:  12,
				Top:   0,
				Style: consts.Bold,
				Align: consts.Center,
			})
		})
	})
	m.Row(6, func() {
		m.Col(3, func() {
			m.Text("Nama Siswa", props.Text{
				Top:   1,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
		m.Col(9, func() {
			m.Text(reportModel.StudentName, props.Text{
				Top:   1,
				Align: consts.Left,
			})
		})
	})
	m.Row(6, func() {
		m.Col(3, func() {
			m.Text("Nama Sekolah", props.Text{
				Top:   1,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
		m.Col(9, func() {
			m.Text(reportModel.SchoolName, props.Text{
				Top:   1,
				Align: consts.Left,
			})
		})
	})
	m.Row(6, func() {
		m.Col(3, func() {
			m.Text("Tanggal Survey", props.Text{
				Top:   1,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
		m.Col(9, func() {
			m.Text(reportModel.DateOfSurvey.Format("02 January 2006"), props.Text{
				Top:   1,
				Align: consts.Left,
			})
		})
	})

	// REPORT GRAPHS
	m.Row(12, func() {
		m.Col(12, func() {
			m.Text("Grafik Hasil Survey", props.Text{
				Size:  12,
				Top:   6,
				Style: consts.Bold,
				Align: consts.Center,
			})
		})
	})

	m.Row(45, func() {
		m.Col(4, func() {
			scaChart, err := getSCAPercentageChart(reportModel.SCAPercentage)
			if err == nil {
				m.Base64Image(scaChart, consts.Png)
			}
		})
		m.Col(8, func() {
			scaChart, err := getSCADMFChart(reportModel.DValue, reportModel.MValue, reportModel.FValue)
			if err == nil {
				m.Base64Image(scaChart, consts.Png)
			}
		})
	})

	// OPERATOR'S SUGGESTION
	m.Row(6, func() {
		m.Col(12, func() {
			m.Text("Operator's Suggestions", props.Text{
				Size:  12,
				Top:   0,
				Style: consts.Bold,
				Align: consts.Center,
			})
		})
	})

	m.Row(6, func() {
		m.Col(3, func() {
			m.Text("RECURRING", props.Text{
				Top:   1,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
		m.Col(9, func() {
			m.Text(reportModel.OperatorSuggestionRecurring, props.Text{
				Top:   1,
				Align: consts.Left,
			})
		})
	})

	m.Row(6, func() {
		m.Col(3, func() {
			m.Text("FLUORIDE", props.Text{
				Top:   1,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
		m.Col(9, func() {
			m.Text(reportModel.OperatorSuggestionFluoride, props.Text{
				Top:   1,
				Align: consts.Left,
			})
		})
	})

	m.Row(6, func() {
		m.Col(3, func() {
			m.Text("DIET", props.Text{
				Top:   1,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
		m.Col(9, func() {
			m.Text(reportModel.OperatorSuggestionDiet, props.Text{
				Top:   1,
				Align: consts.Left,
			})
		})
	})

	m.Row(6, func() {
		m.Col(3, func() {
			m.Text("SEALANT", props.Text{
				Top:   1,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
		m.Col(9, func() {
			m.Text(reportModel.OperatorSuggestionSealant, props.Text{
				Top:   1,
				Align: consts.Left,
			})
		})
	})

	m.Row(6, func() {
		m.Col(3, func() {
			m.Text("ART", props.Text{
				Top:   1,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
		m.Col(9, func() {
			m.Text(reportModel.OperatorSuggestionART, props.Text{
				Top:   1,
				Align: consts.Left,
			})
		})
	})

	// PARENT'S SUGGESTION
	m.Row(12, func() {
		m.Col(12, func() {
			m.Text("Parent's Suggestions", props.Text{
				Size:  12,
				Top:   12,
				Style: consts.Bold,
				Align: consts.Center,
			})
		})
	})

	m.Row(10, func() {
		m.Col(12, func() {
			m.Text("Reminder", props.Text{
				Size:  10,
				Top:   6,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
	})

	for _, reminder := range reportModel.ParentReminder {
		m.Row(6, func() {
			m.Col(12, func() {
				m.Text("- "+reminder, props.Text{
					Top:   1,
					Align: consts.Left,
				})
			})
		})
	}

	m.Row(10, func() {
		m.Col(12, func() {
			m.Text("Guidance", props.Text{
				Size:  10,
				Top:   6,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
	})

	for _, guidance := range reportModel.ParentGuidance {
		m.Row(6, func() {
			m.Col(12, func() {
				m.Text("- "+guidance, props.Text{
					Top:   1,
					Align: consts.Left,
				})
			})
		})
	}

	m.Row(10, func() {
		m.Col(12, func() {
			m.Text("Supervision", props.Text{
				Size:  10,
				Top:   6,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
	})

	for _, supervision := range reportModel.ParentSupervision {
		m.Row(6, func() {
			m.Col(12, func() {
				m.Text("- "+supervision, props.Text{
					Top:   1,
					Align: consts.Left,
				})
			})
		})
	}

	// TEACHER'S SUGGESTION
	m.Row(12, func() {
		m.Col(12, func() {
			m.Text("Teacher's Suggestions", props.Text{
				Size:  12,
				Top:   12,
				Style: consts.Bold,
				Align: consts.Center,
			})
		})
	})

	m.Row(10, func() {
		m.Col(12, func() {
			m.Text("Reminder", props.Text{
				Size:  10,
				Top:   6,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
	})

	for _, reminder := range reportModel.TeacherReminder {
		m.Row(6, func() {
			m.Col(12, func() {
				m.Text("- "+reminder, props.Text{
					Top:   1,
					Align: consts.Left,
				})
			})
		})
	}

	m.Row(10, func() {
		m.Col(12, func() {
			m.Text("Guidance", props.Text{
				Size:  10,
				Top:   6,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
	})

	for _, guidance := range reportModel.TeacherGuidance {
		m.Row(6, func() {
			m.Col(12, func() {
				m.Text("- "+guidance, props.Text{
					Top:   1,
					Align: consts.Left,
				})
			})
		})
	}

	end := time.Now()
	fmt.Println(end.Sub(begin))
	return m.Output()
}

func getSCAPercentageChart(riskPercentage float64) (chartAsBase64 string, err error) {
	graph := chart.BarChart{
		Title: "Subjective Score",
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		YAxis: chart.YAxis{
			Name: "The YAxis",
			Style: chart.Style{
				Hidden:      false,
				StrokeColor: drawing.ColorBlack,
				StrokeWidth: 1,
			},
			Range: &chart.ContinuousRange{
				Min: 0.0,
				Max: 100.0,
			},
		},
		XAxis: chart.Style{
			Hidden:      false,
			StrokeColor: drawing.ColorBlack,
			StrokeWidth: 1,
		},
		Height:   250,
		Width:    250,
		BarWidth: 70,
		Bars: []chart.Value{
			{
				Value: riskPercentage,
				Label: "Risk (%)",
				Style: chart.Style{
					FillColor:   chart.ColorOrange,
					StrokeColor: chart.ColorOrange,
				},
			},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err = graph.Render(chart.PNG, buffer)
	chartAsBase64 = base64.StdEncoding.EncodeToString(buffer.Bytes())
	return
}

func getSCADMFChart(D, M, F float64) (chartAsBase64 string, err error) {
	highestValue := D

	if M > highestValue {
		highestValue = M
	}

	if F > highestValue {
		highestValue = F
	}

	if highestValue == 0 {
		highestValue = 1
	}

	graph := chart.BarChart{
		Title: "DMF",
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		YAxis: chart.YAxis{
			Name: "The YAxis",
			Style: chart.Style{
				Hidden:      false,
				StrokeColor: drawing.ColorBlack,
				StrokeWidth: 1,
			},
			Range: &chart.ContinuousRange{
				Min: 0.0,
				Max: highestValue,
			},
		},
		XAxis: chart.Style{
			Hidden:      false,
			StrokeColor: drawing.ColorBlack,
			StrokeWidth: 1,
		},
		Height:   250,
		Width:    500,
		BarWidth: 70,
		Bars: []chart.Value{
			{
				Value: D,
				Label: "D",
				Style: chart.Style{
					FillColor:   chart.ColorRed,
					StrokeColor: chart.ColorRed,
				},
			},
			{
				Value: M,
				Label: "M",
				Style: chart.Style{
					FillColor:   chart.ColorRed,
					StrokeColor: chart.ColorRed,
				},
			},
			{
				Value: F,
				Label: "F",
				Style: chart.Style{
					FillColor:   chart.ColorRed,
					StrokeColor: chart.ColorRed,
				},
			},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err = graph.Render(chart.PNG, buffer)
	chartAsBase64 = base64.StdEncoding.EncodeToString(buffer.Bytes())
	return
}

func zipSource(source, target string) error {
	// 1. Create a ZIP file and zip.Writer
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	// 2. Go through all the files of the source
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 3. Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set compression
		header.Method = zip.Deflate

		// 4. Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// 5. Create writer for the file header and save content of the file
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}

func generateXLSXFromCSV(csvPath string, XLSXPath string, delimiter string) error {
	csvFile, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer csvFile.Close()
	reader := csv.NewReader(csvFile)
	if len(delimiter) > 0 {
		reader.Comma = rune(delimiter[0])
	} else {
		reader.Comma = rune(',')
	}
	xlsxFile := xlsx.NewFile()
	sheet, err := xlsxFile.AddSheet("sheet1")
	if err != nil {
		return err
	}
	fields, err := reader.Read()
	for err == nil {
		row := sheet.AddRow()
		for _, field := range fields {
			cell := row.AddCell()
			cell.Value = field
		}
		fields, err = reader.Read()
	}
	if err != nil {
		fmt.Printf(err.Error())
	}
	return xlsxFile.Save(XLSXPath)
}
