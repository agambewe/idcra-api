package model

import "time"

type QuestionData struct {
	Questions []string `json:"questions"`
}

type SurveyReportCSV struct {
	StudentName     string    `db:"studentname"`
	SchoolName      string    `db:"schoolname"`
	S1Q1            string    `db:"s1q1"`
	S1Q2            string    `db:"s1q2"`
	S1Q3            string    `db:"s1q3"`
	S1Q4            string    `db:"s1q4"`
	S1Q5            string    `db:"s1q5"`
	S1Q6            string    `db:"s1q6"`
	S1Q7            string    `db:"s1q7"`
	S2Q1            string    `db:"s2q1"`
	S2Q2            string    `db:"s2q2"`
	S2Q3            string    `db:"s2q3"`
	S2Q4            string    `db:"s2q4"`
	S2Q5            string    `db:"s2q5"`
	S2Q6            string    `db:"s2q6"`
	S2Q7            string    `db:"s2q7"`
	S2Q8            string    `db:"s2q8"`
	S2Q9            string    `db:"s2q9"`
	LowerD          string    `db:"lower_d"`
	LowerE          string    `db:"lower_e"`
	LowerF          string    `db:"lower_f"`
	UpperD          string    `db:"upper_d"`
	UpperM          string    `db:"upper_m"`
	UpperF          string    `db:"upper_f"`
	SubjectiveScore string    `db:"subjective_score"`
	DateOfSurvey    time.Time `db:"created_at"`
}

type SurveyReport struct {

	// Preset values
	StudentName   string    `db:"studentname"`
	SchoolName    string    `db:"schoolname"`
	DateOfSurvey  time.Time `db:"dateofsurvey"`
	SCAPercentage float64   `db:"scapercentage"`
	DValue        float64   `db:"dvalue"`
	MValue        float64   `db:"mvalue"`
	FValue        float64   `db:"fvalue"`

	// Calculated values
	RiskProfile                 string
	OperatorSuggestionRecurring string
	OperatorSuggestionFluoride  string
	OperatorSuggestionDiet      string
	OperatorSuggestionSealant   string
	OperatorSuggestionART       string
	ParentReminder              []string
	ParentGuidance              []string
	ParentSupervision           []string
	TeacherReminder             []string
	TeacherGuidance             []string
}

func (sr *SurveyReport) Setup() {
	sr.RiskProfile = "low"

	if sr.SCAPercentage > 66 {
		sr.RiskProfile = "high"
	} else if sr.SCAPercentage > 33 && sr.SCAPercentage <= 66 {
		sr.RiskProfile = "medium"
	}

	switch sr.RiskProfile {

	case "low":
		sr.ParentReminder = []string{
			"Orang tua mengingatkan agar kontrol ke dokter gigi setiap 6 bulan sekali",
		}
		sr.ParentGuidance = []string{
			"Orang tua mengajarkan cara menyikat gigi yang benar",
			"Orang tua mengingatkan agar menyikat gigi 2x sehari dengan pasta gigi ber fluoride",
		}
		sr.ParentSupervision = []string{
			"Orang tua memberikan pengawasan terhadap makanan manis dan lengket yang dikonsumsi sehari - hari",
		}
		sr.TeacherReminder = []string{
			"Guru mengingatkan agar kontrol ke dokter gigi setiap 6 bulan sekali",
		}
		sr.TeacherGuidance = []string{
			"Guru mengajarkan cara menyikat gigi yang benar",
			"Guru mengingatkan agar menyikat gigi 2x sehari dengan pasta gigi ber fluoride",
		}
		sr.OperatorSuggestionRecurring = "setiap 6-12 bulan"
		sr.OperatorSuggestionFluoride = "pasta gigi 2x sehari"
		sr.OperatorSuggestionDiet = "pemeliharaan asupan diet"
		sr.OperatorSuggestionSealant = "fissure sealant dilakukan jika diperlukan"
		sr.OperatorSuggestionART = "pengawasan karies baru"

	case "medium":
		sr.ParentReminder = []string{
			"Orang tua mengingatkan agar kontrol ke dokter gigi setiap 4-6 bulan sekali",
		}
		sr.ParentGuidance = []string{
			"Orang tua mengajarkan cara menyikat gigi yang benar",
			"Orang tua mengingatkan agar menyikat gigi 2x sehari dengan pasta gigi ber fluoride",
			"Orang tua mengingatkan agar dilakukan perawatan topical aplikasi fluoride",
		}
		sr.ParentSupervision = []string{
			"Orang tua melakukan diet makanan manis dan lengket yang dikonsumsi sehari- hari",
		}
		sr.TeacherReminder = []string{
			"Guru mengingatkan agar kontrol ke dokter gigi setiap 4-6 bulan sekali",
		}
		sr.TeacherGuidance = []string{
			"Guru mengajarkan cara menyikat gigi yang benar",
			"Guru mengingatkan agar menyikat gigi 2x sehari dengan pasta gigi ber fluoride",
			"Guru mengingatkan agar dilakukan perawatan topical aplikasi fluoride",
		}
		sr.OperatorSuggestionRecurring = "setiap 4-6 bulan"
		sr.OperatorSuggestionFluoride = "pasta gigi 2x sehari + Topikal aplikasi"
		sr.OperatorSuggestionDiet = "diet dengan pengawasan"
		sr.OperatorSuggestionSealant = "fissure sealant dilakukan jika diperlukan"
		sr.OperatorSuggestionART = "pengawasan karies baru + restorasi dari kavitas baru"

	case "high":
		sr.ParentReminder = []string{
			"Orang tua mengingatkan agar kontrol ke dokter gigi setiap 3-4 bulan sekali",
		}
		sr.ParentGuidance = []string{
			"Orang tua mengajarkan cara menyikat gigi yang benar",
			"Orang tua mengingatkan agar menyikat gigi 2x sehari dengan pasta gigi ber fluoride",
			"Orang tua mengingatkan agar dilakukan perawatan topical aplikasi fluoride",
		}
		sr.ParentSupervision = []string{
			"Orang tua melakukan diet makanan manis dan lengket yang dikonsumsi sehari- hari",
			"Orang tua mengganti konsumsi permen yang manis dengan permen xylitol",
		}
		sr.TeacherReminder = []string{
			"Guru mengingatkan agar kontrol ke dokter gigi setiap 3-4 bulan sekali",
		}
		sr.TeacherGuidance = []string{
			"Guru mengajarkan cara menyikat gigi yang benar",
			"Guru mengingatkan agar menyikat gigi 2x sehari dengan pasta gigi ber fluoride",
			"Guru mengingatkan agar dilakukan perawatan topical aplikasi fluoride",
		}
		sr.OperatorSuggestionRecurring = "setiap 3-4 bulan"
		sr.OperatorSuggestionFluoride = "topikal aplikasi + pasta gigi 2x sehari"
		sr.OperatorSuggestionDiet = "diet dengan pengawasan + xylitol"
		sr.OperatorSuggestionSealant = "direkomendasikan fissure sealant"
		sr.OperatorSuggestionSealant = "pengawasan karies baru + restorasi dari kavitas baru"

	}
}
