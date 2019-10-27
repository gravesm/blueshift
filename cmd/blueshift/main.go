package main

import (
	"fmt"
	"github.com/gravesm/blueshift/pkg/server"
	"github.com/gravesm/blueshift/pkg/services"
	"github.com/gravesm/blueshift/pkg/store"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	cmd := cli.NewApp()
	cmd.Commands = []cli.Command{
		{
			Name: "init",
			Action: func(c *cli.Context) error {
				db, err := gorm.Open("sqlite3", "test.db")
				if err != nil {
					return err
				}
				store.Initialize(db)
				return nil
			},
		},
		{
			Name: "server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "address",
					Value: ":6000",
					Usage: "Address to listen on",
				},
			},
			Action: func(c *cli.Context) error {
				db, err := gorm.Open("sqlite3", "test.db")
				if err != nil {
					log.Fatal(err)
				}
				collection := store.NewDbCollection(db)
				server := server.NewServer(collection, services.FileStreamHandler{"files"}, "templates")
				srv := &http.Server{
					Handler:      server,
					Addr:         c.String("address"),
					WriteTimeout: 15 * time.Second,
					ReadTimeout:  15 * time.Second,
				}
				log.Fatal(srv.ListenAndServe())
				return nil
			},
		},
	}

	err := cmd.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
