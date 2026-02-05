package main

import (
	"fmt"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
	"io"
	"os"
	"mime"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"encoding/hex"
	"crypto/rand"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	//set an upload limit 
	const maxUploadSize =  1 << 30 //
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	//validating the request 
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil{
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

	videoMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Couldnt find video", err)
		return
	}

	if videoMetadata.UserID != userID{
		respondWithError(w, http.StatusUnauthorized, "This video is not yours", nil)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Unable to parse form data", err)
		return
	}
	defer file.Close()

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type", err)
		return
	}
	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Invalid file type", nil)
		return
	}

	tempf, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Unable to create temp file", err)
		return
	}
	
	defer os.Remove(tempf.Name())
	defer tempf.Close()

	if _, err := io.Copy(tempf, file); err != nil{
		respondWithError(w, http.StatusInternalServerError, "Error saving file", err)
		return
	}

	tempf.Seek(0, io.SeekStart)
	
	

	// after Seek(0, io.SeekStart)
	
	ratio, err := getVideoAspectRatio(tempf.Name())
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "Could not get aspect ratio", err)
		return
	}
	
	//change it into a fast file type
	_, err = processVideoForFastStart(tempf.Name())
	if err != nil{
		respondWithError(w , http.StatusInternalServerError, "Could not covert the file", err)
		return 
	}
	
	pFileName, err := processVideoForFastStart(tempf.Name())
	processedFile, err := os.Open(pFileName)
	
	defer os.Remove(processedFile.Name())
	defer processedFile.Close()

	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "Could not open fast file", err)
		return
	}	

	var prefix string
	if ratio == "16:9"{
		prefix = "landscape/"
	}else if ratio == "9:16"{
		prefix = "portrait/"
	}else{
		prefix = "other/"
	}



	keyBytes := make([]byte, 32)
	// what function fills keyBytes with random data?
	// (hint: from the crypto/rand package)
	rand.Read(keyBytes)

	randomHex := prefix + hex.EncodeToString(keyBytes)

	// now use both arguments, as your getAssetPath expects
	key := getAssetPath(randomHex, mediaType)
	
	//put object
	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: aws.String(cfg.s3Bucket),
		Key:    aws.String(key),
		Body:   processedFile,
		ContentType: aws.String(mediaType),
	})

	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "Error Uploading Image to AWS", err)
		return
	}
	
	url := fmt.Sprintf("https://%v.s3.%v.amazonaws.com/%v", cfg.s3Bucket, cfg.s3Region, key)
	videoMetadata.VideoURL = &url
	err = cfg.db.UpdateVideo(videoMetadata)
	if err != nil {
 	   respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
    	   return
	}

	respondWithJSON(w, http.StatusOK, videoMetadata)
}
