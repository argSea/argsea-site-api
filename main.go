package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/argSea/portfolio_blog_api/argHex/in_adapter"
	"github.com/argSea/portfolio_blog_api/argHex/out_adapter"
	"github.com/argSea/portfolio_blog_api/argHex/service"
	"github.com/argSea/portfolio_blog_api/argHex/stores"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func init() {
	print("Initializing argSea API\n")

	// look for --config in cli args
	config := ""
	log_file := ""
	// for loop with index
	for index, element := range os.Args {
		if "--config" == element {
			config = os.Args[index+1]
		}

		if "--log" == element {
			log_file = os.Args[index+1]
		}
	}

	if "" == log_file {
		log.Fatal("No log file found")
		os.Exit(1)
	}

	print("Using config file: " + config + "\n")
	print("Using log file: " + log_file + "\n")

	//logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log_file_fh, log_file_err := os.OpenFile(log_file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	if nil != log_file_err {
		log.Fatal(log_file_err)
	}

	log.SetOutput(log_file_fh)

	if "" != config {
		viper.SetConfigFile(config)
	} else {
		// die if no config file
		log.Fatal("No config file found")
		os.Exit(1)
	}

	// read config
	err := viper.ReadInConfig()

	if nil != err {
		log.Fatal(err)
		os.Exit(1)
	}
}

