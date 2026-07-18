package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	caselogTable := "caselogs"
	blocksetTable := "blocksets"
	noteTable := "notes"
	hobbyTable := "hobbies"
	siteCopyTable := "siteCopy"
	watchTable := "watch"
	suggestionTable := "suggestions"
	activityTable := "activity"
	revisionTable := "revisions"
	lanternTable := "lantern"
	mediaTable := "media"
	catDesignTable := "catDesigns"
	doodleTable := "doodles"
	carvingTable := "carvings"
	sightingTable := "sightings"
	loginLockTable := "login_locks"

	// routers
	userRouter := router.PathPrefix("/1/user").Subrouter()
	projRouter := router.PathPrefix("/1/project").Subrouter()
	caselogRouter := router.PathPrefix("/1/caselog").Subrouter()
	blocksetRouter := router.PathPrefix("/1/blockset").Subrouter()
	noteRouter := router.PathPrefix("/1/note").Subrouter()
	hobbyRouter := router.PathPrefix("/1/hobby").Subrouter()
	copyRouter := router.PathPrefix("/1/copy").Subrouter()
	watchRouter := router.PathPrefix("/1/watch").Subrouter()
	suggestionRouter := router.PathPrefix("/1/suggestion").Subrouter()
	activityRouter := router.PathPrefix("/1/activity").Subrouter()
	authRouter := router.PathPrefix("/1/auth").Subrouter()
	mediaRouter := router.PathPrefix("/1/media").Subrouter()
	figureheadRouter := router.PathPrefix("/1/figurehead").Subrouter()
	doodleRouter := router.PathPrefix("/1/doodle").Subrouter()
	carvingRouter := router.PathPrefix("/1/carving").Subrouter()
	sightingRouter := router.PathPrefix("/1/sighting").Subrouter()

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

	// shared history + keeper's log: projects and notes snapshot into revisions,
	// every content mutation records an activity entry
	log.Println("Initializing revisions and activity log")
	revisionMordor := stores.NewMordor(mongo_db.DB.Collection(revisionTable), context.Background())
	revisionService := service.NewRevisionService(out_adapter.NewRevisionMongoAdapter(revisionMordor))
	activityMordor := stores.NewMordor(mongo_db.DB.Collection(activityTable), context.Background())
	activityService := service.NewActivityService(out_adapter.NewActivityMongoAdapter(activityMordor))
	in_adapter.NewActivityMuxAdapter(activityService, webAuth, activityRouter)

	// notes (writing desk), wired ahead of projects: the rack's noteIds tie
	// validation needs a read-only handle to the notes collection
	log.Println("Initializing note")
	noteMordor := stores.NewMordor(mongo_db.DB.Collection(noteTable), context.Background())
	noteRepo := out_adapter.NewNoteMongoAdapter(noteMordor)
	noteService := service.NewNoteCRUDService(noteRepo, revisionService, activityService)
	in_adapter.NewNoteMuxAdapter(noteService, webAuth, noteRouter)

	// projects (postcards)
	log.Println("Initializing project")
	projectMordor := stores.NewMordor(mongo_db.DB.Collection(projectTable), context.Background())
	projectRepo := out_adapter.NewProjectMongoAdapter(projectMordor)
	projectService := service.NewProjectCRUDService(projectRepo, noteRepo, revisionService, activityService)
	in_adapter.NewProjectMuxAdapter(projectService, webAuth, projRouter)

	// case studies (the full log): each project's long-form story as its own
	// document. A one-time boot migration lifts any legacy project.caseStudy
	// into a published caselog before the routes mount; it reads the dormant
	// field straight off the projects collection and is idempotent, so every
	// later boot moves nothing. Block sets seed their one header template.
	log.Println("Initializing caselog")
	caselogMordor := stores.NewMordor(mongo_db.DB.Collection(caselogTable), context.Background())
	caselogRepo := out_adapter.NewCaseLogMongoAdapter(caselogMordor)
	caselogService := service.NewCaseLogCRUDService(caselogRepo, projectRepo, revisionService, activityService)

	caselogMigration := service.NewCaseLogMigration(caselogRepo, out_adapter.NewCaseStudySourceMongoAdapter(projectMordor))

	if migrated, err := caselogMigration.Run(); nil != err {
		// the API stays up if the migration fails; a later boot retries the
		// projects that did not land a log yet
		log.Printf("caselog migration failed: %v\n", err)
	} else {
		log.Printf("caselog migration: %d log(s) migrated\n", migrated)
	}

	in_adapter.NewCaseLogMuxAdapter(caselogService, webAuth, caselogRouter)

	log.Println("Initializing blockset")
	blocksetMordor := stores.NewMordor(mongo_db.DB.Collection(blocksetTable), context.Background())
	blocksetService := service.NewBlockSetService(out_adapter.NewBlockSetMongoAdapter(blocksetMordor), activityService)

	if err := blocksetService.Seed(); nil != err {
		// the API stays up without the seed (the endpoints still work); the next
		// boot retries
		log.Printf("blockset seed failed: %v\n", err)
	}

	in_adapter.NewBlockSetMuxAdapter(blocksetService, webAuth, blocksetRouter)

	// hobbies (the ship's log): a one-time boot migration lifts any old-shape
	// docs to the five-state shape before the routes mount. It is idempotent, so
	// every later boot moves nothing.
	log.Println("Initializing hobby")
	hobbyMordor := stores.NewMordor(mongo_db.DB.Collection(hobbyTable), context.Background())
	hobbyRepo := out_adapter.NewHobbyMongoAdapter(hobbyMordor)

	if migrated, err := hobbyRepo.Migrate(); nil != err {
		// the API stays up if the migration fails; unmigrated docs read with an
		// empty state until a later boot lands them
		log.Printf("hobby ships-log migration failed: %v\n", err)
	} else {
		log.Printf("hobby ships-log migration: %d doc(s) migrated\n", migrated)
	}

	hobbyService := service.NewHobbyCRUDService(hobbyRepo, noteRepo, activityService)
	in_adapter.NewHobbyMuxAdapter(hobbyService, webAuth, hobbyRouter)

	// site copy (signal flags), singleton
	log.Println("Initializing site copy")
	siteCopyMordor := stores.NewMordor(mongo_db.DB.Collection(siteCopyTable), context.Background())
	siteCopyService := service.NewSiteCopyService(out_adapter.NewSiteCopyMongoAdapter(siteCopyMordor), activityService)
	in_adapter.NewSiteCopyMuxAdapter(siteCopyService, webAuth, copyRouter)

	// the current watch (the /now record), singleton
	log.Println("Initializing watch")
	watchMordor := stores.NewMordor(mongo_db.DB.Collection(watchTable), context.Background())
	watchService := service.NewWatchService(out_adapter.NewWatchMongoAdapter(watchMordor), activityService)
	in_adapter.NewWatchMuxAdapter(watchService, webAuth, watchRouter)

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

	// the carving shop: raw-svg carvings bolted onto site spots; the seed
	// plants the shipped builtin carvings, one per spot, inserting whichever
	// are missing so every spot always has its builtin to bolt back to
	log.Println("Initializing carving")
	carvingMordor := stores.NewMordor(mongo_db.DB.Collection(carvingTable), context.Background())
	carvingService := service.NewCarvingService(out_adapter.NewCarvingMongoAdapter(carvingMordor), activityService)

	if err := carvingService.Seed(); nil != err {
		// the API stays up without the seed (the shop endpoints still work);
		// the next boot retries
		log.Printf("carving seed failed: %v\n", err)
	}

	in_adapter.NewCarvingMuxAdapter(carvingService, webAuth, carvingRouter)

	// doodles (marginalia sketches for the Keeper's Journal): structured
	// shapes only, no publish/seed/pose lifecycle
	log.Println("Initializing doodle")
	doodleMordor := stores.NewMordor(mongo_db.DB.Collection(doodleTable), context.Background())
	doodleService := service.NewDoodleService(out_adapter.NewDoodleMongoAdapter(doodleMordor), activityService)
	in_adapter.NewDoodleMuxAdapter(doodleService, webAuth, doodleRouter)

	// the harbor's tally (sightings): anonymous first-party analytics. The shore
	// pings each page view and light/note open; the watch room reads only
	// aggregates back. A TTL keeps the ledger bounded, no cookies, no consent.
	log.Println("Initializing sightings")

	// the salt lives in the config file with the other secrets; the env var
	// stays as an override for runs without one
	sightingSalt := viper.GetString("sighting.salt")

	if "" == sightingSalt {
		sightingSalt = os.Getenv("SIGHTING_SALT")
	}

	if "" == sightingSalt {
		// no salt configured: hash visitors with a per-boot random salt. It
		// resets on restart, which only splits uniques across a restart; the
		// hashes stay anonymous either way.
		sightingSalt = randomSightingSalt()
		log.Println("sighting.salt is unset; using a per-boot salt that resets on restart")
	}

	sightingMordor := stores.NewMordor(mongo_db.DB.Collection(sightingTable), context.Background())
	sightingRepo := out_adapter.NewSightingMongoAdapter(sightingMordor)

	if err := sightingRepo.EnsureIndexes(); nil != err {
		// the endpoints work without the indexes; the TTL and the window read
		// just run unindexed until a boot lands them
		log.Printf("sighting index setup failed: %v\n", err)
	}

	sightingService := service.NewSightingService(sightingRepo, sightingSalt)
	in_adapter.NewSightingMuxAdapter(sightingService, webAuth, sightingRouter)

	// users: kept (auth depends on it)
	log.Println("Initializing user")
	userMordor := stores.NewMordor(mongo_db.DB.Collection(userTable), context.Background())
	userMongoAdapter := out_adapter.NewUserMongoAdapter(userMordor)
	userService := service.NewUserCRUDService(userMongoAdapter)
	in_adapter.NewUserMuxAdapter(userService, mediaService, webAuth, userRouter)

	// auth: kept; sessions are issued through the same shared WebAuth store. The
	// login lockout ledger locks a client IP after its sixth bad hail; the only
	// reset is deleting its doc. A unique index on ip keeps one doc per client.
	log.Println("Initializing auth")
	loginLockMordor := stores.NewMordor(mongo_db.DB.Collection(loginLockTable), context.Background())
	loginLockAdapter := out_adapter.NewLoginLockMongoAdapter(loginLockMordor)

	if err := loginLockAdapter.EnsureIndexes(); nil != err {
		// the login still works without the index; the upsert just risks racing
		// two docs for one IP until a later boot lands it
		log.Printf("login lock index setup failed: %v\n", err)
	}

	userLoginService := service.NewUserLoginService(userMongoAdapter, loginLockAdapter, loginLockAdapter)
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
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Content-Range", "range", "X-Argsea-Console"})
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

// randomSightingSalt mints a boot-time salt when SIGHTING_SALT is unset, so the
// visitor hash is never keyed on an empty or predictable salt.
func randomSightingSalt() string {
	buf := make([]byte, 32)

	if _, err := rand.Read(buf); nil != err {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}

	return hex.EncodeToString(buf)
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
