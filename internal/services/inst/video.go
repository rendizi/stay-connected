package stories

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func downloadVideo(url string, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractFrames(videoPath string, outputDir string) error {
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vf", "fps=1", fmt.Sprintf("%s/frame_%%03d.jpg", outputDir))
	return cmd.Run()
}
