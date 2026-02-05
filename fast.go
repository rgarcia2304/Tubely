package main 
import(
	"os/exec"
	"log"
)
func processVideoForFastStart(filePath string) (string, error){
	fastPath := filePath + ".processing"
	log.Println(fastPath)
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4",fastPath)

	err := cmd.Run()
	log.Println("An error occured", err)

	if err != nil{
		return "", err
	}
	
	return fastPath, nil
}

