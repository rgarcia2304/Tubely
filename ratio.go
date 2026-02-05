package main
import(
	"os/exec"
	"bytes"
	"encoding/json"
)
func getVideoAspectRatio(filePath string) (string, error){
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var b bytes.Buffer
	cmd.Stdout = &b
	err := cmd.Run()

	if err != nil{
		return "hi", err
	}
	
	type Stream struct{
		Width int `json:"width"`
		Height int `json:"height"`
	}

	type FFProbeOutput struct {
	Streams []Stream `json:"streams"`
	}

	
	var probeOutput FFProbeOutput

	dec := json.NewDecoder(&b)

	err = dec.Decode(&probeOutput)

	if err != nil{
		return "hello", err	
	}

	//do math to verify aspect ratios
	ratio := float64(float64(probeOutput.Streams[0].Height) / float64(probeOutput.Streams[0].Width))
	if ratio > 0.5  && ratio < 0.6{
		return "16:9", nil
	}else if ratio > 1.7 && ratio < 1.8 {
		return "9:16", nil
	}else{
		return "other", nil 
	}

}
