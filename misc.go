package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

type Data struct {
	Streams []struct {
		Width              int    `json:"width"`
		Height             int    `json:"height"`
		DisplayAspectRatio string `json:"display_aspect_ratio"`
	} `json:"streams"`
}

func getVideoAspectRatio(filePath string) (string, error) {
	var output bytes.Buffer
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	cmd.Stdout = &output
	if err := cmd.Run(); err != nil {
		return "", err
	}
	
	data := &Data{}
	json.Unmarshal(output.Bytes(), data)
	if data.Streams[0].DisplayAspectRatio != "0:1" && data.Streams[0].DisplayAspectRatio != "" {
		return data.Streams[0].DisplayAspectRatio, nil
	}
	aspectRatio := float32(data.Streams[0].Width) / float32(data.Streams[0].Height)
	if aspectRatio > 15.0/9.0 && aspectRatio < 17.0/9.0	{
		return "16:9", nil
	} else if aspectRatio > 8.0/16.0 && aspectRatio < 10.0/16.0 {
		return "9:16", nil
	} else {
		return "other", nil
	}
}

func processVideoForFastStart(filePath string) (string, error) {
	outputPath := filePath + ".processing"
	cmd := exec.Command("ffmpeg",  "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)
	if err := cmd.Run(); err != nil {
		return filePath, err
	}
	return outputPath, nil
}
