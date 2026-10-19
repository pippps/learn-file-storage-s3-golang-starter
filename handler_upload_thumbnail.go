package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"mime"
	"crypto/rand"
	"encoding/base64"

	"github.com/pippps/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse form", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	videoMD, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find video", err)
		return
	}
	if videoMD.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to update this video", nil)
		return
	}
	
	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil || (mediaType != "image/jpeg" && mediaType != "image/png") {
		respondWithError(w, http.StatusUnsupportedMediaType, "Unsupported media type", err)
		return 
	}
	
	tnExtension, found := strings.CutPrefix(mediaType, "image/")
	if !found {
		respondWithError(w, http.StatusInternalServerError, "media type parsing problem", fmt.Errorf("problem cutting type mime from subtype"))
	}
	ran := make([]byte, 32)
	_, err = rand.Read(ran)
	if err != nil {
		respondWithError(w,  http.StatusInternalServerError, "error uploading image", err)
	}
	ranB64 := base64.RawURLEncoding.EncodeToString(ran)

	fileName := fmt.Sprintf("assets/%v.%s",ranB64,tnExtension)
	
	fileOS, err := os.Create(fileName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't upload image", err)
		return 
	}
	defer fileOS.Close()

	_, err = io.Copy(fileOS,file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't upload image", err)
		return
	}
	
	tnURL := fmt.Sprintf("http://localhost:8091/%s",fileName)
	videoMD.ThumbnailURL = &tnURL

	err = cfg.db.UpdateVideo(videoMD)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoMD)
}
