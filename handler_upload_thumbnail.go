package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
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

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	// "thumbnail" should match the HTML form input name
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	// `file` is an `io.Reader` that we can read from to get the image data
	mediaType := header.Header.Get("Content-Type")
	/*
		fileData, err := io.ReadAll(file)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Unable to read form file", err)
			return
		}
	*/
	databaseVideo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get database video metadata", err)
		return
	}
	if userID != databaseVideo.UserID {
		respondWithError(w, http.StatusUnauthorized, "User not owner of video", err)
		return
	}
	/*
		thumbnailStruct := thumbnail{
			data:      fileData,
			mediaType: mediaType,
		}
	*/
	//videoThumbnails[videoID] = thumbnailStruct
	//thumbnailURL := fmt.Sprintf("http://`localhost:8091/api/thumbnails/%s", videoIDString)
	//encodedData := base64.StdEncoding.EncodeToString(fileData)
	mimeMediaType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get mime media type for file", err)
		return
	}
	if mimeMediaType != "image/jpeg" && mimeMediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Wrong image type", nil)
		return
	}
	key := make([]byte, 32)
	rand.Read(key)
	rawStringURL := base64.RawURLEncoding.EncodeToString(key)
	extension := strings.Split(mediaType, "/")[1]
	FileLocationEnd := fmt.Sprintf("/assets/%s.%s", rawStringURL, extension)
	FilePath := filepath.Join(cfg.assetsRoot, fmt.Sprintf("%s.%s", rawStringURL, extension))
	createdFile, err := os.Create(FilePath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get create file", err)
		return
	}
	io.Copy(createdFile, file)
	//thumbnailURL := fmt.Sprintf("data:%s;base64,%s", mediaType, encodedData)
	thumbnailURL := fmt.Sprintf("http://localhost:8091%s", FileLocationEnd)
	databaseVideo.ThumbnailURL = &thumbnailURL
	err = cfg.db.UpdateVideo(databaseVideo)
	respondWithJSON(w, http.StatusOK, databaseVideo)
}
