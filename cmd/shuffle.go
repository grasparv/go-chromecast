// Copyright © 2018 Jonathan Pentecost <pentecostjonathan@gmail.com>
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
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/grasparv/go-chromecast/ui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// shuffleCmd represents the playlist command
var shuffleCmd = &cobra.Command{
	Use:   "shuffle <directory>",
	Short: "Load and play media on the chromecast",
	Long: `Load and play media files on the chromecast, this will
start a streaming server locally and serve the media file to the
chromecast.

If the media file is an unplayable media type by the chromecast, this
will attempt to transcode the media file to mp4 using ffmpeg. This requires
that ffmpeg is installed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("requires exactly one argument, should be the folder to play media from")
		}
		if fileInfo, err := os.Stat(args[0]); err != nil {
			fmt.Printf("unable to find %q: %v\n", args[0], err)
			return nil
		} else if !fileInfo.Mode().IsDir() {
			fmt.Printf("%q is not a directory\n", args[0])
			return nil
		}
		app, err := castApplication(cmd, args)
		if err != nil {
			fmt.Printf("unable to get cast application: %v\n", err)
			return nil
		}

		contentType, _ := cmd.Flags().GetString("content-type")
		transcode, _ := cmd.Flags().GetBool("transcode")
		forcePlay, _ := cmd.Flags().GetBool("force-play")
		directory := args[0]
		files, err := ioutil.ReadDir(directory)
		if err != nil {
			fmt.Printf("unable to list files from %q: %v", directory, err)
			return nil
		}
		filesToPlay := make([]string, 0, len(files))
		for _, f := range files {
			if !forcePlay && !app.PlayableMediaType(f.Name()) {
				continue
			}

			filesToPlay = append(filesToPlay, f.Name())
		}

		rand.Seed(time.Now().UnixNano())

		rand.Shuffle(len(filesToPlay), func(i, j int) {
			filesToPlay[i], filesToPlay[j] = filesToPlay[j], filesToPlay[i]
		})

		fmt.Println("Attemping to shuffle:")
		for _, f := range filesToPlay {
			fmt.Printf("- %s\n", f)
		}

		filenames := make([]string, len(filesToPlay))
		for i, f := range filesToPlay {
			filename := filepath.Join(directory, f)
			filenames[i] = filename
		}

		// Optionally run a UI when playing this media:
		runWithUI, _ := cmd.Flags().GetBool("with-ui")
		if runWithUI {
			go func() {
				if err := app.QueueLoad(filenames, contentType, transcode); err != nil {
					logrus.WithError(err).Fatal("unable to play playlist on cast application")
				}
			}()

			ccui, err := ui.NewUserInterface(app)
			if err != nil {
				logrus.WithError(err).Fatal("unable to prepare a new user-interface")
			}
			return ccui.Run()
		}

		if err := app.QueueLoad(filenames, contentType, transcode); err != nil {
			fmt.Printf("unable to play playlist on cast application: %v\n", err)
			return nil
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shuffleCmd)
	shuffleCmd.Flags().Bool("transcode", true, "transcode the media to mp4 if media type is unrecognised")
	shuffleCmd.Flags().Bool("force-play", false, "attempt to play a media type even if it is unrecognised")
	shuffleCmd.Flags().StringP("content-type", "c", "", "content-type to serve the media file as")
}
