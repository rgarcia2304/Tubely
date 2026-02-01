package main

import (
	"fmt"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
	"io"
	"encoding/base64"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	// validating the request
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
	const maxMemory =  10 << 20
	r.ParseMultipartForm(maxMemory)

	// "thumbnail" should match the HTML form input name
	file, header, err := r.FormFile("thumbnail")
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}

	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Unable to read image data", err)
		return
	}
	
	// get the videos metadata
	vidData, err := cfg.db.GetVideo(videoID)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Unable to fetch the video with provided ID", err)
		return 
	}

	if vidData.CreateVideoParams.UserID != userID{
		respondWithJSON(w, http.StatusUnauthorized, struct{}{})
		return 
	}
	
	//convert the image data
	encodedImageData := base64.StdEncoding.EncodeToString(imageData)
	dataUrl := fmt.Sprintf("data:%v;base64,%v", header.Header.Get("Content-Type"), encodedImageData)

	//create the new thumbnail struct 
	newThumbnail := thumbnail{data: imageData, mediaType: header.Header.Get("Content-Type")}
	//videoThumbnails[vidData.ID] = newThumbnail

	vidData.ThumbnailURL = &dataUrl
	err = cfg.db.UpdateVideo(vidData)
	
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Unable to update the video", err)
		return
	}
	
	respondWithJSON(w, http.StatusOK, newThumbnail)
}
