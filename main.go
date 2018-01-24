package main

import (
	"errors"
	"fmt"
	"github.com/seichewarning/restapi/middlewares/context"
	"github.com/seichewarning/restapi/middlewares/mongodb"
	"github.com/seichewarning/restapi/middlewares/renderer"
	"github.com/seichewarning/restapi/server"
	"github.com/seichewarning/restapi/sessions"
	"github.com/seichewarning/restapi/users"
	"io/ioutil"
	"time"
)

func main() {

	// try to load signing keys for token authority
	// NOTE: DO NOT USE THESE KEYS FOR PRODUCTION! FOR DEMO ONLY
	privateSigningKey, err := ioutil.ReadFile("keys/demo.rsa")
	if err != nil {
		panic(errors.New(fmt.Sprintf("Error loading private signing key: %v", err.Error())))
	}
	publicSigningKey, err := ioutil.ReadFile("keys/demo.rsa.pub")
	if err != nil {
		panic(errors.New(fmt.Sprintf("Error loading public signing key: %v", err.Error())))
	}

	// create current project context
	ctx := context.New()
	fmt.Println("here!!!")

	// set up DB session
	db := mongodb.New(&mongodb.Options{
		ServerName:   "localhost",
		DatabaseName: "test-go-app",
	})
	fmt.Println("here!!!")
	_ = db.NewSession()
	fmt.Println("here!!!")

	// set up Renderer (unrolled_render)
	renderer := renderer.New(&renderer.Options{
		IndentJSON: true,
	}, renderer.JSON)

	// set up users resource
	usersResource := users.NewResource(ctx, &users.Options{
		Database: db,
		Renderer: renderer,
	})

	// set up sessions resource
	sessionsResource := sessions.NewResource(ctx, &sessions.Options{
		PrivateSigningKey:     privateSigningKey,
		PublicSigningKey:      publicSigningKey,
		Database:              db,
		Renderer:              renderer,
		UserRepositoryFactory: usersResource.UserRepositoryFactory,
	})

	// init server
	s := server.NewServer(&server.Config{
		Context: ctx,
	})

	// set up router
	ac := server.NewAccessController(ctx, renderer)
	router := server.NewRouter(s.Context, ac)

	// add REST resources to router
	router.AddResources(sessionsResource, usersResource)

	// add middlewares
	s.UseMiddleware(sessionsResource.NewAuthenticator())

	// setup router
	s.UseRouter(router)

	// bam!
	s.Run(":3001", server.Options{
		Timeout: 10 * time.Second,
	})
}