func main() {
	//signal to kill and print final info
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Shutting down argSea API")
		fmt.Println("Shutting down argSea API")
		os.Exit(0)
	}()

	//mux
	router := mux.NewRouter()
	router.Use(baseMiddleWare)
	// router.StrictSlash(true)

	//Cache credentials
	mHost := viper.GetString("mongo.host") + ":" + viper.GetString("mongo.port")
	mUser := viper.GetString("mongo.user")
	mPass := viper.GetString("mongo.pass")
	mDB := viper.GetString("mongo.dbName")
	jSecret := []byte(viper.GetString("jwt.secret"))

	//setup mongo
	mongo_db, mongo_err := stores.NewMongoStore(mUser, mPass, mHost, mDB)

	defer mongo_db.Client.Disconnect(context.Background())

	if nil != mongo_err {
		fmt.Fprintf(os.Stderr, "error: %v\n", mongo_err)
		log.Fatal(mongo_err)
		os.Exit(1)
	}

	user_table := "users"
	projectTable := "projects"
	resumeTable := "resume"
	skillTable := "skills"
	authDB := 13

	// setup redis
	redis_host := viper.GetString("redis.host")
	redis_port := viper.GetString("redis.port")
	redis_pass := viper.GetString("redis.pass")
	redis_user := viper.GetString("redis.user")

	redis_store, redis_err := stores.NewRedisStore(redis_host, redis_port, redis_user, redis_pass, authDB)

	if nil != redis_err {
		fmt.Fprintf(os.Stderr, "error: %v\n", redis_err)
		log.Fatal(redis_err)
		os.Exit(1)
	}

	// media for users
	save_path := viper.GetString("media.images.save_path")
	web_path := viper.GetString("media.images.web_path")
	mediaWebstoreAdapter := out_adapter.NewMediaWebstoreAdapter(save_path, web_path)
	mediaService := service.NewMediaService(mediaWebstoreAdapter)

	//user
	userRouter := router.PathPrefix("/1/user").Subrouter()
	projRouter := router.PathPrefix("/1/project").Subrouter()
	resumeRouter := router.PathPrefix("/1/resume").Subrouter()
	authRouter := router.PathPrefix("/1/auth").Subrouter()
	skillRouter := router.PathPrefix("/1/skill").Subrouter()

	//resume
	log.Println("Initializing resume")
	// resumeDrivenAdapter := out_adapter.NewResumeFakeOutAdapter()
	resumeMordor := stores.NewMordor(mongo_db.DB.Collection(resumeTable), context.Background())
	resumeMongoAdapter := out_adapter.NewResumeMongoAdapter(resumeMordor)
	resumeService := service.NewResumeCRUDService(resumeMongoAdapter)
	in_adapter.NewResumeMuxAdapter(resumeService, resumeRouter)

	//project
	log.Println("Initializing project")
	// projectDrivenAdapter := out_adapter.NewProjectFakeOutAdapter()
	projectMordor := stores.NewMordor(mongo_db.DB.Collection(projectTable), context.Background())
	projectMongoAdapter := out_adapter.NewProjectMongoAdapter(projectMordor)
	projectService := service.NewProjectCRUDService(projectMongoAdapter)
	in_adapter.NewProjectMuxAdapter(projectService, projRouter, mediaService)

	//User
	log.Println("Initializing user")
	// userDrivenAdapter := out_adapter.NewUserFakeOutAdapter()
	userMordor := stores.NewMordor(mongo_db.DB.Collection(user_table), context.Background())
	userMongoAdapter := out_adapter.NewUserMongoAdapter(userMordor)
	userService := service.NewUserCRUDService(userMongoAdapter)
	userResumeService := service.NewUserResumeService(resumeMongoAdapter)
	userProjectService := service.NewUserProjectService(projectMongoAdapter)
	in_adapter.NewUserMuxAdapter(userService, userResumeService, userProjectService, mediaService, userRouter)

	//Auth
	log.Println("Initializing auth")
	authRivia := stores.NewRivia(redis_store, authDB)
	authRedisAdapter := out_adapter.NewAuthRedisAdapter(authRivia)
	userAuthService := service.NewSessionAuthService(authRedisAdapter)
	userLoginService := service.NewUserLoginService(userMongoAdapter)
	// userJWTService := service.NewJWTAuthService(jSecret)
	in_adapter.NewAuthMuxAdapter(userAuthService, userLoginService, jSecret, authRouter)

	//Skills
	log.Println("Initializing skills")
	skillMordor := stores.NewMordor(mongo_db.DB.Collection(skillTable), context.Background())
	skillMongoAdapter := out_adapter.NewSkillMongoAdapter(skillMordor)
	skillService := service.NewSkillCRUDService(skillMongoAdapter)
	in_adapter.NewSkillMuxAdapter(skillService, skillRouter)

	// echo back origins
	origins := handlers.AllowedOrigins([]string{"https://argsea.com", "https://www.argsea.com", "https://argsea.dev", "https://www.argsea.dev", "http://127.0.0.1:5173"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Content-Range", "range"})
	exposedHeaders := handlers.ExposedHeaders([]string{"Content-Range", "X-Total-Count"})
	credential := handlers.AllowCredentials()

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         ":8181",
		Handler:      handlers.CORS(origins, methods, headers, exposedHeaders, credential)(router),
	}

	err := srv.ListenAndServe()

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}

func baseMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		// if GET request, allow all origins
		// if r.Method == "GET" {
		// 	w.Header().Set("Access-Control-Allow-Origin", "*")
		// } else {
		// 	allowed_origins := []string{"https://argsea.com", "https://www.argsea.com", "https://argsea.dev", "https://www.argsea.dev"}
		// 	// get origin header
		// 	origin := r.Header.Get("Origin")

		// 	origin_check := false
		// 	// check if origin is in allowed origins
		// 	for _, allowed_origin := range allowed_origins {
		// 		if origin == allowed_origin {
		// 			origin_check = true
		// 			break
		// 		}
		// 	}

		// 	// if origin is not in allowed origins, set to first allowed origin
		// 	if !origin_check {
		// 		origin = allowed_origins[0]
		// 	}

		// 	// set allowed origins header to origin
		// 	w.Header().Set("Access-Control-Allow-Origin", origin)
		// }

		// handle preflight
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Content-Range")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.WriteHeader(http.StatusOK)
			return
		}

		fmt.Println(r.URL)
		fmt.Println(r.Method)

		next.ServeHTTP(w, r)
	})
}
