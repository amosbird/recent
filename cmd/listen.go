// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen to a bunch of directories for file creation events.",
	Long: `All files created under those directories will be recorded in ~/.recent.db sqlite file.
Each record has a time field and a path field. How to get the best usage of it is up to the user.`,
	Run: func(cmd *cobra.Command, args []string) {
		home := os.Getenv("HOME")
		db, err := sql.Open("sqlite3", home+"/.recent.db")
		if err != nil {
			log.Fatal("Failed to open database")
		}
		defer db.Close()

		_, err = db.Exec("create table if not exists files (time integer primary key, path string)")
		if err != nil {
			log.Fatal("Failed to create table: ", err)
		}

		stmt, err := db.Prepare("insert into files values(?, ?)")

		if err != nil {
			log.Fatal("Failed to prepare statement: ", err)
		}

		defer stmt.Close()
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal("Failed to create fsnotify watcher: ", err)
		}

		done := make(chan bool)

		// Process events
		go func() {
			for {
				select {
				case event := <-watcher.Events:
					if event.Op&fsnotify.Create == fsnotify.Create {
						log.Println("event: ", event.Name)
						time := time.Now().UnixNano()
						res, err := stmt.Exec(time, event.Name)
						if err != nil {
							log.Fatal("Failed to insert record: ", err)
						}
						affected, _ := res.RowsAffected()
						if affected != 1 {
							log.Fatalf("Expected %d for affected rows, but %d: ", 1, affected)
						}
					}
				case err := <-watcher.Errors:
					log.Println("error: ", err)
				}
			}
		}()

		for _, dir := range args {
			err = watcher.Add(dir)
			if err != nil {
				log.Fatal("Failed to add watching dir: ", err)
			}
		}

		// Hang so program doesn't exit
		<-done

		/* ... do stuff ... */
		watcher.Close()
	},
}

func init() {
	RootCmd.AddCommand(listenCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listenCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listenCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
