package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	gcontext "github.com/kerti/idcra-api/context"
	h "github.com/kerti/idcra-api/handler"
	"github.com/kerti/idcra-api/resolver"
	"github.com/kerti/idcra-api/schema"
	"github.com/kerti/idcra-api/service"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/kerti/idcra-api/loader"
	"golang.org/x/net/context"
)

func main() {
	config := gcontext.LoadConfig(".")

	db, err := gcontext.OpenDB(config)
	if err != nil {
		log.Fatalf("Unable to connect to db: %s \n", err)
	}
	ctx := context.Background()
	log := service.NewLogger(config)
	roleService := service.NewRoleService(db, log)
	authService := service.NewAuthService(config, log)
	studentService := service.NewStudentService(db, log)
	schoolService := service.NewSchoolService(db, log)
	diagnosisAndActionService := service.NewDiagnosisAndActionService(db, log)
	caseService := service.NewCaseService(db, log)
	surveyService := service.NewSurveyService(db, caseService, log)
	reportService := service.NewReportService(db, log)
	userService := service.NewUserService(db, roleService, studentService, log)

	ctx = context.WithValue(ctx, "config", config)
	ctx = context.WithValue(ctx, "log", log)
	ctx = context.WithValue(ctx, "roleService", roleService)
	ctx = context.WithValue(ctx, "userService", userService)
	ctx = context.WithValue(ctx, "authService", authService)

	ctx = context.WithValue(ctx, "studentService", studentService)
	ctx = context.WithValue(ctx, "schoolService", schoolService)
	ctx = context.WithValue(ctx, "diagnosisAndActionService", diagnosisAndActionService)
	ctx = context.WithValue(ctx, "caseService", caseService)
	ctx = context.WithValue(ctx, "surveyService", surveyService)
	ctx = context.WithValue(ctx, "reportService", reportService)

	graphqlSchema := graphql.MustParseSchema(schema.GetRootSchema(), &resolver.Resolver{})

	http.Handle("/login", h.AddContext(ctx, h.Login()))

	loggerHandler := &h.LoggerHandler{DebugMode: config.DebugMode}
	http.Handle("/query", h.AddContext(ctx, loggerHandler.Logging(h.Authenticate(&h.GraphQL{Schema: graphqlSchema, Loaders: loader.NewLoaderCollection()}))))

	http.Handle("/reports/surveys/", h.AddContext(ctx, loggerHandler.Logging(h.Authenticate(h.SurveyReport()))))
	http.Handle("/reports/school/", h.AddContext(ctx, loggerHandler.Logging(h.Authenticate(h.SchoolReport()))))

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "graphiql.html")
	}))

	port := flag.String("port", "3001", "a port")
	flag.Parse()

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *port), nil))
}
