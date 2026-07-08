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

	"github.com/argSea/argsea-site-api/argHex/in_adapter"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
	"github.com/argSea/argsea-site-api/argHex/stores"
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
	mAuthDB := viper.GetString("mongo.authenticationDatabase")

	// authSource falls back to the target database when not set explicitly
	if "" == mAuthDB {
		mAuthDB = mDB
	}

	jSecret := []byte(viper.GetString("jwt.secret"))

	// JWT auth cannot run without a signing secret; fail fast with a pointer to the fix
	if 0 == len(jSecret) {
		missing_secret := "jwt.secret is missing or empty in the config file; add it (see config.example.json)"
		// log goes to the log file; also surface it on the console. log.Fatal exits.
		fmt.Fprintf(os.Stderr, "error: %s\n", missing_secret)
		log.Fatal(missing_secret)
	}

	//setup mongo
	mongo_db, mongo_err := stores.NewMongoStore(mUser, mPass, mHost, mDB, mAuthDB)

	defer mongo_db.Client.Disconnect(context.Background())

	if nil != mongo_err {
		fmt.Fprintf(os.Stderr, "error: %v\n", mongo_err)
		log.Fatal(mongo_err)
		os.Exit(1)
	}

	userTable := "users"
	projectTable := "projects"
	noteTable := "notes"
	hobbyTable := "hobbies"
	siteCopyTable := "siteCopy"
	suggestionTable := "suggestions"
	activityTable := "activity"
	revisionTable := "revisions"
	lanternTable := "lantern"
	mediaTable := "media"
	catDesignTable := "catDesigns"
	doodleTable := "doodles"

	// routers
	userRouter := router.PathPrefix("/1/user").Subrouter()
	projRouter := router.PathPrefix("/1/project").Subrouter()
	noteRouter := router.PathPrefix("/1/note").Subrouter()
	hobbyRouter := router.PathPrefix("/1/hobby").Subrouter()
	copyRouter := router.PathPrefix("/1/copy").Subrouter()
	suggestionRouter := router.PathPrefix("/1/suggestion").Subrouter()
	activityRouter := router.PathPrefix("/1/activity").Subrouter()
	authRouter := router.PathPrefix("/1/auth").Subrouter()
	mediaRouter := router.PathPrefix("/1/media").Subrouter()
	figureheadRouter := router.PathPrefix("/1/figurehead").Subrouter()
	doodleRouter := router.PathPrefix("/1/doodle").Subrouter()

	// the session cookie's domain is deploy-specific (prod vs a local vhost);
	// configurable, with the historical hardcoded value as the default
	cookieDomain := viper.GetString("auth.cookie_domain")

	if "" == cookieDomain {
		cookieDomain = "argsea.com"
	}

	// in-process auth: one JWT validator + one cookie/bearer extractor shared
	// by every adapter (no HTTP round-trips to a validate endpoint)
	userAuthService := service.NewJWTAuthService(jSecret)
	webAuth := in_adapter.NewWebAuth(userAuthService, jSecret, cookieDomain)

	// shared history + ship's log: projects and notes snapshot into revisions,
	// every content mutation records an activity entry
	log.Println("Initializing revisions and activity log")
	revisionMordor := stores.NewMordor(mongo_db.DB.Collection(revisionTable), context.Background())
	revisionService := service.NewRevisionService(out_adapter.NewRevisionMongoAdapter(revisionMordor))
	activityMordor := stores.NewMordor(mongo_db.DB.Collection(activityTable), context.Background())
	activityService := service.NewActivityService(out_adapter.NewActivityMongoAdapter(activityMordor))
	in_adapter.NewActivityMuxAdapter(activityService, webAuth, activityRouter)

	// projects (postcards)
	log.Println("Initializing project")
	projectMordor := stores.NewMordor(mongo_db.DB.Collection(projectTable), context.Background())
	projectService := service.NewProjectCRUDService(out_adapter.NewProjectMongoAdapter(projectMordor), revisionService, activityService)
	in_adapter.NewProjectMuxAdapter(projectService, webAuth, projRouter)

	// notes (writing desk)
	log.Println("Initializing note")
	noteMordor := stores.NewMordor(mongo_db.DB.Collection(noteTable), context.Background())
	noteService := service.NewNoteCRUDService(out_adapter.NewNoteMongoAdapter(noteMordor), revisionService, activityService)
	in_adapter.NewNoteMuxAdapter(noteService, webAuth, noteRouter)

	// hobbies (graveyard)
	log.Println("Initializing hobby")
	hobbyMordor := stores.NewMordor(mongo_db.DB.Collection(hobbyTable), context.Background())
	hobbyService := service.NewHobbyCRUDService(out_adapter.NewHobbyMongoAdapter(hobbyMordor), activityService)
	in_adapter.NewHobbyMuxAdapter(hobbyService, webAuth, hobbyRouter)

	// site copy (signal flags), singleton
	log.Println("Initializing site copy")
	siteCopyMordor := stores.NewMordor(mongo_db.DB.Collection(siteCopyTable), context.Background())
	siteCopyService := service.NewSiteCopyService(out_adapter.NewSiteCopyMongoAdapter(siteCopyMordor), activityService)
	in_adapter.NewSiteCopyMuxAdapter(siteCopyService, webAuth, copyRouter)

	// suggestion pool (the "next: ???" chips)
	log.Println("Initializing suggestions")
	suggestionMordor := stores.NewMordor(mongo_db.DB.Collection(suggestionTable), context.Background())
	suggestionService := service.NewSuggestionService(out_adapter.NewSuggestionMongoAdapter(suggestionMordor), activityService)
	in_adapter.NewSuggestionMuxAdapter(suggestionService, webAuth, suggestionRouter)

	// media (the darkroom): metadata in mongo, files on disk behind the
	// webstore adapter; the same service still carries the legacy base64 path
	// the user adapter uploads through
	log.Println("Initializing media")
	save_path := viper.GetString("media.images.save_path")
	web_path := viper.GetString("media.images.web_path")
	mediaMordor := stores.NewMordor(mongo_db.DB.Collection(mediaTable), context.Background())
	mediaService := service.NewMediaService(out_adapter.NewMediaWebstoreAdapter(save_path, web_path), out_adapter.NewMediaMetaMongoAdapter(mediaMordor), activityService)
	in_adapter.NewMediaMuxAdapter(mediaService, webAuth, mediaRouter)

	// the figurehead shop (cat designs): the seed plants the shipped v1 cats
	// into an empty collection so "go back to v1" is always possible; on every
	// later boot it is a no-op
	log.Println("Initializing figurehead")
	catDesignMordor := stores.NewMordor(mongo_db.DB.Collection(catDesignTable), context.Background())
	figureheadService := service.NewFigureheadService(out_adapter.NewCatDesignMongoAdapter(catDesignMordor), activityService)

	if err := figureheadService.Seed(); nil != err {
		// the API stays up without the seed (the shop endpoints still work);
		// the next boot retries
		log.Printf("figurehead seed failed: %v\n", err)
	}

	in_adapter.NewFigureheadMuxAdapter(figureheadService, webAuth, figureheadRouter)

	// doodles (marginalia sketches for the Keeper's Journal): structured
	// shapes only, no publish/seed/pose lifecycle
	log.Println("Initializing doodle")
	doodleMordor := stores.NewMordor(mongo_db.DB.Collection(doodleTable), context.Background())
	doodleService := service.NewDoodleService(out_adapter.NewDoodleMongoAdapter(doodleMordor), activityService)
	in_adapter.NewDoodleMuxAdapter(doodleService, webAuth, doodleRouter)

	// users: kept (auth depends on it)
	log.Println("Initializing user")
	userMordor := stores.NewMordor(mongo_db.DB.Collection(userTable), context.Background())
	userMongoAdapter := out_adapter.NewUserMongoAdapter(userMordor)
	userService := service.NewUserCRUDService(userMongoAdapter)
	in_adapter.NewUserMuxAdapter(userService, mediaService, webAuth, userRouter)

	// auth: kept; sessions are issued through the same shared WebAuth store
	log.Println("Initializing auth")
	userLoginService := service.NewUserLoginService(userMongoAdapter)
	in_adapter.NewAuthMuxAdapter(userAuthService, userLoginService, webAuth, authRouter)

	// the lantern: deploy-on-hoist. The config section IS the feature flag:
	// no lantern section, no routes mounted, nothing else changes.
	if viper.IsSet("lantern") {
		log.Println("Initializing lantern")

		lanternTimeout := viper.GetInt("lantern.timeout_seconds")

		if 0 == lanternTimeout {
			lanternTimeout = 600 // ten minutes is generous for a static build
		}

		lanternKeep := viper.GetInt("lantern.keep")

		if 0 == lanternKeep {
			lanternKeep = 2
		}

		lanternConfig := service.LanternConfig{
			SiteDir:  viper.GetString("lantern.site_dir"),
			BuildCmd: viper.GetStringSlice("lantern.build_cmd"),
			DistDir:  viper.GetString("lantern.dist_dir"),
			Keep:     lanternKeep,
			Timeout:  time.Duration(lanternTimeout) * time.Second,
			// env is an array of KEY=VALUE strings, NOT a JSON object: viper
			// lowercases nested map keys on load, which would silently turn
			// ARGSEA_API_URL into argsea_api_url. Slice values pass through intact.
			Env: viper.GetStringSlice("lantern.env"),
		}

		lanternMordor := stores.NewMordor(mongo_db.DB.Collection(lanternTable), context.Background())
		lanternService := service.NewLanternService(
			lanternConfig,
			out_adapter.NewLanternExecAdapter(),
			out_adapter.NewLanternFSReleaseAdapter(viper.GetString("lantern.releases_dir"), viper.GetString("lantern.live_link")),
			out_adapter.NewLanternMongoAdapter(lanternMordor),
			activityService,
		)
		lanternRouter := router.PathPrefix("/1/lantern").Subrouter()
		in_adapter.NewLanternMuxAdapter(lanternService, webAuth, lanternRouter)
	}

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
