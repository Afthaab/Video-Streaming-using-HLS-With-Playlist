package hls

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateHLS(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Error uploading file: %s", err.Error())
		return
	}

	if !strings.HasPrefix(file.Header.Get("Content-Type"), "video/") {
		c.String(http.StatusBadRequest, "Error: only video files are allowed")
		return
	}

	err = c.SaveUploadedFile(file, "./videos/"+file.Filename)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error saving file: %s", err.Error())
		return
	}
	inputFile, err := filepath.Abs("./videos/" + file.Filename)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting file path: %s", err.Error())
		return
	}

	c.String(http.StatusOK, "Folder path: %s", inputFile)

	outputDir := "/home/afthab/Desktop/videoPlaylistStreaming/outputFile"
	segmentTime := "10" // Segment duration in seconds

	// Resolution Variants
	resolutions := []struct {
		Width  int
		Height int
	}{
		{Width: 854, Height: 480},   // 480p
		{Width: 1280, Height: 720},  // 720p
		{Width: 1920, Height: 1080}, // 1080p
	}

	// Creating HLS for each resolution variant
	for _, res := range resolutions {
		outputPath := outputDir + "/" + resToString(res)
		err := createHLSVariant(inputFile, outputPath, segmentTime, res)
		if err != nil {
			log.Println("Error creating HLS variant:", err)
			continue
		}
	}

	// Create master playlist
	err = createMasterPlaylist(outputDir, resolutions)
	if err != nil {
		log.Println("Error creating master playlist:", err)
	}

	c.JSON(200, gin.H{
		"Message": "Okay",
	})
}

// Function to create HLS variant for a given resolution
func createHLSVariant(inputFile string, outputDir string, segmentTime string, resolution struct{ Width, Height int }) error {
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-vf", "scale="+strconv.Itoa(resolution.Width)+":"+strconv.Itoa(resolution.Height),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-hls_time", segmentTime,
		"-hls_list_size", "0",
		"-hls_segment_filename", outputDir+"/%03d.ts",
		outputDir+"/playlist.m3u8",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Function to create master playlist
func createMasterPlaylist(outputDir string, resolutions []struct{ Width, Height int }) error {
	playlistPath := outputDir + "/master.m3u8"
	playlistFile, err := os.Create(playlistPath)
	if err != nil {
		return err
	}
	defer playlistFile.Close()

	playlistFile.WriteString("#EXTM3U\n")
	for i, res := range resolutions {
		variantURL := resToString(res) + "/playlist.m3u8"
		playlistFile.WriteString("#EXT-X-STREAM-INF:BANDWIDTH=500000,RESOLUTION=" + resToString(res) + "\n")
		playlistFile.WriteString(variantURL + "\n")
		if i < len(resolutions)-1 {
			playlistFile.WriteString("\n")
		}
	}

	return nil
}

// Helper function to convert resolution to string format
func resToString(res struct{ Width, Height int }) string {
	return strconv.Itoa(res.Width) + "x" + strconv.Itoa(res.Height)
}
