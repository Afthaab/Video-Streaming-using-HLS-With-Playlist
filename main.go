package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hls/playlist/hls"
)

func main() {
	r := gin.Default()
	r.POST("/select/file", hls.CreateHLS)
	r.Static("/stream", "/home/afthab/Desktop/videoPlaylistStreaming/outputFile/")
	fmt.Println("Port running at :9998/stream/playlist.m3u8")
	r.Run(":9997")
}

// https://hlsjs-dev.video-dev.org/demo/
// http://localhost:9997/stream/master.m3u8
